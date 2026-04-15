package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/grit-app/grit/internal/models"
	"google.golang.org/genai"
)

const healthPrompt = `You are a senior software engineer. Analyze the provided repository context and return a JSON health assessment.

You MUST respond with ONLY valid JSON matching this exact schema — no markdown, no commentary:

{
  "overall_score": <int 0-100>,
  "categories": {
    "readme_quality":        { "score": <int 0-100>, "notes": "<string>" },
    "contributing_guide":    { "score": <int 0-100>, "notes": "<string>" },
    "code_documentation":    { "score": <int 0-100>, "notes": "<string>" },
    "test_coverage_signals": { "score": <int 0-100>, "notes": "<string>" },
    "project_hygiene":       { "score": <int 0-100>, "notes": "<string>" }
  },
  "top_improvements": ["<string>", "<string>", "<string>"]
}

Score guidelines:
- 80-100: Excellent — well-maintained, documented, tested
- 60-79: Good — mostly solid with minor gaps
- 40-59: Fair — noticeable gaps in quality or documentation
- 20-39: Poor — significant issues
- 0-19: Critical — major problems across the board`

const healthRetryPrompt = `Your previous response was not valid JSON. You MUST respond with ONLY valid JSON — no markdown fences, no explanation. Use this exact schema:

{"overall_score":<int>,"categories":{"readme_quality":{"score":<int>,"notes":"<string>"},"contributing_guide":{"score":<int>,"notes":"<string>"},"code_documentation":{"score":<int>,"notes":"<string>"},"test_coverage_signals":{"score":<int>,"notes":"<string>"},"project_hygiene":{"score":<int>,"notes":"<string>"}},"top_improvements":["<string>"]}`

// GenerateHealth generates a structured JSON health score from Gemini.
// On JSON parse failure it retries once with a stricter prompt.
func GenerateHealth(ctx context.Context, client *Client, contextParts []*genai.Part, repo models.Repository) (*models.HealthScore, error) {
	cfg := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	contents := []*genai.Content{
		{Role: "user", Parts: contextParts},
		{Role: "user", Parts: []*genai.Part{genai.NewPartFromText(healthPrompt)}},
	}

	resp, model, err := client.Generate(ctx, contents, cfg)
	if err != nil {
		return nil, err
	}

	hs, parseErr := parseHealthResponse(resp)
	if parseErr != nil {
		slog.Warn("ai: health JSON parse failed, retrying with stricter prompt", "error", parseErr)

		// Retry with stricter prompt.
		retryContents := append(contents, &genai.Content{
			Role: "user",
			Parts: []*genai.Part{genai.NewPartFromText(healthRetryPrompt)},
		})
		resp, model, err = client.Generate(ctx, retryContents, cfg)
		if err != nil {
			return nil, err
		}
		hs, parseErr = parseHealthResponse(resp)
		if parseErr != nil {
			return nil, fmt.Errorf("ai: health score JSON parse failed after retry: %w", parseErr)
		}
	}

	hs.Repository = repo
	hs.GeneratedAt = time.Now().UTC()
	hs.Model = model
	return hs, nil
}

// parseHealthResponse extracts text from the response and parses it as a HealthScore.
func parseHealthResponse(resp *genai.GenerateContentResponse) (*models.HealthScore, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	var text string
	for _, cand := range resp.Candidates {
		if cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if part.Text != "" {
				text += part.Text
			}
		}
	}

	if text == "" {
		return nil, fmt.Errorf("no text in response")
	}

	var hs models.HealthScore
	if err := json.Unmarshal([]byte(text), &hs); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	// Clamp overall score.
	if hs.OverallScore < 0 {
		hs.OverallScore = 0
	}
	if hs.OverallScore > 100 {
		hs.OverallScore = 100
	}

	return &hs, nil
}
