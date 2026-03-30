package complexity

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/grit-app/grit/internal/models"
)

// Analyzer orchestrates complexity analysis for a repository.
type Analyzer struct{}

// NewAnalyzer creates a new complexity analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze runs complexity analysis on files within a cloned repository.
// It filters to supported languages, parses files via a worker pool,
// aggregates results, and returns a ComplexityResult.
func (a *Analyzer) Analyze(
	ctx context.Context,
	cloneDir string,
	coreFiles []models.FileStats,
	repo models.Repository,
) (*models.ComplexityResult, error) {
	// Filter to supported languages.
	var inputs []FileInput
	for _, f := range coreFiles {
		ext := strings.ToLower(filepath.Ext(f.Path))
		if GetLanguageConfig(ext) != nil {
			inputs = append(inputs, FileInput{
				Path: f.Path,
				LOC:  f.TotalLines,
			})
		}
	}

	slog.Info("complexity: starting analysis",
		"total_files", len(coreFiles),
		"supported_files", len(inputs),
		"repo", repo.FullName,
	)

	if len(inputs) == 0 {
		slog.Info("complexity: no supported files found")
		result := BuildResult(repo, nil)
		return &result, nil
	}

	// Parse files using worker pool.
	pool := NewPool(cloneDir)
	fileResults := pool.Run(ctx, inputs)

	slog.Info("complexity: parsing complete",
		"files_parsed", len(fileResults),
		"files_skipped", len(inputs)-len(fileResults),
	)

	// Build aggregated result.
	result := BuildResult(repo, fileResults)
	return &result, nil
}
