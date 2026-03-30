package complexity

import (
	"testing"

	"github.com/grit-app/grit/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAggregate_EmptyInput(t *testing.T) {
	mean, median, p90, totalFn, hotFiles, dist := Aggregate(nil)
	assert.Equal(t, 0.0, mean)
	assert.Equal(t, 0.0, median)
	assert.Equal(t, 0.0, p90)
	assert.Equal(t, 0, totalFn)
	assert.Nil(t, hotFiles)
	assert.Equal(t, models.ComplexityDistribution{}, dist)
}

func TestAggregate_SingleFile(t *testing.T) {
	files := []models.FileComplexity{
		{
			Path:                  "main.go",
			Language:              "Go",
			Cyclomatic:            10,
			Cognitive:             8,
			FunctionCount:         3,
			AvgFunctionComplexity: 3.33,
			MaxFunctionComplexity: 5,
			LOC:                   100,
			ComplexityDensity:     0.1,
		},
	}
	mean, median, p90, totalFn, hotFiles, dist := Aggregate(files)
	assert.Equal(t, 10.0, mean)
	assert.Equal(t, 10.0, median)
	assert.Equal(t, 10.0, p90)
	assert.Equal(t, 3, totalFn)
	assert.Len(t, hotFiles, 1)
	assert.Equal(t, "main.go", hotFiles[0].Path)
	assert.Equal(t, 1, dist.Low) // avg 3.33 → Low bucket
}

func TestAggregate_MultipleFiles(t *testing.T) {
	files := []models.FileComplexity{
		{Cyclomatic: 5, FunctionCount: 2, AvgFunctionComplexity: 2.5, LOC: 50, ComplexityDensity: 0.1},
		{Cyclomatic: 15, FunctionCount: 3, AvgFunctionComplexity: 5.0, LOC: 30, ComplexityDensity: 0.5},
		{Cyclomatic: 30, FunctionCount: 4, AvgFunctionComplexity: 7.5, LOC: 80, ComplexityDensity: 0.375},
		{Cyclomatic: 50, FunctionCount: 5, AvgFunctionComplexity: 10.0, LOC: 100, ComplexityDensity: 0.5},
	}
	mean, median, p90, totalFn, hotFiles, dist := Aggregate(files)

	// Mean of [5, 15, 30, 50] = 25
	assert.Equal(t, 25.0, mean)
	// Median of [5, 15, 30, 50] = (15+30)/2 = 22.5
	assert.Equal(t, 22.5, median)
	// P90 of sorted [5, 15, 30, 50]: rank = 0.9 * 3 = 2.7, lower=2 (30), upper=3 (50)
	// = 30*(1-0.7) + 50*0.7 = 9 + 35 = 44
	assert.Equal(t, 44.0, p90)
	assert.Equal(t, 14, totalFn)
	assert.Len(t, hotFiles, 4)
	// Distribution: avg 2.5→Low, 5.0→Low, 7.5→Medium, 10.0→Medium
	assert.Equal(t, 2, dist.Low)
	assert.Equal(t, 2, dist.Medium)
	assert.Equal(t, 0, dist.High)
	assert.Equal(t, 0, dist.Critical)
}

func TestAggregate_ZeroFunctionFilesExcluded(t *testing.T) {
	files := []models.FileComplexity{
		{Cyclomatic: 0, FunctionCount: 0, AvgFunctionComplexity: 0, LOC: 50, ComplexityDensity: 0},
		{Cyclomatic: 10, FunctionCount: 2, AvgFunctionComplexity: 5.0, LOC: 20, ComplexityDensity: 0.5},
	}
	mean, median, p90, totalFn, _, dist := Aggregate(files)
	// Only file with functions contributes to aggregates.
	assert.Equal(t, 10.0, mean)
	assert.Equal(t, 10.0, median)
	assert.Equal(t, 10.0, p90)
	assert.Equal(t, 2, totalFn)
	assert.Equal(t, 1, dist.Low)
}

