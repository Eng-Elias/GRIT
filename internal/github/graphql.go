package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/grit-app/grit/internal/models"
)

const (
	graphqlURL     = "https://api.github.com/graphql"
	prPageSize     = 25
	prMaxPages     = 4
)

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type prNode struct {
	CreatedAt time.Time `json:"createdAt"`
	MergedAt  time.Time `json:"mergedAt"`
}

type graphqlResponse struct {
	Data struct {
		Repository struct {
			PullRequests struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []prNode `json:"nodes"`
			} `json:"pullRequests"`
		} `json:"repository"`
	} `json:"data"`
}

const prQuery = `query($owner: String!, $repo: String!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequests(states: MERGED, first: 25, after: $cursor, orderBy: {field: CREATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes { createdAt mergedAt }
    }
  }
}`

// FetchPRMergeTimes fetches the last 100 merged PRs and computes merge time percentiles.
// Returns nil (not an error) if no data is available.
func (c *Client) FetchPRMergeTimes(ctx context.Context, owner, repo string) (*models.PRMergeTime, error) {
	return c.fetchPRMergeTimesFromURL(ctx, owner, repo, graphqlURL)
}

// fetchPRMergeTimesFromURL is the internal implementation that accepts a URL for testing.
func (c *Client) fetchPRMergeTimesFromURL(ctx context.Context, owner, repo, url string) (*models.PRMergeTime, error) {
	var allNodes []prNode
	var cursor *string

	for page := 0; page < prMaxPages; page++ {
		vars := map[string]interface{}{
			"owner": owner,
			"repo":  repo,
		}
		if cursor != nil {
			vars["cursor"] = *cursor
		}

		reqBody := graphqlRequest{
			Query:     prQuery,
			Variables: vars,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("github graphql: marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("github graphql: create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("github graphql: request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("github graphql: unexpected status %d: %s", resp.StatusCode, string(respBody))
		}

		var gqlResp graphqlResponse
		err = json.NewDecoder(resp.Body).Decode(&gqlResp)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("github graphql: decode response: %w", err)
		}

		nodes := gqlResp.Data.Repository.PullRequests.Nodes
		allNodes = append(allNodes, nodes...)

		if !gqlResp.Data.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		endCursor := gqlResp.Data.Repository.PullRequests.PageInfo.EndCursor
		cursor = &endCursor
	}

	if len(allNodes) == 0 {
		return nil, nil
	}

	// Compute merge durations in hours
	durations := make([]float64, 0, len(allNodes))
	for _, node := range allNodes {
		if node.MergedAt.IsZero() || node.CreatedAt.IsZero() {
			continue
		}
		hours := node.MergedAt.Sub(node.CreatedAt).Hours()
		if hours >= 0 {
			durations = append(durations, hours)
		}
	}

	if len(durations) == 0 {
		return nil, nil
	}

	sort.Float64s(durations)

	return &models.PRMergeTime{
		MedianHours: percentile(durations, 50),
		P75Hours:    percentile(durations, 75),
		P95Hours:    percentile(durations, 95),
		SampleSize:  len(durations),
	}, nil
}

// percentile computes the p-th percentile from a sorted slice of float64s.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	weight := rank - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}
