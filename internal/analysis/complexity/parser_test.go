package complexity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile_GoExtractsFunctions(t *testing.T) {
	source := []byte(`package main

func Hello() string {
	return "hello"
}

func Add(a, b int) int {
	return a + b
}
`)
	fc := ParseFile(context.Background(), "main.go", source, 10)
	require.NotNil(t, fc)
	assert.Equal(t, "Go", fc.Language)
	assert.Equal(t, 2, fc.FunctionCount)
	assert.Equal(t, "Hello", fc.Functions[0].Name)
	assert.Equal(t, "Add", fc.Functions[1].Name)
}

func TestParseFile_PythonExtractsFunctions(t *testing.T) {
	source := []byte(`def greet(name):
    return f"hello {name}"

def add(a, b):
    return a + b
`)
	fc := ParseFile(context.Background(), "util.py", source, 5)
	require.NotNil(t, fc)
	assert.Equal(t, "Python", fc.Language)
	assert.Equal(t, 2, fc.FunctionCount)
}

func TestParseFile_JavaScriptExtractsFunctions(t *testing.T) {
	source := []byte(`function greet(name) {
    return "hello " + name;
}

function add(a, b) {
    return a + b;
}
`)
	fc := ParseFile(context.Background(), "util.js", source, 7)
	require.NotNil(t, fc)
	assert.Equal(t, "JavaScript", fc.Language)
	assert.Equal(t, 2, fc.FunctionCount)
}

func TestParseFile_UnsupportedExtensionReturnsNil(t *testing.T) {
	source := []byte(`{"key": "value"}`)
	fc := ParseFile(context.Background(), "data.json", source, 1)
	assert.Nil(t, fc)
}

func TestParseFile_EmptySourceReturnsZeroFunctions(t *testing.T) {
	source := []byte(`package empty`)
	fc := ParseFile(context.Background(), "empty.go", source, 1)
	require.NotNil(t, fc)
	assert.Equal(t, 0, fc.FunctionCount)
	assert.Empty(t, fc.Functions)
}

func TestParseFile_PopulatesFileMetrics(t *testing.T) {
	source := []byte(`package main

func SimpleFunc() int {
	return 42
}
`)
	fc := ParseFile(context.Background(), "simple.go", source, 5)
	require.NotNil(t, fc)
	assert.Equal(t, "simple.go", fc.Path)
	assert.Equal(t, "Go", fc.Language)
	assert.Equal(t, 5, fc.LOC)
	assert.Equal(t, 1, fc.FunctionCount)
	// SimpleFunc has CC=1 (no decision points).
	assert.Equal(t, 1, fc.Cyclomatic)
	assert.Equal(t, 1, fc.MaxFunctionComplexity)
	assert.Equal(t, 1.0, fc.AvgFunctionComplexity)
}

func TestParseFile_FunctionStartEndLines(t *testing.T) {
	source := []byte(`package main

func First() {
}

func Second() {
}
`)
	fc := ParseFile(context.Background(), "lines.go", source, 7)
	require.NotNil(t, fc)
	require.Len(t, fc.Functions, 2)
	assert.Equal(t, "First", fc.Functions[0].Name)
	assert.True(t, fc.Functions[0].StartLine >= 3)
	assert.Equal(t, "Second", fc.Functions[1].Name)
	assert.True(t, fc.Functions[1].StartLine >= 6)
}
