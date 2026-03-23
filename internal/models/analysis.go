package models

import "time"

type FileStats struct {
	Path         string `json:"path"`
	Language     string `json:"language"`
	TotalLines   int    `json:"total_lines"`
	CodeLines    int    `json:"code_lines"`
	CommentLines int    `json:"comment_lines"`
	BlankLines   int    `json:"blank_lines"`
	ByteSize     int64  `json:"byte_size"`
}

type LanguageBreakdown struct {
	Language     string  `json:"language"`
	FileCount    int     `json:"file_count"`
	TotalLines   int     `json:"total_lines"`
	CodeLines    int     `json:"code_lines"`
	CommentLines int     `json:"comment_lines"`
	BlankLines   int     `json:"blank_lines"`
	Percentage   float64 `json:"percentage"`
}

type AnalysisResult struct {
	Repository        Repository          `json:"repository"`
	Metadata          GitHubMetadata      `json:"metadata"`
	CommitActivity    CommitActivity      `json:"commit_activity"`
	Files             []FileStats         `json:"files"`
	Languages         []LanguageBreakdown `json:"languages"`
	TotalFiles        int                 `json:"total_files"`
	TotalLines        int                 `json:"total_lines"`
	TotalCodeLines    int                 `json:"total_code_lines"`
	TotalCommentLines int                 `json:"total_comment_lines"`
	TotalBlankLines   int                 `json:"total_blank_lines"`
	AnalyzedAt        time.Time           `json:"analyzed_at"`
	Cached            bool                `json:"cached"`
	CachedAt          *time.Time          `json:"cached_at"`
}
