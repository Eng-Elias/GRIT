package ai

import (
	"errors"
	"testing"

	"github.com/grit-app/grit/internal/models"
)

func TestValidateChatRequest_Valid(t *testing.T) {
	req := &models.ChatRequest{
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}
	if err := ValidateChatRequest(req); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateChatRequest_Empty(t *testing.T) {
	req := &models.ChatRequest{Messages: nil}
	if !errors.Is(ValidateChatRequest(req), ErrEmptyMessages) {
		t.Error("expected ErrEmptyMessages")
	}
}

func TestValidateChatRequest_LastRoleNotUser(t *testing.T) {
	req := &models.ChatRequest{
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hi"},
			{Role: "assistant", Content: "Hello"},
		},
	}
	if !errors.Is(ValidateChatRequest(req), ErrLastRoleNotUser) {
		t.Error("expected ErrLastRoleNotUser")
	}
}

func TestValidateChatRequest_WhitespaceContent(t *testing.T) {
	req := &models.ChatRequest{
		Messages: []models.ChatMessage{
			{Role: "user", Content: "   "},
		},
	}
	if !errors.Is(ValidateChatRequest(req), ErrEmptyContent) {
		t.Error("expected ErrEmptyContent for whitespace-only content")
	}
}

func TestTruncateTurns_UnderLimit(t *testing.T) {
	msgs := make([]models.ChatMessage, 5)
	for i := range msgs {
		msgs[i] = models.ChatMessage{Role: "user", Content: "msg"}
	}
	result := truncateTurns(msgs)
	if len(result) != 5 {
		t.Errorf("expected 5, got %d", len(result))
	}
}

func TestTruncateTurns_OverLimit(t *testing.T) {
	msgs := make([]models.ChatMessage, 25)
	for i := range msgs {
		msgs[i] = models.ChatMessage{Role: "user", Content: "msg"}
	}
	result := truncateTurns(msgs)
	if len(result) != maxTurns {
		t.Errorf("expected %d, got %d", maxTurns, len(result))
	}
	// Should keep the most recent messages.
	if &result[0] == &msgs[0] {
		t.Error("expected oldest messages to be dropped")
	}
}

func TestMapRole(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"user", "user"},
		{"assistant", "model"},
		{"model", "model"},
		{"system", "user"},
		{"", "user"},
	}
	for _, tt := range tests {
		got := mapRole(tt.input)
		if got != tt.want {
			t.Errorf("mapRole(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
