package github

import (
	"context"
	"fmt"
	"time"

	"github.com/grit-app/grit/internal/models"
)

type repoResponse struct {
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	Homepage        string    `json:"homepage"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	WatchersCount   int       `json:"watchers_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	Size            int64     `json:"size"`
	DefaultBranch   string    `json:"default_branch"`
	Language        string    `json:"language"`
	License         *struct {
		Name   string `json:"name"`
		SpdxID string `json:"spdx_id"`
	} `json:"license"`
	Topics         []string  `json:"topics"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	PushedAt       time.Time `json:"pushed_at"`
	HasWiki        bool      `json:"has_wiki"`
	HasProjects    bool      `json:"has_projects"`
	HasDiscussions bool      `json:"has_discussions"`
	Private        bool      `json:"private"`
}

type communityResponse struct {
	HealthPercentage int `json:"health_percentage"`
}

func (c *Client) FetchMetadata(ctx context.Context, owner, repo string) (*models.GitHubMetadata, int64, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	var repoResp repoResponse
	if err := c.getJSON(ctx, url, "repos", &repoResp); err != nil {
		return nil, 0, "", err
	}

	if repoResp.Private && c.token == "" {
		return nil, 0, "", &PrivateRepoError{Owner: owner, Repo: repo}
	}

	meta := &models.GitHubMetadata{
		Name:            repoResp.Name,
		Description:     repoResp.Description,
		HomepageURL:     repoResp.Homepage,
		Stars:           repoResp.StargazersCount,
		Forks:           repoResp.ForksCount,
		Watchers:        repoResp.WatchersCount,
		OpenIssues:      repoResp.OpenIssuesCount,
		SizeKB:          repoResp.Size,
		DefaultBranch:   repoResp.DefaultBranch,
		PrimaryLanguage: repoResp.Language,
		Topics:          repoResp.Topics,
		CreatedAt:       repoResp.CreatedAt,
		UpdatedAt:       repoResp.UpdatedAt,
		PushedAt:        repoResp.PushedAt,
		HasWiki:         repoResp.HasWiki,
		HasProjects:     repoResp.HasProjects,
		HasDiscussions:  repoResp.HasDiscussions,
	}

	if repoResp.License != nil {
		meta.LicenseName = repoResp.License.Name
		meta.LicenseSPDX = repoResp.License.SpdxID
	}

	if meta.Topics == nil {
		meta.Topics = []string{}
	}

	communityURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/community/profile", owner, repo)
	var community communityResponse
	if err := c.getJSON(ctx, communityURL, "community", &community); err == nil {
		meta.CommunityHealth = community.HealthPercentage
	}

	return meta, repoResp.Size, repoResp.DefaultBranch, nil
}
