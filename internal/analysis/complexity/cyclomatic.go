package complexity

import (
	"github.com/odvcencio/gotreesitter"
)

// CyclomaticComplexity computes McCabe's cyclomatic complexity for a function node.
// CC = 1 + count(decision_points)
func CyclomaticComplexity(node *gotreesitter.Node, source []byte, cfg *LanguageConfig) int {
	count := 0
	walkCyclomatic(node, source, cfg, &count, node)
	return 1 + count
}

func walkCyclomatic(node *gotreesitter.Node, source []byte, cfg *LanguageConfig, count *int, root *gotreesitter.Node) {
	if node == nil {
		return
	}

	nodeType := node.Type(cfg.Language)

	// Don't recurse into nested function definitions — they get their own score.
	if node != root && isFunctionNode(nodeType, cfg) {
		return
	}

	// The root function node itself does not contribute to scoring.
	if node == root {
		for i := 0; i < node.ChildCount(); i++ {
			child := node.Child(i)
			walkCyclomatic(child, source, cfg, count, root)
		}
		return
	}

	// Count decision points.
	if isDecisionPoint(nodeType, cfg) {
		*count++
	}

	// Count logical operators in binary expressions.
	if nodeType == "binary_expression" || nodeType == "boolean_operator" {
		op := extractOperator(node, source, cfg.Language)
		if isLogicalOperator(op, cfg) {
			*count++
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		walkCyclomatic(child, source, cfg, count, root)
	}
}

// isDecisionPoint checks if the node type is a decision point for the language.
func isDecisionPoint(nodeType string, cfg *LanguageConfig) bool {
	for _, dp := range cfg.DecisionPointTypes {
		if nodeType == dp {
			return true
		}
	}
	return false
}

// isLogicalOperator checks if the operator string is a logical operator for the language.
func isLogicalOperator(op string, cfg *LanguageConfig) bool {
	for _, lo := range cfg.LogicalOperators {
		if op == lo {
			return true
		}
	}
	return false
}

// extractOperator extracts the operator from a binary_expression or boolean_operator node.
func extractOperator(node *gotreesitter.Node, source []byte, lang *gotreesitter.Language) string {
	// The operator is typically an anonymous child node between the two operands.
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if !child.IsNamed() {
			content := child.Text(source)
			switch content {
			case "&&", "||", "and", "or":
				return content
			}
		}
	}
	// Some grammars store the operator in a named "operator" field.
	opNode := node.ChildByFieldName("operator", lang)
	if opNode != nil {
		return opNode.Text(source)
	}
	return ""
}
