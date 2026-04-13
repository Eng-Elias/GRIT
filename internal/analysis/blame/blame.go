package blame

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// LineBlameDatum holds the blame attribution for a single line.
type LineBlameDatum struct {
	AuthorName  string
	AuthorEmail string
	Date        time.Time
}

// FileBlameResult holds blame data for a single file.
type FileBlameResult struct {
	Path     string
	Language string
	Lines    []LineBlameDatum
}

// BlameFile runs git blame on a single file and returns per-line attribution.
// It normalizes emails (strip angle brackets, lowercase) and identifies the language from extension.
func BlameFile(ctx context.Context, repo *git.Repository, commitHash plumbing.Hash, path string) (*FileBlameResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	lang := LanguageForFile(path)
	if lang == "" {
		return nil, fmt.Errorf("blame: unsupported file extension: %s", path)
	}

	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("blame: resolve commit %s: %w", commitHash, err)
	}

	result, err := git.Blame(commit, path)
	if err != nil {
		return nil, fmt.Errorf("blame: %s: %w", path, err)
	}

	lines := make([]LineBlameDatum, 0, len(result.Lines))
	for _, line := range result.Lines {
		// Line.Author is the email, Line.AuthorName is the display name.
		email := normalizeEmail(line.Author)
		if email == "" {
			continue
		}
		lines = append(lines, LineBlameDatum{
			AuthorName:  line.AuthorName,
			AuthorEmail: email,
			Date:        line.Date,
		})
	}

	return &FileBlameResult{
		Path:     path,
		Language: lang,
		Lines:    lines,
	}, nil
}

// blameFileFromCommit is a test-friendly variant that accepts a *object.Commit directly.
func blameFileFromCommit(ctx context.Context, commit *object.Commit, path, lang string) (*FileBlameResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	result, err := git.Blame(commit, path)
	if err != nil {
		return nil, fmt.Errorf("blame: %s: %w", path, err)
	}

	lines := make([]LineBlameDatum, 0, len(result.Lines))
	for _, line := range result.Lines {
		email := normalizeEmail(line.Author)
		if email == "" {
			continue
		}
		lines = append(lines, LineBlameDatum{
			AuthorName:  line.AuthorName,
			AuthorEmail: email,
			Date:        line.Date,
		})
	}

	return &FileBlameResult{
		Path:     path,
		Language: lang,
		Lines:    lines,
	}, nil
}

// normalizeEmail strips angle brackets and lowercases the email address.
func normalizeEmail(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(s, ">")
	return strings.ToLower(s)
}
