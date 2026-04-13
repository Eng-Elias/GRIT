package complexity

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/grit-app/grit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzer_AnalyzeFixtureRepo(t *testing.T) {
	// Locate testdata relative to this test file.
	repoRoot := locateRepoRoot(t)
	cloneDir := filepath.Join(repoRoot, "testdata", "fixture-repo")

	coreFiles := []models.FileStats{
		{Path: "complexity_sample.go", Language: "Go", TotalLines: 65},
		{Path: "complexity_sample.py", Language: "Python", TotalLines: 35},
		{Path: "complexity_sample.js", Language: "JavaScript", TotalLines: 40},
	}

	repo := models.Repository{Owner: "test", Name: "fixture-repo", FullName: "test/fixture-repo"}
	analyzer := NewAnalyzer()

	result, err := analyzer.Analyze(context.Background(), cloneDir, coreFiles, repo)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test", result.Repository.Owner)
	assert.Equal(t, 3, result.TotalFilesAnalyzed)
	assert.True(t, result.TotalFunctionCount > 0, "should have found functions")
	assert.False(t, result.AnalyzedAt.IsZero())
}

func TestAnalyzer_SkipsUnsupportedLanguages(t *testing.T) {
	repoRoot := locateRepoRoot(t)
	cloneDir := filepath.Join(repoRoot, "testdata", "fixture-repo")

	coreFiles := []models.FileStats{
		{Path: "complexity_sample.go", Language: "Go", TotalLines: 65},
		{Path: "README.md", Language: "Markdown", TotalLines: 10},
		{Path: "data.json", Language: "JSON", TotalLines: 5},
	}

	repo := models.Repository{Owner: "test", Name: "repo"}
	analyzer := NewAnalyzer()

	result, err := analyzer.Analyze(context.Background(), cloneDir, coreFiles, repo)
	require.NoError(t, err)
	// Only the .go file should be analyzed.
	assert.Equal(t, 1, result.TotalFilesAnalyzed)
}

func TestAnalyzer_EmptyFileList(t *testing.T) {
	analyzer := NewAnalyzer()
	repo := models.Repository{Owner: "test", Name: "empty"}

	result, err := analyzer.Analyze(context.Background(), "/nonexistent", nil, repo)
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalFilesAnalyzed)
	assert.Equal(t, 0, result.TotalFunctionCount)
}

func TestAnalyzer_NoSupportedFiles(t *testing.T) {
	analyzer := NewAnalyzer()
	repo := models.Repository{Owner: "test", Name: "docs-only"}

	coreFiles := []models.FileStats{
		{Path: "README.md", Language: "Markdown", TotalLines: 100},
		{Path: "config.yaml", Language: "YAML", TotalLines: 20},
	}

	result, err := analyzer.Analyze(context.Background(), "/tmp", coreFiles, repo)
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalFilesAnalyzed)
}

func locateRepoRoot(t *testing.T) string {
	t.Helper()
	// Walk up from test directory to find go.mod.
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repo root (go.mod)")
		}
		dir = parent
	}
}
