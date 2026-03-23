package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	gritmetrics "github.com/grit-app/grit/internal/metrics"
)

type RateLimitError struct {
	RetryAfter string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("github: rate limited, retry after %s seconds", e.RetryAfter)
}

type NotFoundError struct {
	Owner string
	Repo  string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("github: repository %s/%s not found", e.Owner, e.Repo)
}

type PrivateRepoError struct {
	Owner string
	Repo  string
}

func (e *PrivateRepoError) Error() string {
	return fmt.Sprintf("github: repository %s/%s is private", e.Owner, e.Repo)
}

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		token: token,
	}
}

func (c *Client) do(ctx context.Context, method, url, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	gritmetrics.GitHubAPIRequestsTotal.WithLabelValues(endpoint).Inc()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: request failed: %w", err)
	}

	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining == "0" && resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		reset := resp.Header.Get("X-RateLimit-Reset")
		retryAfter := "60"
		if reset != "" {
			if resetTime, err := strconv.ParseInt(reset, 10, 64); err == nil {
				seconds := resetTime - time.Now().Unix()
				if seconds > 0 {
					retryAfter = strconv.FormatInt(seconds, 10)
				}
			}
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}
	}

	return resp, nil
}

func (c *Client) getJSON(ctx context.Context, url, endpoint string, v interface{}) error {
	resp, err := c.do(ctx, http.MethodGet, url, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{}
	}

	if resp.StatusCode == http.StatusForbidden {
		return &PrivateRepoError{}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
