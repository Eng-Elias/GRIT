package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountLines_GoFile(t *testing.T) {
	lang := LookupLanguage(".go", "main.go")
	input := "package main\n\n// This is a comment\nimport \"fmt\"\n\n/*\nMulti-line comment\n*/\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"
	counts := CountLines(strings.NewReader(input), lang)

	// Lines: package(code) blank //(comment) import(code) blank /*(comment) Multi(comment) */(comment) blank func(code) Println(code) }(code)
	assert.Equal(t, 12, counts.Total)
	assert.Equal(t, 5, counts.Code)
	assert.Equal(t, 4, counts.Comment)
	assert.Equal(t, 3, counts.Blank)
}

func TestCountLines_PythonFile(t *testing.T) {
	lang := LookupLanguage(".py", "app.py")
	input := `# A Python file

"""
Docstring block
"""

def hello():
    print("hi")
`
	counts := CountLines(strings.NewReader(input), lang)

	assert.Equal(t, 8, counts.Total)
	assert.Equal(t, 2, counts.Code)
	assert.Equal(t, 4, counts.Comment) // # comment + """ + docstring + """
	assert.Equal(t, 2, counts.Blank)
}

func TestCountLines_BlankOnly(t *testing.T) {
	lang := LanguageDef{Name: "Text"}
	input := `

`
	counts := CountLines(strings.NewReader(input), lang)
	assert.Equal(t, 2, counts.Total)
	assert.Equal(t, 0, counts.Code)
	assert.Equal(t, 0, counts.Comment)
	assert.Equal(t, 2, counts.Blank)
}

func TestCountLines_Empty(t *testing.T) {
	lang := LanguageDef{Name: "Text"}
	counts := CountLines(strings.NewReader(""), lang)
	assert.Equal(t, 0, counts.Total)
}

func TestCountLines_CodeOnly(t *testing.T) {
	lang := LookupLanguage(".go", "main.go")
	input := `package main
func main() {}
`
	counts := CountLines(strings.NewReader(input), lang)
	assert.Equal(t, 2, counts.Total)
	assert.Equal(t, 2, counts.Code)
	assert.Equal(t, 0, counts.Comment)
	assert.Equal(t, 0, counts.Blank)
}

func TestCountLines_HTMLComments(t *testing.T) {
	lang := LookupLanguage(".html", "index.html")
	input := `<html>
<!-- This is a comment -->
<body>
<!--
Multi-line
-->
</body>
</html>
`
	counts := CountLines(strings.NewReader(input), lang)
	assert.Equal(t, 8, counts.Total)
	assert.Equal(t, 4, counts.Code)
	assert.Equal(t, 4, counts.Comment)
	assert.Equal(t, 0, counts.Blank)
}

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"text", []byte("hello world"), false},
		{"binary", []byte{0x00, 0x01, 0x02}, true},
		{"empty", []byte{}, false},
		{"text with newlines", []byte("line1\nline2\n"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsBinary(tt.data))
		})
	}
}
