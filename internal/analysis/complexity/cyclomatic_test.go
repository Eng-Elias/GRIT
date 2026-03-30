package complexity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sitter "github.com/smacker/go-tree-sitter"
)

func parseFunctionNode(t *testing.T, source []byte, ext string) *sitter.Node {
	t.Helper()
	cfg := GetLanguageConfig(ext)
	require.NotNil(t, cfg)

	parser := sitter.NewParser()
	parser.SetLanguage(cfg.Language)
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)

	// Find the first function node.
	var fn *sitter.Node
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if fn != nil {
			return
		}
		if isFunctionNode(n.Type(), cfg) {
			fn = n
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	return fn
}

func TestCyclomaticComplexity_GoEmptyFunction(t *testing.T) {
	source := []byte(`package main
func Empty() {}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	assert.Equal(t, 1, CyclomaticComplexity(fn, source, cfg))
}

func TestCyclomaticComplexity_GoIfElse(t *testing.T) {
	source := []byte(`package main
func Check(x int) int {
	if x > 0 {
		return x
	} else {
		return -x
	}
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// 1 base + 1 if = 2
	assert.Equal(t, 2, CyclomaticComplexity(fn, source, cfg))
}

func TestCyclomaticComplexity_GoForLoop(t *testing.T) {
	source := []byte(`package main
func Loop(items []int) int {
	sum := 0
	for _, v := range items {
		sum += v
	}
	return sum
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// 1 base + 1 for = 2
	assert.Equal(t, 2, CyclomaticComplexity(fn, source, cfg))
}

func TestCyclomaticComplexity_GoLogicalOperators(t *testing.T) {
	source := []byte(`package main
func Logic(a, b bool) bool {
	if a && b {
		return true
	}
	return false
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// 1 base + 1 if + 1 && = 3
	assert.Equal(t, 3, CyclomaticComplexity(fn, source, cfg))
}

func TestCyclomaticComplexity_GoSwitch(t *testing.T) {
	source := []byte(`package main
func Classify(x int) string {
	switch {
	case x < 0:
		return "negative"
	case x == 0:
		return "zero"
	case x > 0:
		return "positive"
	}
	return ""
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// 1 base + 3 expression_case = 4
	assert.Equal(t, 4, CyclomaticComplexity(fn, source, cfg))
}

func TestCyclomaticComplexity_PythonIfElifFor(t *testing.T) {
	source := []byte(`def check(items):
    result = []
    for item in items:
        if item > 0:
            result.append(item)
        elif item == 0:
            pass
    return result
`)
	cfg := GetLanguageConfig(".py")
	fn := parseFunctionNode(t, source, ".py")
	require.NotNil(t, fn)
	// 1 base + 1 for + 1 if + 1 elif = 4
	assert.Equal(t, 4, CyclomaticComplexity(fn, source, cfg))
}

func TestCyclomaticComplexity_JSWhileTernary(t *testing.T) {
	source := []byte(`function process(x) {
    let result = x > 0 ? x : -x;
    while (result > 10) {
        result--;
    }
    return result;
}
`)
	cfg := GetLanguageConfig(".js")
	fn := parseFunctionNode(t, source, ".js")
	require.NotNil(t, fn)
	// 1 base + 1 ternary + 1 while = 3
	assert.Equal(t, 3, CyclomaticComplexity(fn, source, cfg))
}
