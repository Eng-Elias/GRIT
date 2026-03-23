package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupLanguage_Extensions(t *testing.T) {
	tests := []struct {
		ext      string
		filename string
		expected string
	}{
		{".go", "main.go", "Go"},
		{".py", "app.py", "Python"},
		{".js", "index.js", "JavaScript"},
		{".jsx", "App.jsx", "JavaScript"},
		{".ts", "index.ts", "TypeScript"},
		{".tsx", "App.tsx", "TypeScript"},
		{".java", "Main.java", "Java"},
		{".kt", "Main.kt", "Kotlin"},
		{".c", "main.c", "C"},
		{".h", "header.h", "C"},
		{".cpp", "main.cpp", "C++"},
		{".cc", "main.cc", "C++"},
		{".hpp", "header.hpp", "C++"},
		{".cs", "Program.cs", "C#"},
		{".rs", "main.rs", "Rust"},
		{".swift", "main.swift", "Swift"},
		{".rb", "app.rb", "Ruby"},
		{".php", "index.php", "PHP"},
		{".pl", "script.pl", "Perl"},
		{".lua", "init.lua", "Lua"},
		{".r", "analysis.r", "R"},
		{".scala", "Main.scala", "Scala"},
		{".hs", "Main.hs", "Haskell"},
		{".ex", "app.ex", "Elixir"},
		{".erl", "mod.erl", "Erlang"},
		{".clj", "core.clj", "Clojure"},
		{".dart", "main.dart", "Dart"},
		{".zig", "main.zig", "Zig"},
		{".nim", "main.nim", "Nim"},
		{".ml", "main.ml", "OCaml"},
		{".fs", "Program.fs", "F#"},
		{".sh", "build.sh", "Shell"},
		{".bash", "run.bash", "Shell"},
		{".ps1", "script.ps1", "PowerShell"},
		{".sql", "query.sql", "SQL"},
		{".html", "index.html", "HTML"},
		{".css", "style.css", "CSS"},
		{".scss", "style.scss", "SCSS"},
		{".less", "style.less", "LESS"},
		{".xml", "config.xml", "XML"},
		{".json", "package.json", "JSON"},
		{".yaml", "config.yaml", "YAML"},
		{".yml", "docker.yml", "YAML"},
		{".toml", "config.toml", "TOML"},
		{".md", "README.md", "Markdown"},
		{".tf", "main.tf", "Terraform"},
		{".proto", "service.proto", "Protobuf"},
		{".graphql", "schema.graphql", "GraphQL"},
		{".svelte", "App.svelte", "Svelte"},
		{".vue", "App.vue", "Vue"},
	}

	for _, tt := range tests {
		t.Run(tt.expected+"_"+tt.ext, func(t *testing.T) {
			lang := LookupLanguage(tt.ext, tt.filename)
			assert.Equal(t, tt.expected, lang.Name)
		})
	}
}

func TestLookupLanguage_Filenames(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"Dockerfile", "Dockerfile"},
		{"Makefile", "Makefile"},
		{"CMakeLists.txt", "CMake"},
		{"Rakefile", "Ruby"},
		{"Gemfile", "Ruby"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			lang := LookupLanguage("", tt.filename)
			assert.Equal(t, tt.expected, lang.Name)
		})
	}
}

func TestLookupLanguage_Unknown(t *testing.T) {
	lang := LookupLanguage(".xyz", "unknown.xyz")
	assert.Equal(t, "Other", lang.Name)
}

func TestLookupLanguage_AtLeast40Languages(t *testing.T) {
	seen := make(map[string]bool)
	for _, def := range extensionMap {
		seen[def.Name] = true
	}
	assert.GreaterOrEqual(t, len(seen), 40, "must support at least 40 languages")
}
