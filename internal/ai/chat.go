package ai

import (
	"context"
	"errors"
	"strings"

	"github.com/grit-app/grit/internal/models"
	"google.golang.org/genai"
)

const maxTurns = 20

var (
	ErrEmptyMessages    = errors.New("ai: messages must not be empty")
	ErrLastRoleNotUser  = errors.New("ai: last message role must be 'user'")
	ErrEmptyContent     = errors.New("ai: message content must not be empty or whitespace")
)

// ChatChunkCallback is called for each streamed text chunk during chat.
type ChatChunkCallback func(chunk string)

// ValidateChatRequest checks that the ChatRequest is well-formed.
func ValidateChatRequest(req *models.ChatRequest) error {
	if len(req.Messages) == 0 {
		return ErrEmptyMessages
	}
	last := req.Messages[len(req.Messages)-1]
	if last.Role != "user" {
		return ErrLastRoleNotUser
	}
	for _, m := range req.Messages {
		if strings.TrimSpace(m.Content) == "" {
			return ErrEmptyContent
		}
	}
	return nil
}

// truncateTurns keeps the most recent maxTurns messages, dropping the oldest.
func truncateTurns(messages []models.ChatMessage) []models.ChatMessage {
	if len(messages) <= maxTurns {
		return messages
	}
	return messages[len(messages)-maxTurns:]
}

// GenerateChat streams a chat response from Gemini, prepending repository
// context as a system-level message. The onChunk callback forwards each
// streamed text chunk to the caller.
func GenerateChat(ctx context.Context, client *Client, contextParts []*genai.Part, req *models.ChatRequest, onChunk ChatChunkCallback) error {
	if err := ValidateChatRequest(req); err != nil {
		return err
	}

	messages := truncateTurns(req.Messages)

	// Build contents: system context + conversation turns.
	var contents []*genai.Content

	// Prepend context as the first user message (acts as system context).
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: contextParts,
	})
	contents = append(contents, &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			genai.NewPartFromText("I have reviewed the repository context. How can I help you?"),
		},
	})

	// Append conversation turns.
	for _, m := range messages {
		contents = append(contents, &genai.Content{
			Role: mapRole(m.Role),
			Parts: []*genai.Part{
				genai.NewPartFromText(m.Content),
			},
		})
	}

	iterFn, _, err := client.GenerateStream(ctx, contents, nil).Iter()
	if err != nil {
		return err
	}

	iterFn(func(resp *genai.GenerateContentResponse, iterErr error) bool {
		if iterErr != nil {
			return false
		}
		for _, cand := range resp.Candidates {
			if cand.Content == nil {
				continue
			}
			for _, part := range cand.Content.Parts {
				if part.Text != "" && onChunk != nil {
					onChunk(part.Text)
				}
			}
		}
		return true
	})

	return nil
}

// mapRole converts client role names to Gemini API role names.
func mapRole(role string) string {
	switch role {
	case "assistant", "model":
		return "model"
	default:
		return "user"
	}
}
