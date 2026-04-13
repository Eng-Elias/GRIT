package churn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/grit-app/grit/internal/models"
)

func TestClassifyRisk(t *testing.T) {
	thresholds := models.Thresholds{
		ChurnP50:      5,
		ChurnP75:      10,
		ChurnP90:      20,
		ComplexityP50: 3,
		ComplexityP75: 8,
		ComplexityP90: 15,
	}

	tests := []struct {
		name       string
		churn      int
		complexity int
		want       string
	}{
		{"critical: both above p75", 15, 10, "critical"},
		{"critical: both well above p75", 25, 20, "critical"},
		{"high: churn above p90 only", 25, 2, "high"},
		{"high: complexity above p90 only", 3, 20, "high"},
		{"medium: churn above p50", 8, 1, "medium"},
		{"medium: complexity above p50", 1, 5, "medium"},
		{"low: both below p50", 3, 2, "low"},
		{"low: both at zero", 0, 0, "low"},
		{"boundary: churn exactly at p75, complexity exactly at p75", 10, 8, "medium"},
		{"boundary: churn just above p75, complexity just above p75", 11, 9, "critical"},
		{"boundary: churn exactly at p90", 20, 1, "medium"},
		{"boundary: churn just above p90", 21, 1, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyRisk(tt.churn, tt.complexity, thresholds)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClassifyRisk_ZeroThresholds(t *testing.T) {
	thresholds := models.Thresholds{}

	// With zero thresholds, everything is "low" since nothing is > 0
	assert.Equal(t, "low", ClassifyRisk(0, 0, thresholds))
	// Any positive value is > 0, so should be critical (both > p75=0)
	assert.Equal(t, "critical", ClassifyRisk(1, 1, thresholds))
}

func TestBuildRiskMatrix_BasicJoin(t *testing.T) {
	files := []models.FileChurn{
		{Path: "main.go", Churn: 20, LastModified: time.Now()},
		{Path: "utils.go", Churn: 5, LastModified: time.Now()},
		{Path: "readme.md", Churn: 2, LastModified: time.Now()},
	}

	complexityMap := map[string]int{
		"main.go":  15,
		"utils.go": 3,
		// readme.md has no complexity data → excluded from matrix
	}
	languageMap := map[string]string{
		"main.go":  "Go",
		"utils.go": "Go",
	}
	locMap := map[string]int{
		"main.go":  200,
		"utils.go": 50,
	}

	thresholds := models.Thresholds{
		ChurnP50: 5, ChurnP75: 10, ChurnP90: 20,
		ComplexityP50: 3, ComplexityP75: 8, ComplexityP90: 15,
	}

	matrix, zone := BuildRiskMatrix(files, complexityMap, languageMap, locMap, thresholds)

	// Only files with complexity data should be in matrix
	assert.Len(t, matrix, 2)
	assert.Equal(t, "main.go", matrix[0].Path)
	assert.Equal(t, "critical", matrix[0].RiskLevel)
	assert.Equal(t, "Go", matrix[0].Language)
	assert.Equal(t, 200, matrix[0].LOC)

	// utils.go: churn=5 (== p50, not >), complexity=3 (== p50, not >) → low
	assert.Equal(t, "low", matrix[1].RiskLevel)

	// Risk zone should contain only critical entries
	assert.Len(t, zone, 1)
	assert.Equal(t, "main.go", zone[0].Path)
}

func TestBuildRiskMatrix_EmptyInput(t *testing.T) {
	matrix, zone := BuildRiskMatrix(nil, nil, nil, nil, models.Thresholds{})
	assert.Nil(t, matrix)
	assert.Nil(t, zone)
}

func TestBuildRiskMatrix_NoComplexityMatch(t *testing.T) {
	files := []models.FileChurn{
		{Path: "a.go", Churn: 10, LastModified: time.Now()},
	}
	matrix, zone := BuildRiskMatrix(files, map[string]int{}, nil, nil, models.Thresholds{})
	assert.Nil(t, matrix)
	assert.Nil(t, zone)
}

func TestBuildRiskMatrix_RiskZoneSortedByProduct(t *testing.T) {
	files := []models.FileChurn{
		{Path: "a.go", Churn: 15, LastModified: time.Now()},
		{Path: "b.go", Churn: 20, LastModified: time.Now()},
	}
	complexityMap := map[string]int{"a.go": 20, "b.go": 10}
	languageMap := map[string]string{"a.go": "Go", "b.go": "Go"}
	locMap := map[string]int{"a.go": 100, "b.go": 100}

	thresholds := models.Thresholds{
		ChurnP50: 1, ChurnP75: 2, ChurnP90: 3,
		ComplexityP50: 1, ComplexityP75: 2, ComplexityP90: 3,
	}

	_, zone := BuildRiskMatrix(files, complexityMap, languageMap, locMap, thresholds)

	assert.Len(t, zone, 2)
	// a.go: 15*20=300, b.go: 20*10=200 → a.go first
	assert.Equal(t, "a.go", zone[0].Path)
	assert.Equal(t, "b.go", zone[1].Path)
}
