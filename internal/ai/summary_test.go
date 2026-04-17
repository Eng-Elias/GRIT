package ai

import (
	"testing"
)

func TestParseSummary_FullOutput(t *testing.T) {
	raw := `DESCRIPTION:
A web framework for building APIs.

ARCHITECTURE:
MVC pattern with middleware pipeline.

TECH_STACK:
Go, Chi, Redis, PostgreSQL

RED_FLAGS:
No tests, hardcoded secrets

ENTRY_POINTS:
cmd/server/main.go, internal/handler/routes.go`

	s := parseSummary(raw)

	if s.Description != "A web framework for building APIs." {
		t.Errorf("unexpected description: %q", s.Description)
	}
	if s.Architecture != "MVC pattern with middleware pipeline." {
		t.Errorf("unexpected architecture: %q", s.Architecture)
	}
	if len(s.TechStack) != 4 {
		t.Errorf("expected 4 tech stack items, got %d: %v", len(s.TechStack), s.TechStack)
	}
	if len(s.RedFlags) != 2 {
		t.Errorf("expected 2 red flags, got %d: %v", len(s.RedFlags), s.RedFlags)
	}
	if len(s.EntryPoints) != 2 {
		t.Errorf("expected 2 entry points, got %d: %v", len(s.EntryPoints), s.EntryPoints)
	}
}

func TestParseSummary_NoRedFlags(t *testing.T) {
	raw := `DESCRIPTION:
Clean project.

ARCHITECTURE:
Simple.

TECH_STACK:
Go

RED_FLAGS:
none

ENTRY_POINTS:
main.go`

	s := parseSummary(raw)
	if s.RedFlags != nil {
		t.Errorf("expected nil red flags for 'none', got %v", s.RedFlags)
	}
}

func TestParseSummary_MissingSections(t *testing.T) {
	raw := `DESCRIPTION:
Just a description.`

	s := parseSummary(raw)
	if s.Description == "" {
		t.Error("expected description to be parsed")
	}
	if s.Architecture != "" {
		t.Errorf("expected empty architecture, got %q", s.Architecture)
	}
	if s.TechStack != nil {
		t.Errorf("expected nil tech stack, got %v", s.TechStack)
	}
}

func TestExtractSection(t *testing.T) {
	text := "DESCRIPTION:\nHello World\n\nARCHITECTURE:\nSome arch"
	got := extractSection(text, "DESCRIPTION:")
	if got != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", got)
	}
}

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"Go, Redis, NATS", 3},
		{"", 0},
		{"  , , Go , ", 1},
		{"single", 1},
	}
	for _, tt := range tests {
		got := splitCSV(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitCSV(%q) = %d items, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestFormatSSEEvent(t *testing.T) {
	got := FormatSSEEvent("hello")
	want := "data: hello\n\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
