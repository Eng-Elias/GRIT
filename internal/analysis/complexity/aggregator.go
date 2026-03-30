package complexity

import (
	"math"
	"sort"
	"time"

	"github.com/grit-app/grit/internal/models"
)

const maxHotFiles = 20

// Aggregate computes repository-level statistics from per-file complexity data.
// Files with 0 functions are excluded from aggregate statistics and distribution.
func Aggregate(files []models.FileComplexity) (
	mean, median, p90 float64,
	totalFunctions int,
	hotFiles []models.FileComplexity,
	dist models.ComplexityDistribution,
) {
	if len(files) == 0 {
		return 0, 0, 0, 0, nil, models.ComplexityDistribution{}
	}

	// Collect cyclomatic values from files with at least 1 function.
	var cyclomatics []float64
	for _, f := range files {
		totalFunctions += f.FunctionCount
		if f.FunctionCount > 0 {
			cyclomatics = append(cyclomatics, float64(f.Cyclomatic))
		}
	}

	if len(cyclomatics) > 0 {
		sort.Float64s(cyclomatics)
		mean = computeMean(cyclomatics)
		median = computeMedian(cyclomatics)
		p90 = computePercentile(cyclomatics, 0.90)
	}

	// Hot files: sorted by complexity_density descending, max 20.
	hotFiles = computeHotFiles(files)

	// Distribution histogram by avg function complexity.
	dist = computeDistribution(files)

	return mean, median, p90, totalFunctions, hotFiles, dist
}

// BuildResult constructs a complete ComplexityResult from file-level data.
func BuildResult(repo models.Repository, files []models.FileComplexity) models.ComplexityResult {
	mean, median, p90, totalFunctions, hotFiles, dist := Aggregate(files)

	return models.ComplexityResult{
		Repository:         repo,
		Files:              files,
		HotFiles:           hotFiles,
		TotalFilesAnalyzed: len(files),
		TotalFunctionCount: totalFunctions,
		MeanComplexity:     math.Round(mean*100) / 100,
		MedianComplexity:   math.Round(median*100) / 100,
		P90Complexity:      math.Round(p90*100) / 100,
		Distribution:       dist,
		AnalyzedAt:         time.Now().UTC(),
	}
}

func computeMean(sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	return sum / float64(len(sorted))
}

func computeMedian(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func computePercentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	rank := p * float64(n-1)
	lower := int(math.Floor(rank))
	upper := lower + 1
	if upper >= n {
		return sorted[n-1]
	}
	weight := rank - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func computeHotFiles(files []models.FileComplexity) []models.FileComplexity {
	// Only include files with functions.
	var candidates []models.FileComplexity
	for _, f := range files {
		if f.FunctionCount > 0 && f.ComplexityDensity > 0 {
			// Copy without functions array to reduce payload size.
			hot := f
			hot.Functions = nil
			candidates = append(candidates, hot)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ComplexityDensity > candidates[j].ComplexityDensity
	})

	if len(candidates) > maxHotFiles {
		candidates = candidates[:maxHotFiles]
	}
	return candidates
}

func computeDistribution(files []models.FileComplexity) models.ComplexityDistribution {
	var dist models.ComplexityDistribution
	for _, f := range files {
		if f.FunctionCount == 0 {
			continue
		}
		avg := f.AvgFunctionComplexity
		switch {
		case avg >= 21:
			dist.Critical++
		case avg >= 11:
			dist.High++
		case avg >= 6:
			dist.Medium++
		default:
			dist.Low++
		}
	}
	return dist
}