func TestComputeHotFiles_MaxCap(t *testing.T) {
	var files []models.FileComplexity
	for i := 0; i < 30; i++ {
		files = append(files, models.FileComplexity{
			Path:              "file_" + itoa(i) + ".go",
			FunctionCount:     1,
			ComplexityDensity: float64(i) * 0.1,
		})
	}
	hotFiles := computeHotFiles(files)
	assert.Len(t, hotFiles, maxHotFiles)
	// First hot file should have highest density.
	assert.Equal(t, "file_29.go", hotFiles[0].Path)
}

func TestComputeHotFiles_SortedByDensityDescending(t *testing.T) {
	files := []models.FileComplexity{
		{Path: "low.go", FunctionCount: 1, ComplexityDensity: 0.1},
		{Path: "high.go", FunctionCount: 1, ComplexityDensity: 0.9},
		{Path: "mid.go", FunctionCount: 1, ComplexityDensity: 0.5},
	}
	hotFiles := computeHotFiles(files)
	assert.Equal(t, "high.go", hotFiles[0].Path)
	assert.Equal(t, "mid.go", hotFiles[1].Path)
	assert.Equal(t, "low.go", hotFiles[2].Path)
}

func TestComputeHotFiles_OmitsFunctionsArray(t *testing.T) {
	files := []models.FileComplexity{
		{
			Path:              "main.go",
			FunctionCount:     2,
			ComplexityDensity: 0.5,
			Functions: []models.FunctionComplexity{
				{Name: "foo", Cyclomatic: 3},
			},
		},
	}
	hotFiles := computeHotFiles(files)
	assert.Len(t, hotFiles, 1)
	assert.Nil(t, hotFiles[0].Functions)
}

func TestComputeDistribution_Buckets(t *testing.T) {
	tests := []struct {
		name     string
		avg      float64
		wantLow  int
		wantMed  int
		wantHigh int
		wantCrit int
	}{
		{"low boundary", 1.0, 1, 0, 0, 0},
		{"low upper", 5.0, 1, 0, 0, 0},
		{"medium lower", 6.0, 0, 1, 0, 0},
		{"medium upper", 10.0, 0, 1, 0, 0},
		{"high lower", 11.0, 0, 0, 1, 0},
		{"high upper", 20.0, 0, 0, 1, 0},
		{"critical lower", 21.0, 0, 0, 0, 1},
		{"critical high", 100.0, 0, 0, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := []models.FileComplexity{
				{FunctionCount: 1, AvgFunctionComplexity: tt.avg},
			}
			dist := computeDistribution(files)
			assert.Equal(t, tt.wantLow, dist.Low, "low")
			assert.Equal(t, tt.wantMed, dist.Medium, "medium")
			assert.Equal(t, tt.wantHigh, dist.High, "high")
			assert.Equal(t, tt.wantCrit, dist.Critical, "critical")
		})
	}
}

func TestBuildResult_PopulatesAllFields(t *testing.T) {
	repo := models.Repository{Owner: "test", Name: "repo"}
	files := []models.FileComplexity{
		{
			Path:                  "main.go",
			Language:              "Go",
			Cyclomatic:            10,
			FunctionCount:         3,
			AvgFunctionComplexity: 3.33,
			LOC:                   50,
			ComplexityDensity:     0.2,
		},
	}
	result := BuildResult(repo, files)
	assert.Equal(t, "test", result.Repository.Owner)
	assert.Equal(t, "repo", result.Repository.Name)
	assert.Len(t, result.Files, 1)
	assert.Equal(t, 1, result.TotalFilesAnalyzed)
	assert.Equal(t, 3, result.TotalFunctionCount)
	assert.Equal(t, 10.0, result.MeanComplexity)
	assert.False(t, result.AnalyzedAt.IsZero())
	assert.False(t, result.Cached)
	assert.Nil(t, result.CachedAt)
}
