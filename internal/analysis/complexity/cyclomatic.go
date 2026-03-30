package complexity

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// CyclomaticComplexity computes McCabe's cyclomatic complexity for a function node.
// CC = 1 + count(decision_points)
func CyclomaticComplexity(node *sitter.Node, source []byte, cfg *LanguageConfig) int {
	count := 0
	walkCyclomatic(node, source, cfg, &count)
	return 1 + count
}

func walkCyclomatic(node *sitter.Node, source []byte, cfg *LanguageConfig, count *int) {
	if node == nil {
		return
	}

	nodeType := node.Type()

	// Don't recurse into nested function definitions — they get their own score.
	if node.Parent() != nil && isFunctionNode(nodeType, cfg) {
		return
	}

	// Count decision points.
	if isDecisionPoint(nodeType, cfg) {
		*count++
	}

	// Count logical operators in binary expressions.
	if nodeType == "binary_expression" || nodeType == "boolean_operator" {
		op := extractOperator(node, source)
		if isLogicalOperator(op, cfg) {
			*count++
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		walkCyclomatic(child, source, cfg, count)
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
func extractOperator(node *sitter.Node, source []byte) string {
	// The operator is typically an anonymous child node between the two operands.
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if !child.IsNamed() {
			content := child.Content(source)
			switch content {
			case "&&", "||", "and", "or":
				return content
			}
		}
	}
	// Some grammars store the operator in a named "operator" field.
	opNode := node.ChildByFieldName("operator")
	if opNode != nil {
		return opNode.Content(source)
	}
	return ""
}
