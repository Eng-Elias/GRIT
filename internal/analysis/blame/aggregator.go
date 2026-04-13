package blame

import (
	"math"
	"sort"
	"time"

	"github.com/grit-app/grit/internal/models"
)

// authorAccumulator collects per-author data during aggregation.
type authorAccumulator struct {
	email           string
	name            string
	latestDate      time.Time
	earliestDate    time.Time
	totalLines      int
	files           map[string]bool
	languageLines   map[string]int
}

// AggregateResult holds the output of the aggregation step.
type AggregateResult struct {
	Authors    []models.Author
	TotalLines int
}

// Aggregate takes per-file blame results and produces deduplicated per-author statistics.
// Emails are deduplicated case-insensitively. The display name is taken from the
// most recent commit date for each email.
func Aggregate(fileResults map[string]*FileBlameResult) *AggregateResult {
	accumulators := make(map[string]*authorAccumulator)
	totalLines := 0

	for _, fr := range fileResults {
		for _, line := range fr.Lines {
			totalLines++
			email := line.AuthorEmail // already lowercased by normalizeEmail
			acc, ok := accumulators[email]
			if !ok {
				acc = &authorAccumulator{
					email:         email,
					name:          line.AuthorName,
					latestDate:    line.Date,
					earliestDate:  line.Date,
					files:         make(map[string]bool),
					languageLines: make(map[string]int),
				}
				accumulators[email] = acc
			}
			acc.totalLines++
			acc.files[fr.Path] = true
			acc.languageLines[fr.Language]++

			// Use the display name from the most recent commit.
			if line.Date.After(acc.latestDate) {
				acc.latestDate = line.Date
				acc.name = line.AuthorName
			}
			if line.Date.Before(acc.earliestDate) {
				acc.earliestDate = line.Date
			}
		}
	}

	if totalLines == 0 {
		return &AggregateResult{Authors: []models.Author{}, TotalLines: 0}
	}

	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	authors := make([]models.Author, 0, len(accumulators))
	for _, acc := range accumulators {
		authors = append(authors, models.Author{
			Email:            acc.email,
			Name:             acc.name,
			TotalLinesOwned:  acc.totalLines,
			OwnershipPercent: roundPercent(float64(acc.totalLines) / float64(totalLines) * 100),
			FilesTouched:     len(acc.files),
			FirstCommitDate:  acc.earliestDate,
			LastCommitDate:   acc.latestDate,
			PrimaryLanguages: topLanguages(acc.languageLines, 3),
			IsActive:         acc.latestDate.After(sixMonthsAgo),
		})
	}

	// Sort by total_lines_owned descending.
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].TotalLinesOwned > authors[j].TotalLinesOwned
	})

	return &AggregateResult{Authors: authors, TotalLines: totalLines}
}

// ComputeBusFactor finds the minimum number of authors whose combined ownership
// exceeds 80% of total lines. Returns the count and the key people slice.
func ComputeBusFactor(authors []models.Author) (int, []models.Author) {
	if len(authors) == 0 {
		return 0, nil
	}

	var cumulative float64
	var keyPeople []models.Author
	for _, a := range authors {
		cumulative += a.OwnershipPercent
		keyPeople = append(keyPeople, a)
		if cumulative >= 80.0 {
			break
		}
	}

	return len(keyPeople), keyPeople
}

// ComputeFileContributors produces a per-file top-3 author breakdown.
func ComputeFileContributors(fileResults map[string]*FileBlameResult) []models.FileContributors {
	result := make([]models.FileContributors, 0, len(fileResults))

	for _, fr := range fileResults {
		if len(fr.Lines) == 0 {
			continue
		}

		// Count lines per author for this file.
		authorLines := make(map[string]int)
		authorNames := make(map[string]string)
		for _, line := range fr.Lines {
			authorLines[line.AuthorEmail]++
			// Use the most recent name for this email in this file.
			authorNames[line.AuthorEmail] = line.AuthorName
		}

		totalLines := len(fr.Lines)

		// Build sorted slice.
		type authorCount struct {
			email string
			name  string
			lines int
		}
		counts := make([]authorCount, 0, len(authorLines))
		for email, lines := range authorLines {
			counts = append(counts, authorCount{email: email, name: authorNames[email], lines: lines})
		}
		sort.Slice(counts, func(i, j int) bool {
			return counts[i].lines > counts[j].lines
		})

		// Take top 3.
		top := 3
		if len(counts) < top {
			top = len(counts)
		}

		topAuthors := make([]models.FileAuthor, top)
		for i := 0; i < top; i++ {
			topAuthors[i] = models.FileAuthor{
				Name:             counts[i].name,
				Email:            counts[i].email,
				LinesOwned:       counts[i].lines,
				OwnershipPercent: roundPercent(float64(counts[i].lines) / float64(totalLines) * 100),
			}
		}

		result = append(result, models.FileContributors{
			Path:       fr.Path,
			TotalLines: totalLines,
			TopAuthors: topAuthors,
		})
	}

	// Sort by path for deterministic output.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result
}

// topLanguages returns the top N languages by line count.
func topLanguages(langLines map[string]int, n int) []string {
	type langCount struct {
		lang  string
		count int
	}

	counts := make([]langCount, 0, len(langLines))
	for lang, count := range langLines {
		counts = append(counts, langCount{lang, count})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	top := n
	if len(counts) < top {
		top = len(counts)
	}

	result := make([]string, top)
	for i := 0; i < top; i++ {
		result[i] = counts[i].lang
	}
	return result
}

// roundPercent rounds a percentage to 1 decimal place.
func roundPercent(v float64) float64 {
	return math.Round(v*10) / 10
}
