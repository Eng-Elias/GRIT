package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteError_Structure(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		code       string
		message    string
		wantStatus int
	}{
		{"bad request", http.StatusBadRequest, "bad_request", "invalid input", 400},
		{"not found", http.StatusNotFound, "not_found", "repo not found", 404},
		{"forbidden", http.StatusForbidden, "private_repo", "private repo", 403},
		{"rate limited", http.StatusTooManyRequests, "rate_limited", "rate limited", 429},
		{"unavailable", http.StatusServiceUnavailable, "service_unavailable", "redis down", 503},
		{"timeout", http.StatusGatewayTimeout, "analysis_timeout", "analysis timed out", 504},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.status, tt.code, tt.message)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var resp ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.code, resp.Error)
			assert.Equal(t, tt.message, resp.Message)
		})
	}
}

func TestWriteRateLimited_RetryAfterHeader(t *testing.T) {
	w := httptest.NewRecorder()
	WriteRateLimited(w, "120")

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "120", w.Header().Get("Retry-After"))
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "value", resp["key"])
}
