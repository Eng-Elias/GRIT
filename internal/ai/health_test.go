package ai

import (
	"testing"

	"google.golang.org/genai"
)

func TestParseHealthResponse_Valid(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromText(`{
							"overall_score": 75,
							"categories": {
								"readme_quality": {"score": 80, "notes": "Good README"},
								"contributing_guide": {"score": 60, "notes": "Basic guide"},
								"code_documentation": {"score": 70, "notes": "Inline comments"},
								"test_coverage_signals": {"score": 85, "notes": "Good coverage"},
								"project_hygiene": {"score": 80, "notes": "Clean structure"}
							},
							"top_improvements": ["Add CONTRIBUTING.md", "Increase test coverage"]
						}`),
					},
				},
			},
		},
	}

	hs, err := parseHealthResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hs.OverallScore != 75 {
		t.Errorf("expected overall_score 75, got %d", hs.OverallScore)
	}
	if hs.Categories.ReadmeQuality.Score != 80 {
		t.Errorf("expected readme_quality score 80, got %d", hs.Categories.ReadmeQuality.Score)
	}
	if len(hs.TopImprovements) != 2 {
		t.Errorf("expected 2 improvements, got %d", len(hs.TopImprovements))
	}
}

func TestParseHealthResponse_InvalidJSON(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromText(`not valid json`),
					},
				},
			},
		},
	}

	_, err := parseHealthResponse(resp)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseHealthResponse_EmptyResponse(t *testing.T) {
	_, err := parseHealthResponse(nil)
	if err == nil {
		t.Error("expected error for nil response")
	}

	resp := &genai.GenerateContentResponse{Candidates: []*genai.Candidate{}}
	_, err = parseHealthResponse(resp)
	if err == nil {
		t.Error("expected error for empty candidates")
	}
}

func TestParseHealthResponse_NilContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: nil},
		},
	}
	_, err := parseHealthResponse(resp)
	if err == nil {
		t.Error("expected error for nil content")
	}
}

func TestParseHealthResponse_ClampScore(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromText(`{
							"overall_score": 150,
							"categories": {
								"readme_quality": {"score": 80, "notes": ""},
								"contributing_guide": {"score": 60, "notes": ""},
								"code_documentation": {"score": 70, "notes": ""},
								"test_coverage_signals": {"score": 85, "notes": ""},
								"project_hygiene": {"score": 80, "notes": ""}
							},
							"top_improvements": []
						}`),
					},
				},
			},
		},
	}

	hs, err := parseHealthResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hs.OverallScore != 100 {
		t.Errorf("expected clamped score 100, got %d", hs.OverallScore)
	}
}
