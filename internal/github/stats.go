package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grit-app/grit/internal/models"
)

type participationResponse struct {
	All   []int `json:"all"`
	Owner []int `json:"owner"`
}

func (c *Client) FetchCommitActivity(ctx context.Context, owner, repo string) (*models.CommitActivity, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	activity := &models.CommitActivity{}

	participation, err := c.fetchParticipationWithRetry(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("github: fetch participation: %w", err)
	}
	activity.WeeklyCounts = participation.All
	if activity.WeeklyCounts == nil {
		activity.WeeklyCounts = []int{}
	}

	total, err := c.fetchTotalCommits(ctx, owner, repo)
	if err != nil {
		activity.TotalCommits = 0
	} else {
		activity.TotalCommits = total
	}

	return activity, nil
}

func (c *Client) fetchParticipationWithRetry(ctx context.Context, owner, repo string) (*participationResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/stats/participation", owner, repo)

	delays := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 8 * time.Second, 8 * time.Second}

	for attempt, delay := range delays {
		resp, err := c.do(ctx, http.MethodGet, url, "stats")
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			var p participationResponse
			err := json.NewDecoder(resp.Body).Decode(&p)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("github: decode participation: %w", err)
			}
			return &p, nil
		}

		resp.Body.Close()

		if resp.StatusCode == http.StatusAccepted {
			if attempt == len(delays)-1 {
				return &participationResponse{All: []int{}}, nil
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		return nil, fmt.Errorf("github: unexpected status %d for participation stats", resp.StatusCode)
	}

	return &participationResponse{All: []int{}}, nil
}

func (c *Client) fetchTotalCommits(ctx context.Context, owner, repo string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=1", owner, repo)

	resp, err := c.do(ctx, http.MethodGet, url, "commits")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github: unexpected status %d for commits", resp.StatusCode)
	}

	link := resp.Header.Get("Link")
	if link == "" {
		return 1, nil
	}

	for _, part := range strings.Split(link, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="last"`) {
			start := strings.Index(part, "page=")
			if start == -1 {
				continue
			}
			start += 5
			end := strings.IndexByte(part[start:], '>')
			if end == -1 {
				end = strings.IndexByte(part[start:], '&')
			}
			if end == -1 {
				end = len(part[start:])
			}
			pageStr := part[start : start+end]
			total, err := strconv.Atoi(pageStr)
			if err != nil {
				return 0, fmt.Errorf("github: parse total commits: %w", err)
			}
			return total, nil
		}
	}

	return 1, nil
}
