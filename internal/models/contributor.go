package models

import "time"

// Author represents a deduplicated contributor identified by normalized email.
type Author struct {
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	TotalLinesOwned  int       `json:"total_lines_owned"`
	OwnershipPercent float64   `json:"ownership_percent"`
	FilesTouched     int       `json:"files_touched"`
	FirstCommitDate  time.Time `json:"first_commit_date"`
	LastCommitDate   time.Time `json:"last_commit_date"`
	PrimaryLanguages []string  `json:"primary_languages"`
	IsActive         bool      `json:"is_active"`
}

// FileAuthor is a single author's contribution to a specific file.
type FileAuthor struct {
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	LinesOwned       int     `json:"lines_owned"`
	OwnershipPercent float64 `json:"ownership_percent"`
}

// FileContributors is a per-file ownership breakdown.
type FileContributors struct {
	Path       string       `json:"path"`
	TotalLines int          `json:"total_lines"`
	TopAuthors []FileAuthor `json:"top_authors"`
}

// ContributorResult is the complete blame analysis output for a repository.
type ContributorResult struct {
	Repository         Repository         `json:"repository"`
	Authors            []Author           `json:"authors"`
	BusFactor          int                `json:"bus_factor"`
	KeyPeople          []Author           `json:"key_people"`
	FileContributors   []FileContributors `json:"file_contributors"`
	TotalLinesAnalyzed int                `json:"total_lines_analyzed"`
	TotalFilesAnalyzed int                `json:"total_files_analyzed"`
	Partial            bool               `json:"partial"`
	AnalyzedAt         time.Time          `json:"analyzed_at"`
	Cached             bool               `json:"cached"`
	CachedAt           *time.Time         `json:"cached_at"`
}

// AuthorBrief is minimal author info for summary embedding.
type AuthorBrief struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	LinesOwned int    `json:"lines_owned"`
}

// ContributorSummary is abbreviated contributor data embedded in the main analysis response.
type ContributorSummary struct {
	Status          string        `json:"status"`
	BusFactor       int           `json:"bus_factor"`
	TopContributors []AuthorBrief `json:"top_contributors"`
	TotalAuthors    int           `json:"total_authors"`
	ActiveAuthors   int           `json:"active_authors"`
	ContributorsURL string        `json:"contributors_url"`
}
