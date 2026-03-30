package complexity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLanguageConfig_SupportedExtensions(t *testing.T) {
	tests := []struct {
		ext      string
		wantName string
	}{
		{".go", "Go"},
		{".ts", "TypeScript"},
		{".js", "JavaScript"},
		{".py", "Python"},
		{".rs", "Rust"},
		{".java", "Java"},
		{".c", "C"},
		{".cpp", "C++"},
		{".rb", "Ruby"},
	}
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			cfg := GetLanguageConfig(tt.ext)
			require.NotNil(t, cfg, "expected config for %s", tt.ext)
			assert.Equal(t, tt.wantName, cfg.Name)
			assert.NotNil(t, cfg.Language, "Language pointer should not be nil")
			assert.NotEmpty(t, cfg.FunctionNodeTypes, "FunctionNodeTypes should not be empty")
			assert.NotEmpty(t, cfg.DecisionPointTypes, "DecisionPointTypes should not be empty")
			assert.NotEmpty(t, cfg.NestingTypes, "NestingTypes should not be empty")
			assert.NotEmpty(t, cfg.LogicalOperators, "LogicalOperators should not be empty")
		})
	}
}

func TestGetLanguageConfig_UnsupportedExtension(t *testing.T) {
	unsupported := []string{".txt", ".md", ".yaml", ".json", ".xml", ".sql", ".sh", ""}
	for _, ext := range unsupported {
		t.Run(ext, func(t *testing.T) {
			cfg := GetLanguageConfig(ext)
			assert.Nil(t, cfg, "expected nil for unsupported extension %q", ext)
		})
	}
}

func TestGetLanguageConfig_Aliases(t *testing.T) {
	aliases := []struct {
		ext        string
		wantSameAs string
	}{
		{".cc", ".cpp"},
		{".cxx", ".cpp"},
		{".hpp", ".cpp"},
		{".h", ".c"},
		{".jsx", ".js"},
		{".tsx", ".ts"},
		{".mjs", ".js"},
		{".cjs", ".js"},
		{".mts", ".ts"},
		{".cts", ".ts"},
	}
	for _, tt := range aliases {
		t.Run(tt.ext+"->"+tt.wantSameAs, func(t *testing.T) {
			cfg := GetLanguageConfig(tt.ext)
			expected := GetLanguageConfig(tt.wantSameAs)
			require.NotNil(t, cfg, "alias %s should resolve", tt.ext)
			require.NotNil(t, expected, "base %s should exist", tt.wantSameAs)
			assert.Equal(t, expected.Name, cfg.Name)
		})
	}
}

func TestSupportedExtensions_ReturnsNonEmpty(t *testing.T) {
	exts := SupportedExtensions()
	assert.True(t, len(exts) >= 9, "should have at least 9 supported extensions, got %d", len(exts))
}

func TestGetLanguageConfig_FunctionNodeTypes(t *testing.T) {
	goConfig := GetLanguageConfig(".go")
	require.NotNil(t, goConfig)
	assert.Contains(t, goConfig.FunctionNodeTypes, "function_declaration")
	assert.Contains(t, goConfig.FunctionNodeTypes, "method_declaration")
	assert.Contains(t, goConfig.FunctionNodeTypes, "func_literal")

	pyConfig := GetLanguageConfig(".py")
	require.NotNil(t, pyConfig)
	assert.Contains(t, pyConfig.FunctionNodeTypes, "function_definition")

	jsConfig := GetLanguageConfig(".js")
	require.NotNil(t, jsConfig)
	assert.Contains(t, jsConfig.FunctionNodeTypes, "function_declaration")
	assert.Contains(t, jsConfig.FunctionNodeTypes, "arrow_function")
}
