package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient("test-token")
	resp, err := client.do(context.Background(), http.MethodGet, srv.URL+"/test", "test")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "Bearer test-token", gotAuth)
}

func TestClient_NoAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient("")
	resp, err := client.do(context.Background(), http.MethodGet, srv.URL+"/test", "test")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "", gotAuth)
}

func TestClient_RateLimitDetection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "9999999999")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"rate limit exceeded"}`))
	}))
	defer srv.Close()

	client := NewClient("")
	_, err := client.do(context.Background(), http.MethodGet, srv.URL+"/test", "test")
	require.Error(t, err)

	var rateLimitErr *RateLimitError
	assert.ErrorAs(t, err, &rateLimitErr)
}

func TestClient_NotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	client := NewClient("")
	var result map[string]interface{}
	err := client.getJSON(context.Background(), srv.URL+"/test", "test", &result)
	require.Error(t, err)

	var notFoundErr *NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)
}

func TestClient_PrivateRepoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	client := NewClient("")
	var result map[string]interface{}
	err := client.getJSON(context.Background(), srv.URL+"/test", "test", &result)
	require.Error(t, err)

	var privateErr *PrivateRepoError
	assert.ErrorAs(t, err, &privateErr)
}

func TestClient_ETAGHeader(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient("")
	resp, err := client.do(context.Background(), http.MethodGet, srv.URL+"/test", "test")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "application/vnd.github.v3+json", gotAccept)
}
