package complexity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCognitiveComplexity_GoEmptyFunction(t *testing.T) {
	source := []byte(`package main
func Empty() {}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	assert.Equal(t, 0, CognitiveComplexity(fn, source, cfg))
}

func TestCognitiveComplexity_GoFlatIf(t *testing.T) {
	source := []byte(`package main
func Flat(x int) int {
	if x > 0 {
		return x
	}
	return 0
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// if at nesting=0: +1+0 = 1
	assert.Equal(t, 1, CognitiveComplexity(fn, source, cfg))
}

func TestCognitiveComplexity_GoNestedIfInIf(t *testing.T) {
	source := []byte(`package main
func Nested(x, y int) int {
	if x > 0 {
		if y > 0 {
			return x + y
		}
	}
	return 0
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// outer if at nesting=0: +1+0 = 1, nesting→1
	// inner if at nesting=1: +1+1 = 2
	// total = 3
	assert.Equal(t, 3, CognitiveComplexity(fn, source, cfg))
}

func TestCognitiveComplexity_GoTripleNestedLoop(t *testing.T) {
	source := []byte(`package main
func Deep(matrix [][][]int) int {
	total := 0
	for _, plane := range matrix {
		for _, row := range plane {
			for _, val := range row {
				total += val
			}
		}
	}
	return total
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// for at nesting=0: +1+0 = 1, nesting→1
	// for at nesting=1: +1+1 = 2, nesting→2
	// for at nesting=2: +1+2 = 3, nesting→3
	// total = 6
	assert.Equal(t, 6, CognitiveComplexity(fn, source, cfg))
}

func TestCognitiveComplexity_GoLogicalOperators(t *testing.T) {
	source := []byte(`package main
func Logic(a, b, c bool) bool {
	if a && b || c {
		return true
	}
	return false
}
`)
	cfg := GetLanguageConfig(".go")
	fn := parseFunctionNode(t, source, ".go")
	require.NotNil(t, fn)
	// if at nesting=0: +1
	// &&: +1 (first logical op)
	// ||: +1 (operator change)
	// total = 3
	assert.Equal(t, 3, CognitiveComplexity(fn, source, cfg))
}

func TestCognitiveComplexity_PythonNestedForIf(t *testing.T) {
	source := []byte(`def process(matrix):
    total = 0
    for row in matrix:
        for val in row:
            if val > 0:
                total += val
    return total
`)
	cfg := GetLanguageConfig(".py")
	fn := parseFunctionNode(t, source, ".py")
	require.NotNil(t, fn)
	// for at nesting=0: +1, nesting→1
	// for at nesting=1: +2, nesting→2
	// if at nesting=2: +3
	// total = 6
	assert.Equal(t, 6, CognitiveComplexity(fn, source, cfg))
}

func TestCognitiveComplexity_JSNestedIfTernary(t *testing.T) {
	source := []byte(`function check(x, y) {
    if (x > 0) {
        return y > 0 ? x + y : x - y;
    }
    return 0;
}
`)
	cfg := GetLanguageConfig(".js")
	fn := parseFunctionNode(t, source, ".js")
	require.NotNil(t, fn)
	// if at nesting=0: +1, nesting→1
	// ternary at nesting=1: +2
	// total = 3
	assert.Equal(t, 3, CognitiveComplexity(fn, source, cfg))
}
