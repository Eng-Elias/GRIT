package churn

import (
	"io"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/grit-app/grit/internal/models"
)

// sourceExtensions is the allowlist of file extensions considered as source code.
var sourceExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
	".java": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
	".cs": true, ".rb": true, ".rs": true, ".swift": true, ".kt": true,
	".scala": true, ".php": true, ".pl": true, ".pm": true, ".r": true,
	".sh": true, ".sql": true, ".proto": true, ".graphql": true,
}

// excludeExtensions is the denylist of non-source file extensions.
var excludeExtensions = map[string]bool{
	".md": true, ".txt": true, ".yaml": true, ".yml": true, ".json": true,
	".toml": true, ".xml": true, ".csv": true, ".svg": true, ".png": true,
	".jpg": true, ".gif": true, ".ico": true, ".woff": true, ".ttf": true,
	".lock": true, ".sum": true, ".mod": true,
}

// IsSourceExtension returns true if the file extension indicates source code.
func IsSourceExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if excludeExtensions[ext] {
		return false
	}
	return sourceExtensions[ext]
}

// MonthsInactive returns the approximate number of months between lastModified and now.
func MonthsInactive(lastModified time.Time, now time.Time) int {
	if lastModified.IsZero() {
		return 0
	}
	days := now.Sub(lastModified).Hours() / 24
	return int(math.Floor(days / 30.0))
}

// FindStaleFiles identifies files that match the 3-condition stale heuristic:
//  1. No commits in the past StaleMonths months
//  2. Source code extension
//  3. No import references (baseName substring scan)
//
// repo is used to read file contents at HEAD for the import scan.
func FindStaleFiles(files []models.FileChurn, repo *git.Repository, now time.Time) []models.StaleFile {
	staleCutoff := now.AddDate(0, -StaleMonths, 0)

	// Step 1+2: Filter candidates by recency and extension.
	var candidates []models.FileChurn
	for _, f := range files {
		if f.LastModified.Before(staleCutoff) && IsSourceExtension(f.Path) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Build set of source file contents for import scan.
	sourceContents := readSourceContents(repo)

	var stale []models.StaleFile
	for _, c := range candidates {
		baseName := strings.TrimSuffix(filepath.Base(c.Path), filepath.Ext(c.Path))
		if baseName == "" {
			continue
		}

		if isReferenced(baseName, c.Path, sourceContents) {
			continue
		}

		stale = append(stale, models.StaleFile{
			Path:           c.Path,
			LastModified:   c.LastModified,
			MonthsInactive: MonthsInactive(c.LastModified, now),
		})
	}

	return stale
}

// isReferenced checks if baseName appears as a substring in any source file
// content other than the candidate file itself.
func isReferenced(baseName, candidatePath string, sourceContents map[string]string) bool {
	for path, content := range sourceContents {
		if path == candidatePath {
			continue
		}
		if strings.Contains(content, baseName) {
			return true
		}
	}
	return false
}

// readSourceContents reads all source files from the repo HEAD worktree.
func readSourceContents(repo *git.Repository) map[string]string {
	contents := make(map[string]string)
	if repo == nil {
		return contents
	}

	ref, err := repo.Head()
	if err != nil {
		return contents
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return contents
	}

	tree, err := commit.Tree()
	if err != nil {
		return contents
	}

	tree.Files().ForEach(func(f *object.File) error {
		if !IsSourceExtension(f.Name) {
			return nil
		}
		reader, err := f.Reader()
		if err != nil {
			return nil
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}
		contents[f.Name] = string(data)
		return nil
	})

	return contents
}
