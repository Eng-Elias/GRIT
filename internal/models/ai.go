package models

import "time"

// AISummary is the structured AI-generated codebase analysis.
type AISummary struct {
	Repository   Repository `json:"repository"`
	Description  string     `json:"description"`
	Architecture string     `json:"architecture"`
	TechStack    []string   `json:"tech_stack"`
	RedFlags     []string   `json:"red_flags"`
	EntryPoints  []string   `json:"entry_points"`
	GeneratedAt  time.Time  `json:"generated_at"`
	Model        string     `json:"model"`
	Cached       bool       `json:"cached"`
	CachedAt     *time.Time `json:"cached_at"`
}

// AISummarySummary is abbreviated AI summary data embedded in the main analysis response.
type AISummarySummary struct {
	Status       string `json:"status"`
	AISummaryURL string `json:"ai_summary_url"`
}

// ChatMessage is a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the client-sent payload for the chat endpoint.
type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

// HealthScore is the structured AI-generated repository health assessment.
type HealthScore struct {
	Repository      Repository       `json:"repository"`
	OverallScore    int              `json:"overall_score"`
	Categories      HealthCategories `json:"categories"`
	TopImprovements []string         `json:"top_improvements"`
	GeneratedAt     time.Time        `json:"generated_at"`
	Model           string           `json:"model"`
	Cached          bool             `json:"cached"`
	CachedAt        *time.Time       `json:"cached_at"`
}

// HealthCategories contains the five health assessment categories.
type HealthCategories struct {
	ReadmeQuality       CategoryScore `json:"readme_quality"`
	ContributingGuide   CategoryScore `json:"contributing_guide"`
	CodeDocumentation   CategoryScore `json:"code_documentation"`
	TestCoverageSignals CategoryScore `json:"test_coverage_signals"`
	ProjectHygiene      CategoryScore `json:"project_hygiene"`
}

// CategoryScore is a single health category assessment.
type CategoryScore struct {
	Score int    `json:"score"`
	Notes string `json:"notes"`
}

// ComplexFileSnippet is a snippet from a high-complexity file included in AI context.
type ComplexFileSnippet struct {
	Path       string  `json:"path"`
	Complexity float64 `json:"complexity"`
	Content    string  `json:"content"`
}
