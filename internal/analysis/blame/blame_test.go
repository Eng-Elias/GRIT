package blame

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"angle brackets", "<Alice@Example.COM>", "alice@example.com"},
		{"no brackets", "alice@example.com", "alice@example.com"},
		{"mixed case", "Alice@Example.COM", "alice@example.com"},
		{"whitespace", "  <alice@example.com>  ", "alice@example.com"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeEmail(tt.input))
		})
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"go file", "main.go", true},
		{"python file", "script.py", true},
		{"javascript file", "app.js", true},
		{"typescript file", "app.ts", true},
		{"jsx file", "component.jsx", true},
		{"tsx file", "component.tsx", true},
		{"java file", "Main.java", true},
		{"ruby file", "app.rb", true},
		{"rust file", "main.rs", true},
		{"c file", "main.c", true},
		{"cpp file", "main.cpp", true},
		{"markdown file", "README.md", false},
		{"yaml file", "config.yaml", false},
		{"json file", "package.json", false},
		{"binary", "image.png", false},
		{"no extension", "Makefile", false},
		{"nested path", "internal/handler/analysis.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSourceFile(tt.path))
		})
	}
}

func TestLanguageForFile(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "Go"},
		{"script.py", "Python"},
		{"app.js", "JavaScript"},
		{"app.ts", "TypeScript"},
		{"Main.java", "Java"},
		{"app.rb", "Ruby"},
		{"main.rs", "Rust"},
		{"main.c", "C"},
		{"main.cpp", "C++"},
		{"README.md", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, LanguageForFile(tt.path))
		})
	}
}

func TestBlameFile_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := BlameFile(ctx, nil, [20]byte{}, "main.go")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBlameFile_UnsupportedExtension(t *testing.T) {
	ctx := context.Background()
	_, err := BlameFile(ctx, nil, [20]byte{}, "README.md")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file extension")
}
