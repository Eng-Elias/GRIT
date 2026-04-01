package churn

import (
	"sort"

	"github.com/grit-app/grit/internal/models"
)

// ClassifyRisk determines the risk level for a file based on its churn and complexity
// values relative to the computed percentile thresholds.
//
// Rules (evaluated in order):
//  1. critical: churn > p75 AND complexity > p75
//  2. high:     churn > p90 OR complexity > p90
//  3. medium:   churn > p50 OR complexity > p50
//  4. low:      everything else
func ClassifyRisk(churn int, complexity int, t models.Thresholds) string {
	fc := float64(churn)
	cc := float64(complexity)

	if fc > t.ChurnP75 && cc > t.ComplexityP75 {
		return "critical"
	}
	if fc > t.ChurnP90 || cc > t.ComplexityP90 {
		return "high"
	}
	if fc > t.ChurnP50 || cc > t.ComplexityP50 {
		return "medium"
	}
	return "low"
}

// BuildRiskMatrix joins churn data with complexity data by file path and classifies each.
// complexityMap is keyed by file path → cyclomatic complexity.
// languageMap is keyed by file path → language string.
// locMap is keyed by file path → lines of code.
func BuildRiskMatrix(
	files []models.FileChurn,
	complexityMap map[string]int,
	languageMap map[string]string,
	locMap map[string]int,
	thresholds models.Thresholds,
) (matrix []models.RiskEntry, zone []models.RiskEntry) {
	for _, f := range files {
		cc, ok := complexityMap[f.Path]
		if !ok {
			continue
		}

		entry := models.RiskEntry{
			Path:                 f.Path,
			Churn:                f.Churn,
			ComplexityCyclomatic: cc,
			Language:             languageMap[f.Path],
			LOC:                  locMap[f.Path],
			RiskLevel:            ClassifyRisk(f.Churn, cc, thresholds),
		}

		matrix = append(matrix, entry)
		if entry.RiskLevel == "critical" {
			zone = append(zone, entry)
		}
	}

	// Sort risk_zone by churn × complexity descending.
	sort.Slice(zone, func(i, j int) bool {
		return zone[i].Churn*zone[i].ComplexityCyclomatic > zone[j].Churn*zone[j].ComplexityCyclomatic
	})

	return matrix, zone
}
