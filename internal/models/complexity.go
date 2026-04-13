package models

import "time"

// FunctionComplexity holds per-function complexity metrics extracted from AST parsing.
type FunctionComplexity struct {
	Name       string `json:"name"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	Cyclomatic int    `json:"cyclomatic"`
	Cognitive  int    `json:"cognitive"`
}

// FileComplexity holds per-file complexity metrics aggregated from function-level data.
type FileComplexity struct {
	Path                   string               `json:"path"`
	Language               string               `json:"language"`
	Cyclomatic             int                  `json:"cyclomatic"`
	Cognitive              int                  `json:"cognitive"`
	FunctionCount          int                  `json:"function_count"`
	AvgFunctionComplexity  float64              `json:"avg_function_complexity"`
	MaxFunctionComplexity  int                  `json:"max_function_complexity"`
	LOC                    int                  `json:"loc"`
	ComplexityDensity      float64              `json:"complexity_density"`
	Functions              []FunctionComplexity `json:"functions"`
}

// ComplexityDistribution is a histogram of files bucketed by average function complexity.
type ComplexityDistribution struct {
	Low      int `json:"low"`      // avg function complexity 1–5
	Medium   int `json:"medium"`   // avg function complexity 6–10
	High     int `json:"high"`     // avg function complexity 11–20
	Critical int `json:"critical"` // avg function complexity 21+
}

// ComplexityResult is the complete complexity analysis output for a repository.
type ComplexityResult struct {
	Repository         Repository             `json:"repository"`
	Files              []FileComplexity       `json:"files"`
	HotFiles           []FileComplexity       `json:"hot_files"`
	TotalFilesAnalyzed int                    `json:"total_files_analyzed"`
	TotalFunctionCount int                    `json:"total_function_count"`
	MeanComplexity     float64                `json:"mean_complexity"`
	MedianComplexity   float64                `json:"median_complexity"`
	P90Complexity      float64                `json:"p90_complexity"`
	Distribution       ComplexityDistribution `json:"distribution"`
	AnalyzedAt         time.Time              `json:"analyzed_at"`
	Cached             bool                   `json:"cached"`
	CachedAt           *time.Time             `json:"cached_at"`
}

// ComplexitySummary is abbreviated complexity data embedded in the main analysis response.
type ComplexitySummary struct {
	Status             string  `json:"status"`
	MeanComplexity     float64 `json:"mean_complexity"`
	TotalFunctionCount int     `json:"total_function_count"`
	HotFileCount       int     `json:"hot_file_count"`
	ComplexityURL      string  `json:"complexity_url"`
}
