package complexity

import (
	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// LanguageConfig holds the tree-sitter language and AST node type names
// needed to extract functions and compute complexity for a specific language.
type LanguageConfig struct {
	Name               string
	Language           *gotreesitter.Language
	FunctionNodeTypes  []string
	DecisionPointTypes []string
	// NestingTypes are node types that increase cognitive nesting depth.
	NestingTypes []string
	// LogicalOperators are binary expression operators counted for cyclomatic.
	LogicalOperators []string
}

var languageRegistry = map[string]*LanguageConfig{
	".go": {
		Name:     "Go",
		Language: grammars.GoLanguage(),
		FunctionNodeTypes: []string{
			"function_declaration",
			"method_declaration",
			"func_literal",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"for_statement",
			"expression_case",
			"type_case",
		},
		NestingTypes: []string{
			"if_statement",
			"for_statement",
			"expression_switch_statement",
			"type_switch_statement",
			"func_literal",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".ts": {
		Name:     "TypeScript",
		Language: grammars.TypescriptLanguage(),
		FunctionNodeTypes: []string{
			"function_declaration",
			"method_definition",
			"arrow_function",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"for_statement",
			"for_in_statement",
			"while_statement",
			"do_statement",
			"switch_case",
			"ternary_expression",
			"catch_clause",
		},
		NestingTypes: []string{
			"if_statement",
			"for_statement",
			"for_in_statement",
			"while_statement",
			"do_statement",
			"switch_statement",
			"catch_clause",
			"arrow_function",
			"ternary_expression",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".js": {
		Name:     "JavaScript",
		Language: grammars.JavascriptLanguage(),
		FunctionNodeTypes: []string{
			"function_declaration",
			"method_definition",
			"arrow_function",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"for_statement",
			"for_in_statement",
			"while_statement",
			"do_statement",
			"switch_case",
			"ternary_expression",
			"catch_clause",
		},
		NestingTypes: []string{
			"if_statement",
			"for_statement",
			"for_in_statement",
			"while_statement",
			"do_statement",
			"switch_statement",
			"catch_clause",
			"arrow_function",
			"ternary_expression",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".py": {
		Name:     "Python",
		Language: grammars.PythonLanguage(),
		FunctionNodeTypes: []string{
			"function_definition",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"elif_clause",
			"for_statement",
			"while_statement",
			"except_clause",
		},
		NestingTypes: []string{
			"if_statement",
			"elif_clause",
			"else_clause",
			"for_statement",
			"while_statement",
			"except_clause",
			"function_definition",
		},
		LogicalOperators: []string{"and", "or"},
	},
	".rs": {
		Name:     "Rust",
		Language: grammars.RustLanguage(),
		FunctionNodeTypes: []string{
			"function_item",
		},
		DecisionPointTypes: []string{
			"if_expression",
			"for_expression",
			"while_expression",
			"match_arm",
		},
		NestingTypes: []string{
			"if_expression",
			"for_expression",
			"while_expression",
			"match_expression",
			"closure_expression",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".java": {
		Name:     "Java",
		Language: grammars.JavaLanguage(),
		FunctionNodeTypes: []string{
			"method_declaration",
			"constructor_declaration",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"for_statement",
			"enhanced_for_statement",
			"while_statement",
			"do_statement",
			"switch_label",
			"ternary_expression",
			"catch_clause",
		},
		NestingTypes: []string{
			"if_statement",
			"for_statement",
			"enhanced_for_statement",
			"while_statement",
			"do_statement",
			"switch_expression",
			"catch_clause",
			"lambda_expression",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".c": {
		Name:     "C",
		Language: grammars.CLanguage(),
		FunctionNodeTypes: []string{
			"function_definition",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"for_statement",
			"while_statement",
			"do_statement",
			"case_statement",
			"conditional_expression",
		},
		NestingTypes: []string{
			"if_statement",
			"for_statement",
			"while_statement",
			"do_statement",
			"switch_statement",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".cpp": {
		Name:     "C++",
		Language: grammars.CppLanguage(),
		FunctionNodeTypes: []string{
			"function_definition",
		},
		DecisionPointTypes: []string{
			"if_statement",
			"for_statement",
			"for_range_loop",
			"while_statement",
			"do_statement",
			"case_statement",
			"conditional_expression",
			"catch_clause",
		},
		NestingTypes: []string{
			"if_statement",
			"for_statement",
			"for_range_loop",
			"while_statement",
			"do_statement",
			"switch_statement",
			"catch_clause",
			"lambda_expression",
		},
		LogicalOperators: []string{"&&", "||"},
	},
	".cc":  nil, // alias, resolved in init()
	".cxx": nil, // alias, resolved in init()
	".h":   nil, // alias, resolved in init()
	".hpp": nil, // alias, resolved in init()
	".rb": {
		Name:     "Ruby",
		Language: grammars.RubyLanguage(),
		FunctionNodeTypes: []string{
			"method",
			"singleton_method",
		},
		DecisionPointTypes: []string{
			"if",
			"elsif",
			"unless",
			"for",
			"while",
			"until",
			"when",
			"rescue",
		},
		NestingTypes: []string{
			"if",
			"elsif",
			"else",
			"unless",
			"for",
			"while",
			"until",
			"case",
			"rescue",
			"lambda",
			"do_block",
		},
		LogicalOperators: []string{"and", "or", "&&", "||"},
	},
}

func init() {
	// Set up file extension aliases that share a language config.
	languageRegistry[".cc"] = languageRegistry[".cpp"]
	languageRegistry[".cxx"] = languageRegistry[".cpp"]
	languageRegistry[".h"] = languageRegistry[".c"]
	languageRegistry[".hpp"] = languageRegistry[".cpp"]
	languageRegistry[".jsx"] = languageRegistry[".js"]
	languageRegistry[".tsx"] = languageRegistry[".ts"]
	languageRegistry[".mjs"] = languageRegistry[".js"]
	languageRegistry[".cjs"] = languageRegistry[".js"]
	languageRegistry[".mts"] = languageRegistry[".ts"]
	languageRegistry[".cts"] = languageRegistry[".ts"]
}

// GetLanguageConfig returns the LanguageConfig for the given file extension,
// or nil if the extension is not supported.
func GetLanguageConfig(ext string) *LanguageConfig {
	return languageRegistry[ext]
}

// SupportedExtensions returns all file extensions that have a registered language.
func SupportedExtensions() []string {
	exts := make([]string, 0, len(languageRegistry))
	for ext, cfg := range languageRegistry {
		if cfg != nil {
			exts = append(exts, ext)
		}
	}
	return exts
}
