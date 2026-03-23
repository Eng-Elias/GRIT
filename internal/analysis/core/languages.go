package core

type LanguageDef struct {
	Name               string
	LineCommentPrefixes []string
	BlockCommentStart  string
	BlockCommentEnd    string
}

var extensionMap = map[string]LanguageDef{
	".go":          {Name: "Go", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".py":          {Name: "Python", LineCommentPrefixes: []string{"#"}, BlockCommentStart: `"""`, BlockCommentEnd: `"""`},
	".js":          {Name: "JavaScript", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".jsx":         {Name: "JavaScript", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".ts":          {Name: "TypeScript", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".tsx":         {Name: "TypeScript", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".java":        {Name: "Java", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".kt":          {Name: "Kotlin", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".kts":         {Name: "Kotlin", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".c":           {Name: "C", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".h":           {Name: "C", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".cpp":         {Name: "C++", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".cc":          {Name: "C++", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".cxx":         {Name: "C++", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".hpp":         {Name: "C++", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".cs":          {Name: "C#", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".rs":          {Name: "Rust", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".swift":       {Name: "Swift", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".rb":          {Name: "Ruby", LineCommentPrefixes: []string{"#"}, BlockCommentStart: "=begin", BlockCommentEnd: "=end"},
	".php":         {Name: "PHP", LineCommentPrefixes: []string{"//", "#"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".pl":          {Name: "Perl", LineCommentPrefixes: []string{"#"}, BlockCommentStart: "=pod", BlockCommentEnd: "=cut"},
	".pm":          {Name: "Perl", LineCommentPrefixes: []string{"#"}, BlockCommentStart: "=pod", BlockCommentEnd: "=cut"},
	".lua":         {Name: "Lua", LineCommentPrefixes: []string{"--"}, BlockCommentStart: "--[[", BlockCommentEnd: "]]"},
	".r":           {Name: "R", LineCommentPrefixes: []string{"#"}},
	".R":           {Name: "R", LineCommentPrefixes: []string{"#"}},
	".scala":       {Name: "Scala", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".hs":          {Name: "Haskell", LineCommentPrefixes: []string{"--"}, BlockCommentStart: "{-", BlockCommentEnd: "-}"},
	".ex":          {Name: "Elixir", LineCommentPrefixes: []string{"#"}},
	".exs":         {Name: "Elixir", LineCommentPrefixes: []string{"#"}},
	".erl":         {Name: "Erlang", LineCommentPrefixes: []string{"%"}},
	".hrl":         {Name: "Erlang", LineCommentPrefixes: []string{"%"}},
	".clj":         {Name: "Clojure", LineCommentPrefixes: []string{";"}},
	".cljs":        {Name: "Clojure", LineCommentPrefixes: []string{";"}},
	".dart":        {Name: "Dart", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".zig":         {Name: "Zig", LineCommentPrefixes: []string{"//"}},
	".nim":         {Name: "Nim", LineCommentPrefixes: []string{"#"}, BlockCommentStart: "#[", BlockCommentEnd: "]#"},
	".ml":          {Name: "OCaml", LineCommentPrefixes: nil, BlockCommentStart: "(*", BlockCommentEnd: "*)"},
	".mli":         {Name: "OCaml", LineCommentPrefixes: nil, BlockCommentStart: "(*", BlockCommentEnd: "*)"},
	".fs":          {Name: "F#", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "(*", BlockCommentEnd: "*)"},
	".fsx":         {Name: "F#", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "(*", BlockCommentEnd: "*)"},
	".cob":         {Name: "COBOL", LineCommentPrefixes: []string{"*"}},
	".cbl":         {Name: "COBOL", LineCommentPrefixes: []string{"*"}},
	".f":           {Name: "Fortran", LineCommentPrefixes: []string{"!"}},
	".f90":         {Name: "Fortran", LineCommentPrefixes: []string{"!"}},
	".f95":         {Name: "Fortran", LineCommentPrefixes: []string{"!"}},
	".asm":         {Name: "Assembly", LineCommentPrefixes: []string{";"}},
	".s":           {Name: "Assembly", LineCommentPrefixes: []string{";", "#"}},
	".sh":          {Name: "Shell", LineCommentPrefixes: []string{"#"}},
	".bash":        {Name: "Shell", LineCommentPrefixes: []string{"#"}},
	".zsh":         {Name: "Shell", LineCommentPrefixes: []string{"#"}},
	".ps1":         {Name: "PowerShell", LineCommentPrefixes: []string{"#"}, BlockCommentStart: "<#", BlockCommentEnd: "#>"},
	".sql":         {Name: "SQL", LineCommentPrefixes: []string{"--"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".html":        {Name: "HTML", LineCommentPrefixes: nil, BlockCommentStart: "<!--", BlockCommentEnd: "-->"},
	".htm":         {Name: "HTML", LineCommentPrefixes: nil, BlockCommentStart: "<!--", BlockCommentEnd: "-->"},
	".css":         {Name: "CSS", LineCommentPrefixes: nil, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".scss":        {Name: "SCSS", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".less":        {Name: "LESS", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".xml":         {Name: "XML", LineCommentPrefixes: nil, BlockCommentStart: "<!--", BlockCommentEnd: "-->"},
	".json":        {Name: "JSON", LineCommentPrefixes: nil},
	".yaml":        {Name: "YAML", LineCommentPrefixes: []string{"#"}},
	".yml":         {Name: "YAML", LineCommentPrefixes: []string{"#"}},
	".toml":        {Name: "TOML", LineCommentPrefixes: []string{"#"}},
	".md":          {Name: "Markdown", LineCommentPrefixes: nil},
	".markdown":    {Name: "Markdown", LineCommentPrefixes: nil},
	".dockerfile":  {Name: "Dockerfile", LineCommentPrefixes: []string{"#"}},
	".makefile":    {Name: "Makefile", LineCommentPrefixes: []string{"#"}},
	".cmake":       {Name: "CMake", LineCommentPrefixes: []string{"#"}},
	".gradle":      {Name: "Gradle", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".tf":          {Name: "Terraform", LineCommentPrefixes: []string{"#", "//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".hcl":         {Name: "HCL", LineCommentPrefixes: []string{"#", "//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".proto":       {Name: "Protobuf", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".graphql":     {Name: "GraphQL", LineCommentPrefixes: []string{"#"}},
	".gql":         {Name: "GraphQL", LineCommentPrefixes: []string{"#"}},
	".svelte":      {Name: "Svelte", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "<!--", BlockCommentEnd: "-->"},
	".vue":         {Name: "Vue", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "<!--", BlockCommentEnd: "-->"},
	".v":           {Name: "V", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".jl":          {Name: "Julia", LineCommentPrefixes: []string{"#"}, BlockCommentStart: "#=", BlockCommentEnd: "=#"},
	".m":           {Name: "Objective-C", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".mm":          {Name: "Objective-C++", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".groovy":      {Name: "Groovy", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
	".pas":         {Name: "Pascal", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "{", BlockCommentEnd: "}"},
	".pp":          {Name: "Pascal", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "{", BlockCommentEnd: "}"},
	".ada":         {Name: "Ada", LineCommentPrefixes: []string{"--"}},
	".adb":         {Name: "Ada", LineCommentPrefixes: []string{"--"}},
	".ads":         {Name: "Ada", LineCommentPrefixes: []string{"--"}},
	".vhdl":        {Name: "VHDL", LineCommentPrefixes: []string{"--"}},
	".vhd":         {Name: "VHDL", LineCommentPrefixes: []string{"--"}},
	".sv":          {Name: "SystemVerilog", LineCommentPrefixes: []string{"//"}, BlockCommentStart: "/*", BlockCommentEnd: "*/"},
}

var filenameMap = map[string]LanguageDef{
	"Dockerfile": {Name: "Dockerfile", LineCommentPrefixes: []string{"#"}},
	"Makefile":   {Name: "Makefile", LineCommentPrefixes: []string{"#"}},
	"CMakeLists.txt": {Name: "CMake", LineCommentPrefixes: []string{"#"}},
	"Rakefile":   {Name: "Ruby", LineCommentPrefixes: []string{"#"}},
	"Gemfile":    {Name: "Ruby", LineCommentPrefixes: []string{"#"}},
}

func LookupLanguage(ext string, filename string) LanguageDef {
	if def, ok := filenameMap[filename]; ok {
		return def
	}
	if def, ok := extensionMap[ext]; ok {
		return def
	}
	return LanguageDef{Name: "Other"}
}
