package models

import "time"

type Repository struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	LatestSHA     string `json:"latest_sha"`
	SizeKB        int64  `json:"size_kb,omitempty"`
}

type GitHubMetadata struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	HomepageURL     string    `json:"homepage_url"`
	Stars           int       `json:"stars"`
	Forks           int       `json:"forks"`
	Watchers        int       `json:"watchers"`
	OpenIssues      int       `json:"open_issues"`
	SizeKB          int64     `json:"size_kb"`
	DefaultBranch   string    `json:"default_branch"`
	PrimaryLanguage string    `json:"primary_language"`
	LicenseName     string    `json:"license_name"`
	LicenseSPDX     string    `json:"license_spdx"`
	Topics          []string  `json:"topics"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	HasWiki         bool      `json:"has_wiki"`
	HasProjects     bool      `json:"has_projects"`
	HasDiscussions  bool      `json:"has_discussions"`
	CommunityHealth int       `json:"community_health"`
}

type CommitActivity struct {
	WeeklyCounts []int `json:"weekly_counts"`
	TotalCommits int   `json:"total_commits"`
}
