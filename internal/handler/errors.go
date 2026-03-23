package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   code,
		Message: message,
	})
}

func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "bad_request", message)
}

func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "not_found", message)
}

func WritePrivateRepo(w http.ResponseWriter) {
	WriteError(w, http.StatusForbidden, "private_repo",
		"This repository is private. Provide a GitHub token via Authorization header.")
}

func WriteRateLimited(w http.ResponseWriter, retryAfter string) {
	if retryAfter != "" {
		w.Header().Set("Retry-After", retryAfter)
	}
	WriteError(w, http.StatusTooManyRequests, "rate_limited",
		"GitHub API rate limit exceeded. Try again later.")
}

func WriteServiceUnavailable(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusServiceUnavailable, "service_unavailable", message)
}

func WriteTimeout(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusGatewayTimeout, "analysis_timeout", message)
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
