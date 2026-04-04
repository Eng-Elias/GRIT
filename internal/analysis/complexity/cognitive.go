package complexity

import (
	"github.com/odvcencio/gotreesitter"
)

// CognitiveComplexity computes cognitive complexity for a function node
// using the SonarSource specification (Ann Campbell, 2017).
//
// Rules:
//  1. +1 for each break in linear flow (if, else if, else, for, while, switch, catch, etc.)
//  2. +nesting_level on top of base increment for nesting-eligible structures
//  3. Nesting level increases when entering nesting structures
//  4. Logical operator sequences: +1 per operator change in a boolean chain
func CognitiveComplexity(node *gotreesitter.Node, source []byte, cfg *LanguageConfig) int {
	score := 0
	walkCognitive(node, source, cfg, 0, &score, "", node)
	return score
}

func walkCognitive(node *gotreesitter.Node, source []byte, cfg *LanguageConfig, nesting int, score *int, lastLogicalOp string, root *gotreesitter.Node) {
	if node == nil {
		return
	}

	nodeType := node.Type(cfg.Language)

	// Don't recurse into nested function definitions — they get their own score.
	if node != root && isFunctionNode(nodeType, cfg) {
		return
	}

	// The root function node itself does not contribute to scoring or nesting.
	if node == root {
		for i := 0; i < node.ChildCount(); i++ {
			child := node.Child(i)
			walkCognitive(child, source, cfg, nesting, score, "", root)
		}
		return
	}

	// Handle logical operators in binary expressions.
	if nodeType == "binary_expression" || nodeType == "boolean_operator" {
		op := extractOperator(node, source, cfg.Language)
		if isLogicalOperator(op, cfg) {
			if op != lastLogicalOp {
				// Operator change or first logical op in sequence: +1
				*score++
			}
			// Recurse with updated lastLogicalOp so consecutive same-ops don't add.
			for i := 0; i < node.ChildCount(); i++ {
				child := node.Child(i)
				walkCognitive(child, source, cfg, nesting, score, op, root)
			}
			return
		}
	}

	// Check if this node is a nesting-incrementing structure.
	incrementsNesting := isNestingType(nodeType, cfg)
	// Check if this node is a cognitive increment point (flow break).
	incrementsScore := isCognitiveIncrement(nodeType, cfg)

	if incrementsScore {
		// Base increment +1, plus nesting level for nesting-eligible nodes.
		if incrementsNesting {
			*score += 1 + nesting
		} else {
			// else clause: +1 but no nesting penalty.
			*score++
		}
	}

	newNesting := nesting
	if incrementsNesting {
		newNesting = nesting + 1
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		walkCognitive(child, source, cfg, newNesting, score, "", root)
	}
}

// isCognitiveIncrement returns true if the node type triggers a cognitive complexity increment.
func isCognitiveIncrement(nodeType string, cfg *LanguageConfig) bool {
	// All nesting types get an increment.
	if isNestingType(nodeType, cfg) {
		return true
	}
	// Decision point types that are not nesting types still break linear flow (+1, no nesting penalty).
	if isDecisionPoint(nodeType, cfg) {
		return true
	}
	// Additionally, else/elif clauses get +1 but don't nest further themselves
	// (they ARE handled as nesting types in some language configs, e.g. Python's elif_clause).
	switch nodeType {
	case "else_clause", "elif_clause", "else":
		return true
	}
	return false
}

// isNestingType checks if the node type increases cognitive nesting depth.
func isNestingType(nodeType string, cfg *LanguageConfig) bool {
	for _, nt := range cfg.NestingTypes {
		if nodeType == nt {
			return true
		}
	}
	return false
}
