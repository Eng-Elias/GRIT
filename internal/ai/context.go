package ai

import (
	"fmt"
	"strings"

	"github.com/grit-app/grit/internal/models"
	"google.golang.org/genai"
)

const (
	// MaxTokenBudget is the hard ceiling for all context sections combined.
	MaxTokenBudget = 800_000
	// charsPerToken is the rough approximation used for budget enforcement.
	charsPerToken = 4

	maxReadmeTokens   = 8_000
	maxManifestTokens = 2_000
	maxSnippetLines   = 150
)

// knownManifests lists the package manifests we look for in cloned repos.
var knownManifests = []string{
	"package.json",
	"go.mod",
	"Cargo.toml",
	"requirements.txt",
	"pyproject.toml",
	"pom.xml",
	"build.gradle",
}

// KnownManifests returns the list of manifest filenames to search for.
func KnownManifests() []string { return knownManifests }

// ContextInput bundles everything needed to build the AI context. It is
// constructed by AssembleContext and passed to BuildContext.
type ContextInput struct {
	Metadata     string
	FileTree     []string
	Readme       string
	Manifests    map[string]string
	ComplexFiles []models.ComplexFileSnippet
	DirSummary   string
}

// BuildContext is a pure function that assembles the context sections into
// a []genai.Part, enforcing the token budget. Truncation order (lowest
// priority first): file tree → complex files → manifests → README.
func BuildContext(in ContextInput) []*genai.Part {
	budget := MaxTokenBudget

	// --- Fixed-cost sections ---
	metadataSection := sectionText("REPOSITORY METADATA", in.Metadata)
	budget -= tokenLen(metadataSection)

	dirSection := sectionText("DIRECTORY STRUCTURE", in.DirSummary)
	budget -= tokenLen(dirSection)

	// --- README (capped) ---
	readme := truncateToTokens(in.Readme, maxReadmeTokens)
	readmeSection := ""
	if readme != "" {
		readmeSection = sectionText("README", readme)
		budget -= tokenLen(readmeSection)
	}

	// --- Manifests (capped each) ---
	var manifestParts []string
	for _, name := range knownManifests {
		content, ok := in.Manifests[name]
		if !ok || content == "" {
			continue
		}
		truncated := truncateToTokens(content, maxManifestTokens)
		part := fmt.Sprintf("### %s\n%s", name, truncated)
		manifestParts = append(manifestParts, part)
	}
	manifestSection := ""
	if len(manifestParts) > 0 {
		manifestSection = sectionText("PACKAGE MANIFESTS", strings.Join(manifestParts, "\n\n"))
		budget -= tokenLen(manifestSection)
	}

	// --- Complex files (first 150 lines each) ---
	var complexParts []string
	for _, cf := range in.ComplexFiles {
		lines := strings.Split(cf.Content, "\n")
		if len(lines) > maxSnippetLines {
			lines = lines[:maxSnippetLines]
		}
		snippet := strings.Join(lines, "\n")
		part := fmt.Sprintf("### %s (complexity: %.1f)\n%s", cf.Path, cf.Complexity, snippet)
		tokens := tokenLen(part)
		if tokens > budget {
			break
		}
		complexParts = append(complexParts, part)
		budget -= tokens
	}
	complexSection := ""
	if len(complexParts) > 0 {
		complexSection = sectionText("TOP COMPLEX FILES", strings.Join(complexParts, "\n\n"))
	}

	// --- File tree (uses remaining budget) ---
	treeText := strings.Join(in.FileTree, "\n")
	treeText = truncateToTokens(treeText, budget)
	treeSection := sectionText("FILE TREE", treeText)

	// --- Assemble parts ---
	var text strings.Builder
	text.WriteString(metadataSection)
	if readmeSection != "" {
		text.WriteString(readmeSection)
	}
	if manifestSection != "" {
		text.WriteString(manifestSection)
	}
	if complexSection != "" {
		text.WriteString(complexSection)
	}
	text.WriteString(dirSection)
	text.WriteString(treeSection)

	part := genai.NewPartFromText(text.String())
	return []*genai.Part{part}
}

// sectionText wraps content in a labelled section header.
func sectionText(label, content string) string {
	return fmt.Sprintf("## %s\n\n%s\n\n", label, content)
}

// tokenLen estimates the token count of a string (4 chars ≈ 1 token).
func tokenLen(s string) int {
	n := len(s) / charsPerToken
	if n == 0 && len(s) > 0 {
		n = 1
	}
	return n
}

// truncateToTokens truncates s to approximately maxTokens tokens.
func truncateToTokens(s string, maxTokens int) string {
	maxChars := maxTokens * charsPerToken
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars]
}
