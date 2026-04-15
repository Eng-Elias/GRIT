package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchPRMergeTimes_Success(t *testing.T) {
	created := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	merged := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC) // 24h later

	resp := graphqlResponse{
		Data: struct {
			Repository struct {
				PullRequests struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []prNode `json:"nodes"`
				} `json:"pullRequests"`
			} `json:"repository"`
		}{},
	}
	resp.Data.Repository.PullRequests.PageInfo.HasNextPage = false
	resp.Data.Repository.PullRequests.Nodes = []prNode{
		{CreatedAt: created, MergedAt: merged},
		{CreatedAt: created, MergedAt: merged.Add(12 * time.Hour)}, // 36h
		{CreatedAt: created, MergedAt: merged.Add(48 * time.Hour)}, // 72h
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("")
	result, err := c.fetchPRMergeTimesFromURL(context.Background(), "owner", "repo", server.URL)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 3, result.SampleSize)
	assert.InDelta(t, 36.0, result.MedianHours, 0.01) // median of [24, 36, 72]
	assert.InDelta(t, 72.0, result.P95Hours, 5.0)
}

func TestFetchPRMergeTimes_EmptyResponse(t *testing.T) {
	resp := graphqlResponse{}
	resp.Data.Repository.PullRequests.Nodes = []prNode{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("")
	result, err := c.fetchPRMergeTimesFromURL(context.Background(), "owner", "repo", server.URL)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestFetchPRMergeTimes_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient("")
	result, err := c.fetchPRMergeTimesFromURL(context.Background(), "owner", "repo", server.URL)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPercentile(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}

	assert.InDelta(t, 30.0, percentile(values, 50), 0.01)
	assert.InDelta(t, 40.0, percentile(values, 75), 1.0)
	assert.InDelta(t, 50.0, percentile(values, 95), 2.5)

	// Single value
	assert.InDelta(t, 42.0, percentile([]float64{42}, 50), 0.01)
}
