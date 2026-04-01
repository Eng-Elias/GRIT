package churn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grit-app/grit/internal/models"
)

func TestAnalyzer_BasicIntegration(t *testing.T) {
	repo := createTestRepo(t, 15)
	analyzer := NewAnalyzer()

	repoMeta := models.Repository{
		Owner:    "test",
		Name:     "repo",
		FullName: "test/repo",
	}

	result, err := analyzer.Analyze(repo, nil, repoMeta)
	require.NoError(t, err)

	assert.Equal(t, "test", result.Repository.Owner)
	assert.Equal(t, "repo", result.Repository.Name)
	assert.NotEmpty(t, result.Churn)
	assert.Equal(t, 15, result.TotalCommits)
	assert.NotZero(t, result.TotalFilesChurned)
	assert.False(t, result.AnalyzedAt.IsZero())

	// Without complexity data, risk_matrix and risk_zone should be empty slices
	assert.Empty(t, result.RiskMatrix)
	assert.Empty(t, result.RiskZone)

	// Thresholds churn values should be populated, complexity should be zero
	assert.NotZero(t, result.Thresholds.ChurnP50)
	assert.Zero(t, result.Thresholds.ComplexityP50)
}

func TestAnalyzer_WithComplexityData(t *testing.T) {
	repo := createTestRepo(t, 10)
	analyzer := NewAnalyzer()

	// Build fake complexity data matching files that exist in test repo
	cr := models.ComplexityResult{
		Files: []models.FileComplexity{
			{Path: "file.txt", Cyclomatic: 12, Language: "Go", LOC: 100},
			{Path: "other.go", Cyclomatic: 5, Language: "Go", LOC: 50},
		},
	}
	complexityData, err := json.Marshal(cr)
	require.NoError(t, err)

	repoMeta := models.Repository{Owner: "test", Name: "repo", FullName: "test/repo"}

	result, err := analyzer.Analyze(repo, complexityData, repoMeta)
	require.NoError(t, err)

	// Should have risk matrix entries where churn and complexity overlap
	assert.NotEmpty(t, result.RiskMatrix)
	assert.NotZero(t, result.Thresholds.ComplexityP50)

	// Verify risk levels are valid
	for _, entry := range result.RiskMatrix {
		assert.Contains(t, []string{"critical", "high", "medium", "low"}, entry.RiskLevel)
	}
}

func TestAnalyzer_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	repo, err := createEmptyRepo(dir)
	require.NoError(t, err)

	analyzer := NewAnalyzer()
	repoMeta := models.Repository{Owner: "test", Name: "empty", FullName: "test/empty"}

	result, err := analyzer.Analyze(repo, nil, repoMeta)
	require.NoError(t, err)

	assert.Empty(t, result.Churn)
	assert.Equal(t, 0, result.TotalCommits)
	assert.Empty(t, result.RiskMatrix)
	assert.Empty(t, result.StaleFiles)
}
