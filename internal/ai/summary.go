package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grit-app/grit/internal/models"
	"google.golang.org/genai"
)

const summaryPrompt = `You are a senior software engineer. Analyze the provided repository context and return a structured codebase summary.

Your response MUST follow this exact format with these section headers:

DESCRIPTION:
A concise 2-3 sentence description of the project's purpose and functionality.

ARCHITECTURE:
A 2-3 sentence description of the project's architecture and design patterns.

TECH_STACK:
A comma-separated list of technologies, languages, and frameworks used.

RED_FLAGS:
A comma-separated list of potential issues, code smells, or concerns. Write "none" if there are no red flags.

ENTRY_POINTS:
A comma-separated list of key entry point files or commands.

Do not add any other sections or commentary.`

// SummaryChunkCallback is called for each streamed text chunk.
type SummaryChunkCallback func(chunk string)

// GenerateSummary streams an AI codebase summary from Gemini and returns the
// parsed AISummary. The onChunk callback is invoked for each streamed text
// chunk so the handler can forward it via SSE.
func GenerateSummary(ctx context.Context, client *Client, contextParts []*genai.Part, repo models.Repository, onChunk SummaryChunkCallback) (*models.AISummary, error) {
	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: contextParts,
		},
		{
			Role: "user",
			Parts: []*genai.Part{
				genai.NewPartFromText(summaryPrompt),
			},
		},
	}

	iterFn, model, err := client.GenerateStream(ctx, contents, nil).Iter()
	if err != nil {
		return nil, err
	}

	var full strings.Builder
	iterFn(func(resp *genai.GenerateContentResponse, iterErr error) bool {
		if iterErr != nil {
			return false
		}
		for _, cand := range resp.Candidates {
			if cand.Content == nil {
				continue
			}
			for _, part := range cand.Content.Parts {
				if part.Text != "" {
					full.WriteString(part.Text)
					if onChunk != nil {
						onChunk(part.Text)
					}
				}
			}
		}
		return true
	})

	summary := parseSummary(full.String())
	summary.Repository = repo
	summary.GeneratedAt = time.Now().UTC()
	summary.Model = model
	return summary, nil
}

// parseSummary extracts structured fields from the raw Gemini text output.
func parseSummary(raw string) *models.AISummary {
	s := &models.AISummary{}

	s.Description = extractSection(raw, "DESCRIPTION:")
	s.Architecture = extractSection(raw, "ARCHITECTURE:")
	s.TechStack = splitCSV(extractSection(raw, "TECH_STACK:"))
	s.RedFlags = splitCSV(extractSection(raw, "RED_FLAGS:"))
	s.EntryPoints = splitCSV(extractSection(raw, "ENTRY_POINTS:"))

	if len(s.RedFlags) == 1 && strings.EqualFold(s.RedFlags[0], "none") {
		s.RedFlags = nil
	}

	return s
}

// extractSection extracts text between the given header and the next header
// (or end of string).
func extractSection(text, header string) string {
	idx := strings.Index(text, header)
	if idx < 0 {
		return ""
	}
	rest := text[idx+len(header):]

	// Find next section header (all-caps word followed by colon).
	headers := []string{"DESCRIPTION:", "ARCHITECTURE:", "TECH_STACK:", "RED_FLAGS:", "ENTRY_POINTS:"}
	end := len(rest)
	for _, h := range headers {
		if h == header {
			continue
		}
		if pos := strings.Index(rest, h); pos >= 0 && pos < end {
			end = pos
		}
	}

	return strings.TrimSpace(rest[:end])
}

// splitCSV splits a comma-separated string into trimmed non-empty items.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// FormatSSEEvent formats a text chunk as an SSE data event.
func FormatSSEEvent(data string) string {
	return fmt.Sprintf("data: %s\n\n", data)
}
