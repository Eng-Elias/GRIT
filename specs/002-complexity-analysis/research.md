# Research: AST-Based Code Complexity Analysis

**Feature**: 002-complexity-analysis  
**Date**: 2026-03-30

## R1: Tree-Sitter Go Bindings

**Decision**: Use `github.com/smacker/go-tree-sitter` (smacker fork) with bundled grammar packages.

**Rationale**:
- Most mature Go binding for tree-sitter; actively maintained with 1.1k+ stars
- Bundles grammar packages for all 9 target languages as Go sub-packages
- Grammars are statically compiled via cgo — no runtime shared library loading
- API: `sitter.NewParser()` → `parser.SetLanguage(lang.GetLanguage())` → `parser.ParseCtx(ctx, nil, sourceCode)` → traverse `*sitter.Node` tree
- Thread-safe: each goroutine can create its own `Parser` instance
- Query support via S-expression predicates for filtering AST nodes

**Alternatives considered**:
- `github.com/tree-sitter/go-tree-sitter` (official): Newer but requires external grammar shared libraries loaded at runtime. More complex deployment for Docker. Rejected for self-hosting simplicity (Constitution Principle VI).
- Go `go/ast` + language-specific parsers: Only covers Go. Would need separate parsers per language with incompatible APIs. Rejected for maintenance cost across 9 languages.

**Grammar packages** (all under `github.com/smacker/go-tree-sitter/`):

| Language   | Package        | Function node types                                    |
|------------|----------------|--------------------------------------------------------|
| Go         | `golang`       | `function_declaration`, `method_declaration`           |
| TypeScript | `typescript`   | `function_declaration`, `method_definition`, `arrow_function` |
| JavaScript | `javascript`   | `function_declaration`, `method_definition`, `arrow_function` |
| Python     | `python`       | `function_definition`                                  |
| Rust       | `rust`         | `function_item`                                        |
| Java       | `java`         | `method_declaration`, `constructor_declaration`        |
| C          | `c`            | `function_definition`                                  |
| C++        | `cpp`          | `function_definition`                                  |
| Ruby       | `ruby`         | `method`, `singleton_method`                           |

## R2: Cyclomatic Complexity Calculation

**Decision**: Count decision points in AST nodes + 1 per function, per McCabe's original definition.

**Rationale**:
- Industry-standard metric (McCabe, 1976)
- Simple to compute from AST: walk nodes, increment counter for each decision point type
- Directly comparable with tools like `gocyclo`, ESLint complexity rule, Radon (Python)

**Decision points counted per language** (AST node types):

| Category       | Go                          | JS/TS                        | Python                   | Rust                     | Java                     | C/C++                    | Ruby                     |
|----------------|-----------------------------|------------------------------|--------------------------|--------------------------|--------------------------|--------------------------|--------------------------|
| If             | `if_statement`              | `if_statement`               | `if_statement`           | `if_expression`          | `if_statement`           | `if_statement`           | `if`, `elsif`            |
| Else if        | (chained `if_statement`)    | `else` clause with `if`      | `elif_clause`            | `else if` pattern        | `else if` pattern        | `else if` pattern        | `elsif`                  |
| For            | `for_statement`             | `for_statement`, `for_in_statement` | `for_statement`    | `for_expression`         | `for_statement`, `enhanced_for_statement` | `for_statement` | `for`                    |
| While          | (Go has no while)           | `while_statement`            | `while_statement`        | `while_expression`       | `while_statement`        | `while_statement`        | `while`                  |
| Switch case    | `expression_case`           | `switch_case`                | N/A                      | `match_arm`              | `switch_label`           | `case`                   | `when`                   |
| Logical AND    | `&&` in `binary_expression` | `&&` in `binary_expression`  | `and` in `boolean_operator` | `&&` in `binary_expression` | `&&` in `binary_expression` | `&&` in `binary_expression` | `and`                |
| Logical OR     | `\|\|` in `binary_expression` | `\|\|` in `binary_expression` | `or` in `boolean_operator` | `\|\|` in `binary_expression` | `\|\|` in `binary_expression` | `\|\|` in `binary_expression` | `or`               |
| Ternary        | N/A                         | `ternary_expression`         | `conditional_expression` | N/A                      | `ternary_expression`     | `conditional_expression` | `if` modifier            |

**Formula**: `CC(function) = 1 + count(decision_points_in_function)`

## R3: Cognitive Complexity Calculation

**Decision**: Implement SonarSource's Cognitive Complexity specification (Ann Campbell, 2017) adapted to AST traversal.

