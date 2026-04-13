package churn

import (
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/grit-app/grit/internal/models"
)

const (
	MaxCommits    = 5000
	MaxAge        = 2 * 365 * 24 * time.Hour // 2 years
	StaleMonths   = 6
)

// ChurnResult holds the output of a commit log walk.
type ChurnResult struct {
	Files             []models.FileChurn
	TotalCommits      int
	CommitWindowStart time.Time
	CommitWindowEnd   time.Time
}

// WalkCommitLog walks the default branch commit log and produces per-file churn counts.
// It stops after MaxCommits or when commits are older than MaxAge, whichever comes first.
func WalkCommitLog(repo *git.Repository) (*ChurnResult, error) {
	head, err := repo.Head()
	if err != nil {
		return &ChurnResult{}, nil
	}

	logIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}
	defer logIter.Close()

	cutoff := time.Now().Add(-MaxAge)
	churnMap := make(map[string]int)
	lastModified := make(map[string]time.Time)
	var commitCount int
	var windowStart, windowEnd time.Time

	err = logIter.ForEach(func(c *object.Commit) error {
		if commitCount >= MaxCommits {
			return errStopIter
		}
		if c.Author.When.Before(cutoff) {
			return errStopIter
		}

		commitTime := c.Author.When
		if commitCount == 0 {
			windowEnd = commitTime
		}
		windowStart = commitTime
		commitCount++

		stats, err := c.Stats()
		if err != nil {
			// Skip commits where stats can't be computed (e.g., initial commit in some cases).
			return nil
		}

		for _, s := range stats {
			churnMap[s.Name]++
			if existing, ok := lastModified[s.Name]; !ok || commitTime.After(existing) {
				lastModified[s.Name] = commitTime
			}
		}

		return nil
	})
	if err != nil && err != errStopIter {
		return nil, err
	}

	files := make([]models.FileChurn, 0, len(churnMap))
	for path, count := range churnMap {
		lm := lastModified[path]
		files = append(files, models.FileChurn{
			Path:         path,
			Churn:        count,
			LastModified: lm,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Churn > files[j].Churn
	})

	return &ChurnResult{
		Files:             files,
		TotalCommits:      commitCount,
		CommitWindowStart: windowStart,
		CommitWindowEnd:   windowEnd,
	}, nil
}

// errStopIter is a sentinel error to stop the commit iterator early.
var errStopIter = &stopIterError{}

type stopIterError struct{}

func (e *stopIterError) Error() string { return "stop iteration" }
