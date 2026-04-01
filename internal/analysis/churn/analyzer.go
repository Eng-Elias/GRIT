package churn

import (
	"encoding/json"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/grit-app/grit/internal/models"
)

// Analyzer orchestrates the full churn analysis pipeline:
// commit log walk → percentile calculation → risk matrix → stale detection.
type Analyzer struct{}

// NewAnalyzer creates a new churn analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze runs the full churn analysis pipeline on the given repository.
// complexityData is the raw JSON of the ComplexityResult from cache (may be nil).
// repoMeta is the repository metadata for the result.
func (a *Analyzer) Analyze(repo *git.Repository, complexityData []byte, repoMeta models.Repository) (*models.ChurnMatrixResult, error) {
	now := time.Now().UTC()

	// Step 1: Walk commit log.
	churnResult, err := WalkCommitLog(repo)
	if err != nil {
		return nil, err
	}

	// Step 2: Compute churn percentiles.
	churnValues := make([]float64, len(churnResult.Files))
	for i, f := range churnResult.Files {
		churnValues[i] = float64(f.Churn)
	}
	churnPercentiles := CalcPercentiles(churnValues)

	// Step 3: Parse complexity data and compute complexity percentiles.
	complexityMap := make(map[string]int)
	languageMap := make(map[string]string)
	locMap := make(map[string]int)
	var complexityValues []float64

	if complexityData != nil {
		var cr models.ComplexityResult
		if json.Unmarshal(complexityData, &cr) == nil {
			for _, fc := range cr.Files {
				complexityMap[fc.Path] = fc.Cyclomatic
				languageMap[fc.Path] = fc.Language
				locMap[fc.Path] = fc.LOC
				complexityValues = append(complexityValues, float64(fc.Cyclomatic))
			}
		}
	}
	complexityPercentiles := CalcPercentiles(complexityValues)

	thresholds := models.Thresholds{
		ChurnP50:      churnPercentiles.P50,
		ChurnP75:      churnPercentiles.P75,
		ChurnP90:      churnPercentiles.P90,
		ComplexityP50: complexityPercentiles.P50,
		ComplexityP75: complexityPercentiles.P75,
		ComplexityP90: complexityPercentiles.P90,
	}

	// Step 4: Build risk matrix.
	matrix, zone := BuildRiskMatrix(churnResult.Files, complexityMap, languageMap, locMap, thresholds)

	// Step 5: Detect stale files.
	staleFiles := FindStaleFiles(churnResult.Files, repo, now)

	// Count critical entries.
	criticalCount := len(zone)
	staleCount := len(staleFiles)

	if matrix == nil {
		matrix = []models.RiskEntry{}
	}
	if zone == nil {
		zone = []models.RiskEntry{}
	}
	if staleFiles == nil {
		staleFiles = []models.StaleFile{}
	}

	return &models.ChurnMatrixResult{
		Repository:        repoMeta,
		Churn:             churnResult.Files,
		RiskMatrix:        matrix,
		RiskZone:          zone,
		Thresholds:        thresholds,
		StaleFiles:        staleFiles,
		TotalCommits:      churnResult.TotalCommits,
		CommitWindowStart: churnResult.CommitWindowStart,
		CommitWindowEnd:   churnResult.CommitWindowEnd,
		TotalFilesChurned: len(churnResult.Files),
		CriticalCount:     criticalCount,
		StaleCount:        staleCount,
		AnalyzedAt:        now,
	}, nil
}