**Rationale**:
- Cognitive complexity better reflects human perception of code difficulty than cyclomatic
- Accounts for nesting depth, which cyclomatic ignores
- Widely adopted (SonarQube, CodeClimate, ESLint plugin)

**Rules**:

1. **Increment** (+1) for each break in linear flow:
   - `if`, `else if`, `else`
   - `for`, `while`, `do...while`
   - `switch`, `match`
   - `catch`, `except`
   - `goto`, `break` (to label), `continue` (to label)
   - Logical operator sequences: each `&&` or `||` where the operator changes from the previous one in a boolean chain (e.g., `a && b && c` = +1, but `a && b || c` = +2)

2. **Nesting increment** (+nesting_level) added on top of base increment for:
   - `if`, `else if` (not `else`)
   - `for`, `while`, `do...while`
   - `switch`, `match`
   - `catch`, `except`
   - Nested functions / lambdas

3. **Nesting level increases** when entering:
   - `if`, `else if`, `else`
   - `for`, `while`, `do...while`
   - `switch`, `match`
   - `catch`, `except`
   - Nested functions / lambdas

**Formula**: `CogC(function) = sum(increment + nesting_level) for each flow-breaking structure`

**Implementation approach**: Recursive AST walker that tracks current nesting depth. Each qualifying node adds `1 + current_nesting_depth` to the cognitive score.

## R4: Worker Pool Pattern

**Decision**: Fan-out pattern with `runtime.NumCPU()` goroutines, channel-based work distribution, and `sync.WaitGroup` for completion.

**Rationale**:
- Files are independent — no ordering or dependency between them
- Channel-based fan-out is idiomatic Go and simple to implement
- `NumCPU()` bound prevents over-scheduling; tree-sitter parsing is CPU-bound
- Each goroutine creates its own `sitter.Parser` instance (no shared state)

**Pattern**:
```
files channel → N worker goroutines → results channel → collector goroutine
```

- Input: `[]models.FileStats` from core analysis (filtered to supported languages)
- Workers: each reads file content from git repo, parses with tree-sitter, computes metrics
- Output: `[]models.FileComplexity` collected by single goroutine
- Errors: per-file errors logged and skipped, never abort pool

**Alternatives considered**:
- `errgroup` with semaphore: Slightly cleaner error handling but cancels all goroutines on first error. Rejected because we want to continue on parse errors.
- `sync.Pool` for parsers: Unnecessary — each goroutine lives for the entire analysis duration and uses its own parser.

## R5: NATS Subject and Auto-Trigger

**Decision**: Use NATS subject `grit.jobs.complexity` on the existing `GRIT` stream. Core worker publishes complexity job after successful core completion.

**Rationale**:
- The `GRIT` stream already uses `grit.jobs.>` wildcard for subjects
- Separate subject allows independent consumer groups for each pillar
- Auto-trigger from core worker: after caching core result, publish a complexity payload with the same owner/repo/sha/token
- Separate consumer/subscriber: `PullSubscribe("grit.jobs.complexity", "grit-complexity-worker")`

**Complexity job payload**: Same structure as core `JobPayload` — `job_id`, `owner`, `repo`, `sha`, `token`. The `sha` ensures the complexity analysis uses the same commit as the core analysis.

## R6: Cache Key Strategy

**Decision**: Separate Redis key `{owner}/{repo}:{sha}:complexity` with 24h TTL (per Constitution Principle IV).

**Rationale**:
- Separate from core key (`{owner}/{repo}:{sha}:core`) per pillar isolation
- 24h TTL matches constitution's complexity pillar TTL
- `findComplexity` scan pattern: `{owner}/{repo}:*:complexity` for SHA-unknown lookups
- `DeleteComplexity` follows same SCAN-based pattern as `DeleteAnalysis`

## R7: Complexity Summary Embedding

**Decision**: When the analysis handler returns a cached core result, also look up the complexity cache. If found, attach a `complexity_summary` field to the response.

**Rationale**:
- Avoids requiring a separate API call for basic complexity info
- Summary is lightweight: mean complexity, total function count, hot file count, status
- If complexity cache miss: embed `complexity_summary: { "status": "pending" }` to indicate it's not yet available
- The full complexity data remains at the dedicated endpoint

## R8: Tree-Sitter cgo and Docker

**Decision**: Ensure Docker build includes C compiler (gcc) for cgo compilation of tree-sitter.

**Rationale**:
- `go-tree-sitter` uses cgo to link the tree-sitter C library
- The existing Dockerfile likely uses `golang:1.22-alpine` which includes gcc via `build-base`
- If not present, add `RUN apk add --no-cache build-base` to the Dockerfile build stage
- No runtime C dependencies — tree-sitter is statically linked into the Go binary
