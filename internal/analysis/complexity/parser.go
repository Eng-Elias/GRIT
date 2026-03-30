package complexity

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/grit-app/grit/internal/models"
)

// ParseFile parses a single source file and returns its complexity metrics.
// Returns nil if the file extension is unsupported or if parsing fails.
func ParseFile(ctx context.Context, path string, source []byte, loc int) *models.FileComplexity {
	ext := strings.ToLower(filepath.Ext(path))
	cfg := GetLanguageConfig(ext)
	if cfg == nil {
		return nil
	}

	parser := sitter.NewParser()
	parser.SetLanguage(cfg.Language)

	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		slog.Warn("complexity: failed to parse file", "path", path, "error", err)
		return nil
	}
	if tree == nil {
		slog.Warn("complexity: nil tree from parser", "path", path)
		return nil
	}

	rootNode := tree.RootNode()
	if rootNode == nil {
		slog.Warn("complexity: nil root node", "path", path)
		return nil
	}

	functions := extractFunctions(rootNode, source, cfg)

	fc := &models.FileComplexity{
		Path:          path,
		Language:      cfg.Name,
		LOC:           loc,
		FunctionCount: len(functions),
		Functions:     functions,
	}

	// Aggregate per-file metrics from function-level data.
	totalCyclomatic := 0
	totalCognitive := 0
	maxCyclomatic := 0
	for _, fn := range functions {
		totalCyclomatic += fn.Cyclomatic
		totalCognitive += fn.Cognitive
		if fn.Cyclomatic > maxCyclomatic {
			maxCyclomatic = fn.Cyclomatic
		}
	}

	fc.Cyclomatic = totalCyclomatic
	fc.Cognitive = totalCognitive
	fc.MaxFunctionComplexity = maxCyclomatic

	if fc.FunctionCount > 0 {
		fc.AvgFunctionComplexity = float64(totalCyclomatic) / float64(fc.FunctionCount)
	}
	if fc.LOC > 0 {
		fc.ComplexityDensity = float64(totalCyclomatic) / float64(fc.LOC)
	}

	return fc
}

// extractFunctions walks the AST and returns complexity metrics for each function.
func extractFunctions(root *sitter.Node, source []byte, cfg *LanguageConfig) []models.FunctionComplexity {
	var functions []models.FunctionComplexity
	anonCount := 0

	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		if node == nil {
			return
		}

		nodeType := node.Type()
		if isFunctionNode(nodeType, cfg) {
			name := extractFunctionName(node, source, cfg)
			if name == "" {
				anonCount++
				name = "<anonymous>" + anonSuffix(anonCount)
			}

			cc := CyclomaticComplexity(node, source, cfg)
			cog := CognitiveComplexity(node, source, cfg)

			fn := models.FunctionComplexity{
				Name:       name,
				StartLine:  int(node.StartPoint().Row) + 1,
				EndLine:    int(node.EndPoint().Row) + 1,
				Cyclomatic: cc,
				Cognitive:  cog,
			}
			functions = append(functions, fn)
			// Don't recurse into nested functions from here;
			// they are counted separately by the walker.
			return
		}

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			walk(child)
		}
	}

	walk(root)
	return functions
}

// isFunctionNode checks if the node type is a function declaration for the language.
func isFunctionNode(nodeType string, cfg *LanguageConfig) bool {
	for _, ft := range cfg.FunctionNodeTypes {
		if nodeType == ft {
			return true
		}
	}
	return false
}

// extractFunctionName attempts to extract the function name from a function node.
func extractFunctionName(node *sitter.Node, source []byte, cfg *LanguageConfig) string {
	// Most languages have a "name" field on function nodes.
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		return nameNode.Content(source)
	}

	// For method declarations, try "declarator" field (C/C++/Java).
	declNode := node.ChildByFieldName("declarator")
	if declNode != nil {
		// The declarator may itself have a name child (e.g. function_declarator → declarator → identifier).
		inner := declNode.ChildByFieldName("declarator")
		if inner != nil {
			return inner.Content(source)
		}
		innerName := declNode.ChildByFieldName("name")
		if innerName != nil {
			return innerName.Content(source)
		}
		// If it's a simple identifier, return its content.
		if declNode.Type() == "identifier" {
			return declNode.Content(source)
		}
	}

	// For arrow functions assigned to a variable, check parent for variable name.
	_ = cfg // reserved for future language-specific extraction
	return ""
}

func anonSuffix(n int) string {
	if n <= 1 {
		return ""
	}
	return "_" + strings.Repeat("", 0) + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
