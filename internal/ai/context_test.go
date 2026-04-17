package ai

import (
	"os"
	"strings"
	"testing"

	"github.com/grit-app/grit/internal/models"
)

func TestBuildContext_BasicSections(t *testing.T) {
	in := ContextInput{
		Metadata:   "owner/repo, Go, MIT",
		FileTree:   []string{"main.go", "README.md"},
		Readme:     "# Hello World",
		Manifests:  map[string]string{"go.mod": "module example"},
		DirSummary: "src/ (5 files)",
	}

	parts := BuildContext(in)
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}

	text := parts[0].Text
	for _, want := range []string{"REPOSITORY METADATA", "README", "PACKAGE MANIFESTS", "DIRECTORY STRUCTURE", "FILE TREE"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected section %q in output", want)
		}
	}
}

func TestBuildContext_MissingREADME(t *testing.T) {
	in := ContextInput{
		Metadata:   "owner/repo",
		FileTree:   []string{"main.go"},
		Readme:     "",
		DirSummary: "root (1 file)",
	}

	parts := BuildContext(in)
	text := parts[0].Text
	if strings.Contains(text, "README") {
		t.Error("expected README section to be omitted when empty")
	}
}

func TestBuildContext_EmptyManifests(t *testing.T) {
	in := ContextInput{
		Metadata:   "owner/repo",
		FileTree:   []string{"main.go"},
		Manifests:  map[string]string{},
		DirSummary: "root",
	}

	parts := BuildContext(in)
	text := parts[0].Text
	if strings.Contains(text, "PACKAGE MANIFESTS") {
		t.Error("expected manifests section to be omitted when empty")
	}
}

func TestBuildContext_ComplexFiles(t *testing.T) {
	longContent := strings.Repeat("line\n", 200)
	in := ContextInput{
		Metadata: "owner/repo",
		FileTree: []string{"main.go"},
		ComplexFiles: []models.ComplexFileSnippet{
			{Path: "complex.go", Complexity: 42.5, Content: longContent},
		},
		DirSummary: "root",
	}

	parts := BuildContext(in)
	text := parts[0].Text
	if !strings.Contains(text, "complex.go (complexity: 42.5)") {
		t.Error("expected complex file header in output")
	}
	// Should be truncated to 150 lines.
	snippetStart := strings.Index(text, "complex.go (complexity: 42.5)")
	snippet := text[snippetStart:]
	lineCount := strings.Count(snippet, "line\n")
	if lineCount > maxSnippetLines {
		t.Errorf("expected max %d lines in snippet, got %d", maxSnippetLines, lineCount)
	}
}

func TestBuildContext_TokenBudgetTruncatesFileTree(t *testing.T) {
	// Create a file tree that would vastly exceed the budget if not truncated.
	var tree []string
	for i := range 500_000 {
		tree = append(tree, strings.Repeat("a", 10)+"/"+strings.Repeat("b", 10)+"/file"+string(rune('0'+i%10)))
	}

	in := ContextInput{
		Metadata:   "owner/repo",
		FileTree:   tree,
		DirSummary: "root",
	}

	parts := BuildContext(in)
	text := parts[0].Text
	totalTokens := len(text) / charsPerToken
	if totalTokens > MaxTokenBudget+1000 {
		t.Errorf("token budget exceeded: %d tokens (max %d)", totalTokens, MaxTokenBudget)
	}
}

func TestBuildDirSummary(t *testing.T) {
	paths := []string{"src/main.go", "src/util.go", "README.md", "cmd/app/main.go"}
	summary := buildDirSummary(paths)

	if !strings.Contains(summary, "src (2 files)") {
		t.Error("expected src directory with 2 files")
	}
	if !strings.Contains(summary, "cmd/app (1 files)") {
		t.Error("expected cmd/app directory with 1 file")
	}
	if !strings.Contains(summary, "/ (1 files)") {
		t.Error("expected root directory with 1 file")
	}
}

func TestBuildDirSummary_Empty(t *testing.T) {
	summary := buildDirSummary(nil)
	if summary != "" {
		t.Errorf("expected empty summary, got %q", summary)
	}
}

func TestReadFileFromDisk_Missing(t *testing.T) {
	content := readFileFromDisk("/nonexistent/path/file.txt")
	if content != "" {
		t.Error("expected empty string for missing file")
	}
}

func TestReadFileFromDisk_Valid(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.txt"
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	content := readFileFromDisk(path)
	if content != "hello" {
		t.Errorf("expected 'hello', got %q", content)
	}
}

func TestBuildContext_ReadmeTruncation(t *testing.T) {
	// README longer than 8000 tokens (32000 chars).
	longReadme := strings.Repeat("x", 40_000)
	in := ContextInput{
		Metadata:   "owner/repo",
		FileTree:   []string{"main.go"},
		Readme:     longReadme,
		DirSummary: "root",
	}

	parts := BuildContext(in)
	text := parts[0].Text
	readmeIdx := strings.Index(text, "## README")
	if readmeIdx < 0 {
		t.Fatal("expected README section")
	}
	// The README content should be capped at ~32000 chars.
	readmeContent := text[readmeIdx:]
	nextSection := strings.Index(readmeContent[10:], "## ")
	if nextSection > 0 {
		readmeContent = readmeContent[:nextSection+10]
	}
	readmeChars := strings.Count(readmeContent, "x")
	if readmeChars > maxReadmeTokens*charsPerToken+100 {
		t.Errorf("README not truncated: %d chars (expected max ~%d)", readmeChars, maxReadmeTokens*charsPerToken)
	}
}
