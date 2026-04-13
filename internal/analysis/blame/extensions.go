package blame

import (
	"path/filepath"
	"strings"
)

// SupportedSourceExtensions maps file extensions to language names for blame analysis.
// This mirrors the complexity language registry but is defined locally to avoid
// cross-pillar imports (Constitution Principle II).
var SupportedSourceExtensions = map[string]string{
	".go":   "Go",
	".py":   "Python",
	".js":   "JavaScript",
	".jsx":  "JavaScript",
	".ts":   "TypeScript",
	".tsx":  "TypeScript",
	".java": "Java",
	".rb":   "Ruby",
	".rs":   "Rust",
	".c":    "C",
	".cpp":  "C++",
}

// IsSourceFile returns true if the file has a recognized source code extension.
func IsSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := SupportedSourceExtensions[ext]
	return ok
}

// LanguageForFile returns the language name for a source file, or empty string if not recognized.
func LanguageForFile(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	return SupportedSourceExtensions[ext]
}
