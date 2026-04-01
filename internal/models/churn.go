package models

import "time"

// FileChurn holds per-file churn metric derived from commit log analysis.
type FileChurn struct {
	Path         string    `json:"path"`
	Churn        int       `json:"churn"`
	LastModified time.Time `json:"last_modified"`
}

// RiskEntry holds per-file risk classification joining churn and complexity data.
type RiskEntry struct {
	Path                  string `json:"path"`
	Churn                 int    `json:"churn"`
	ComplexityCyclomatic  int    `json:"complexity_cyclomatic"`
	Language              string `json:"language"`
	LOC                   int    `json:"loc"`
	RiskLevel             string `json:"risk_level"`
}

// StaleFile represents a potentially unused file identified by the dead code heuristic.
type StaleFile struct {
	Path           string    `json:"path"`
	LastModified   time.Time `json:"last_modified"`
	MonthsInactive int       `json:"months_inactive"`
}

// Thresholds holds percentile thresholds for churn and complexity distributions.
type Thresholds struct {
	ChurnP50      float64 `json:"churn_p50"`
	ChurnP75      float64 `json:"churn_p75"`
	ChurnP90      float64 `json:"churn_p90"`
	ComplexityP50 float64 `json:"complexity_p50"`
	ComplexityP75 float64 `json:"complexity_p75"`
	ComplexityP90 float64 `json:"complexity_p90"`
}

// ChurnMatrixResult is the complete churn analysis output for a repository.
type ChurnMatrixResult struct {
	Repository        Repository   `json:"repository"`
	Churn             []FileChurn  `json:"churn"`
	RiskMatrix        []RiskEntry  `json:"risk_matrix"`
	RiskZone          []RiskEntry  `json:"risk_zone"`
	Thresholds        Thresholds   `json:"thresholds"`
	StaleFiles        []StaleFile  `json:"stale_files"`
	TotalCommits      int          `json:"total_commits"`
	CommitWindowStart time.Time    `json:"commit_window_start"`
	CommitWindowEnd   time.Time    `json:"commit_window_end"`
	TotalFilesChurned int          `json:"total_files_churned"`
	CriticalCount     int          `json:"critical_count"`
	StaleCount        int          `json:"stale_count"`
	AnalyzedAt        time.Time    `json:"analyzed_at"`
	Cached            bool         `json:"cached"`
	CachedAt          *time.Time   `json:"cached_at"`
}

// ChurnSummary is abbreviated churn data embedded in the main analysis response.
type ChurnSummary struct {
	Status        string `json:"status"`
	TotalFiles    int    `json:"total_files"`
	CriticalCount int    `json:"critical_count"`
	StaleCount    int    `json:"stale_count"`
	ChurnMatrixURL string `json:"churn_matrix_url"`
}
