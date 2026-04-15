package temporal

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/grit-app/grit/internal/models"
)

const velocityWeeks = 52

type weekKey struct {
	year int
	week int
}

type weekAccum struct {
	additions int
	deletions int
	commits   int
}

type authorAccum struct {
	name      string
	additions int
	deletions int
	weeks     map[weekKey]weekAccum
}

// isoWeekLabel returns a label like "2026-W16".
func isoWeekLabel(year, week int) string {
	return fmt.Sprintf("%d-W%02d", year, week)
}

// weekStartFromISO computes the Monday of the given ISO year/week.
func weekStartFromISO(year, week int) time.Time {
	// Jan 4 is always in week 1 of the ISO year.
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	// Weekday of Jan 4 (Monday=1 ... Sunday=7 in ISO)
	isoDay := int(jan4.Weekday())
	if isoDay == 0 {
		isoDay = 7
	}
	// Monday of week 1
	mondayW1 := jan4.AddDate(0, 0, -(isoDay - 1))
	return mondayW1.AddDate(0, 0, (week-1)*7)
}

// ComputeVelocity walks the commit log for the past 52 weeks and aggregates
// weekly additions/deletions, per-author activity, and rolling commit cadence.
func ComputeVelocity(repo *git.Repository) ([]models.WeeklyActivity, []models.AuthorWeeklyActivity, []models.CommitCadence, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, nil, nil, err
	}

	logIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	defer logIter.Close()

	cutoff := time.Now().UTC().AddDate(0, 0, -velocityWeeks*7)

	weekMap := make(map[weekKey]*weekAccum)
	authorMap := make(map[string]*authorAccum)

	err = logIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When.UTC()
		if commitTime.Before(cutoff) {
			return errStopIter
		}

		year, week := commitTime.ISOWeek()
		wk := weekKey{year, week}

		// Weekly aggregation
		wa, ok := weekMap[wk]
		if !ok {
			wa = &weekAccum{}
			weekMap[wk] = wa
		}
		wa.commits++

		stats, err := c.Stats()
		if err != nil {
			return nil // skip commits where stats can't be computed
		}

		for _, s := range stats {
			wa.additions += s.Addition
			wa.deletions += s.Deletion
		}

		// Per-author aggregation
		email := strings.ToLower(c.Author.Email)
		aa, ok := authorMap[email]
		if !ok {
			aa = &authorAccum{
				name:  c.Author.Name,
				weeks: make(map[weekKey]weekAccum),
			}
			authorMap[email] = aa
		}
		// Keep most recent name
		aa.name = c.Author.Name

		for _, s := range stats {
			aa.additions += s.Addition
			aa.deletions += s.Deletion
		}

		aw := aa.weeks[wk]
		aw.commits++
		for _, s := range stats {
			aw.additions += s.Addition
			aw.deletions += s.Deletion
		}
		aa.weeks[wk] = aw

		return nil
	})
	if err != nil && err != errStopIter {
		return nil, nil, nil, err
	}

	// Convert weekMap to sorted slice
	weekly := buildWeeklyActivity(weekMap)
	authors := rankAuthors(authorMap, 10)
	cadence := computeCommitCadence(weekly)

	return weekly, authors, cadence, nil
}

// buildWeeklyActivity converts the weekMap to a sorted slice of WeeklyActivity.
func buildWeeklyActivity(weekMap map[weekKey]*weekAccum) []models.WeeklyActivity {
	weeks := make([]models.WeeklyActivity, 0, len(weekMap))
	for wk, wa := range weekMap {
		weeks = append(weeks, models.WeeklyActivity{
			Week:      isoWeekLabel(wk.year, wk.week),
			WeekStart: weekStartFromISO(wk.year, wk.week),
			Additions: wa.additions,
			Deletions: wa.deletions,
			Commits:   wa.commits,
		})
	}

	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].WeekStart.Before(weeks[j].WeekStart)
	})

	return weeks
}

// rankAuthors sorts authors by total activity and returns the top N.
func rankAuthors(authorMap map[string]*authorAccum, n int) []models.AuthorWeeklyActivity {
	type ranked struct {
		email    string
		accum    *authorAccum
		activity int
	}

	entries := make([]ranked, 0, len(authorMap))
	for email, aa := range authorMap {
		entries = append(entries, ranked{email, aa, aa.additions + aa.deletions})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].activity > entries[j].activity
	})

	limit := n
	if len(entries) < limit {
		limit = len(entries)
	}

	result := make([]models.AuthorWeeklyActivity, 0, limit)
	for _, e := range entries[:limit] {
		awa := models.AuthorWeeklyActivity{
			Email:          e.email,
			Name:           e.accum.name,
			TotalAdditions: e.accum.additions,
			TotalDeletions: e.accum.deletions,
		}

		for wk, wa := range e.accum.weeks {
			awa.Weeks = append(awa.Weeks, models.WeeklyActivity{
				Week:      isoWeekLabel(wk.year, wk.week),
				WeekStart: weekStartFromISO(wk.year, wk.week),
				Additions: wa.additions,
				Deletions: wa.deletions,
				Commits:   wa.commits,
			})
		}

		sort.Slice(awa.Weeks, func(i, j int) bool {
			return awa.Weeks[i].WeekStart.Before(awa.Weeks[j].WeekStart)
		})

		result = append(result, awa)
	}

	return result
}

// computeCommitCadence computes rolling 4-week commit cadence from weekly activity.
func computeCommitCadence(weeks []models.WeeklyActivity) []models.CommitCadence {
	const windowSize = 4

	if len(weeks) < windowSize {
		return nil
	}

	var cadence []models.CommitCadence
	for i := 0; i <= len(weeks)-windowSize; i++ {
		totalCommits := 0
		for j := i; j < i+windowSize; j++ {
			totalCommits += weeks[j].Commits
		}

		cadence = append(cadence, models.CommitCadence{
			WindowStart:   weeks[i].WeekStart,
			WindowEnd:     weeks[i+windowSize-1].WeekStart.AddDate(0, 0, 6),
			CommitsPerDay: float64(totalCommits) / 28.0,
		})
	}

	return cadence
}
