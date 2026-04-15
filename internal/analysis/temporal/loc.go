package temporal

import (
	"bytes"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/grit-app/grit/internal/analysis/core"
	"github.com/grit-app/grit/internal/models"
)

// parsePeriodMonths converts a period string ("1y","2y","3y") to month count.
func parsePeriodMonths(period string) int {
	switch period {
	case "1y":
		return 12
	case "2y":
		return 24
	default:
		return 36
	}
}

// monthBoundaries returns the first day of each month going back `period` months from now,
// sorted chronologically (oldest first).
func monthBoundaries(now time.Time, period string) []time.Time {
	months := parsePeriodMonths(period)
	boundaries := make([]time.Time, 0, months)
	for i := months - 1; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		boundaries = append(boundaries, time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC))
	}
	return boundaries
}

// resolveMonthlyCommits walks the commit log and resolves the latest commit on or before
// each monthly boundary. Returns a map of boundary → commit hash.
func resolveMonthlyCommits(repo *git.Repository, boundaries []time.Time) (map[time.Time]plumbing.Hash, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	logIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}
	defer logIter.Close()

	result := make(map[time.Time]plumbing.Hash)
	earliest := boundaries[0]

	err = logIter.ForEach(func(c *object.Commit) error {
		commitTime := c.Committer.When.UTC()

		// Stop once we're past the earliest boundary
		if commitTime.Before(earliest.AddDate(0, -1, 0)) {
			return errStopIter
		}

		// For each boundary, pick the latest commit on or before it
		for _, b := range boundaries {
			if _, exists := result[b]; exists {
				continue
			}
			if !commitTime.After(b.AddDate(0, 1, 0).Add(-time.Second)) {
				// commitTime is on or before the end of boundary month
				if !commitTime.Before(earliest.AddDate(0, -1, 0)) {
					result[b] = c.Hash
				}
			}
		}

		return nil
	})
	if err != nil && err != errStopIter {
		return nil, err
	}

	return result, nil
}

// countLOCAtCommit walks the tree of the given commit and counts lines per language.
func countLOCAtCommit(repo *git.Repository, hash plumbing.Hash) (int, map[string]int, error) {
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return 0, nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return 0, nil, err
	}

	totalLOC := 0
	langMap := make(map[string]int)

	err = tree.Files().ForEach(func(f *object.File) error {
		if shouldSkipTemporal(f.Name) {
			return nil
		}

		reader, err := f.Reader()
		if err != nil {
			return nil
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}

		if len(content) > 0 && core.IsBinary(content[:min(len(content), 512)]) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(f.Name))
		filename := filepath.Base(f.Name)
		lang := core.LookupLanguage(ext, filename)

		lines := countNewlines(content)
		totalLOC += lines
		langMap[lang.Name] += lines

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	return totalLOC, langMap, nil
}

// countNewlines counts the number of lines in content (fast path for LOC).
func countNewlines(data []byte) int {
	return bytes.Count(data, []byte{'\n'}) + 1
}

// shouldSkipTemporal checks if a file should be excluded from LOC counting.
func shouldSkipTemporal(path string) bool {
	skipExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg",
		".webp", ".mp3", ".mp4", ".wav", ".avi", ".mov",
		".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".exe", ".dll", ".so", ".dylib", ".bin",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".lock"}

	ext := strings.ToLower(filepath.Ext(path))
	for _, skip := range skipExts {
		if ext == skip {
			return true
		}
	}
	return false
}

// topLanguages returns the top N languages by LOC with the rest grouped as "Other".
func topLanguages(langMap map[string]int, n int) map[string]int {
	if len(langMap) == 0 {
		return langMap
	}

	type langEntry struct {
		name  string
		count int
	}

	entries := make([]langEntry, 0, len(langMap))
	for name, count := range langMap {
		entries = append(entries, langEntry{name, count})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})

	if len(entries) <= n {
		return langMap
	}

	result := make(map[string]int, n+1)
	other := 0
	for i, e := range entries {
		if i < n {
			result[e.name] = e.count
		} else {
			other += e.count
		}
	}
	if other > 0 {
		result["Other"] = other
	}

	return result
}

// ComputeLOCOverTime produces monthly LOC snapshots for the given repository and period.
func ComputeLOCOverTime(repo *git.Repository, period string) ([]models.MonthlySnapshot, error) {
	now := time.Now().UTC()
	boundaries := monthBoundaries(now, period)

	commitMap, err := resolveMonthlyCommits(repo, boundaries)
	if err != nil {
		return nil, err
	}

	var snapshots []models.MonthlySnapshot
	for _, b := range boundaries {
		hash, ok := commitMap[b]
		if !ok {
			continue
		}

		totalLOC, langMap, err := countLOCAtCommit(repo, hash)
		if err != nil {
			continue
		}

		snapshots = append(snapshots, models.MonthlySnapshot{
			Date:       b,
			TotalLOC:   totalLOC,
			ByLanguage: topLanguages(langMap, 10),
		})
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Date.Before(snapshots[j].Date)
	})

	return snapshots, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
