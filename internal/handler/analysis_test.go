package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"bearer token", "Bearer ghp_abc123", "ghp_abc123"},
		{"no header", "", ""},
		{"wrong prefix", "Basic dXNlcjpwYXNz", ""},
		{"bearer only", "Bearer ", ""},
		{"short header", "Bear", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}
			assert.Equal(t, tt.expected, extractToken(r))
		})
	}
}

func TestHandleAnalysis_InvalidOwner(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}", func(w http.ResponseWriter, req *http.Request) {
		if !ownerRepoRegex.MatchString(chi.URLParam(req, "owner")) {
			WriteBadRequest(w, "Invalid repository identifier")
			return
		}
	})

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"valid owner/repo", "/api/facebook/react", http.StatusOK},
		{"owner with spaces", "/api/face%20book/react", http.StatusBadRequest},
		{"owner with slash", "/api/face%2Fbook/react", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestOwnerRepoRegex(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"facebook", true},
		{"go-chi", true},
		{"user.name", true},
		{"under_score", true},
		{"UPPER", true},
		{"123numeric", true},
		{"has space", false},
		{"has/slash", false},
		{"has@at", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.valid, ownerRepoRegex.MatchString(tt.input))
		})
	}
}
