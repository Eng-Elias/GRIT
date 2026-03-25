# Feature Specification: AST-Based Code Complexity Analysis

**Feature Branch**: `002-complexity-analysis`  
**Created**: 2026-03-26  
**Status**: Draft  
**Input**: User description: "Add AST-based code complexity analysis to GRIT. ghloc-web counts lines; GRIT tells you which lines actually matter by measuring structural complexity."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Per-File Complexity Analysis (Priority: P1)

A developer requests complexity analysis for a GitHub repository. The system
parses each supported-language source file using AST-based parsing, computes
cyclomatic complexity, cognitive complexity, function count, and per-function
complexity metrics, and returns a complete per-file complexity breakdown.

**Why this priority**: Per-file complexity metrics are the core value
proposition — they answer "which files are the most structurally complex?"
and form the foundation for all aggregated views and hot-file detection.

**Independent Test**: Can be fully tested by calling
`GET /api/:owner/:repo/complexity` for a public repository and verifying
the response contains per-file complexity metrics with cyclomatic, cognitive,
function count, average function complexity, and max function complexity for
each supported-language file.

**Acceptance Scenarios**:

1. **Given** a repository containing Go, TypeScript, and Python files,
   **When** the user requests `GET /api/:owner/:repo/complexity`,
   **Then** the response includes per-file complexity metrics for each
   supported-language file, with cyclomatic complexity, cognitive complexity,
   function count, average function complexity, and max function complexity.

2. **Given** a repository containing files in an unsupported language
   (e.g., Haskell), **When** the system encounters those files during
   analysis, **Then** it skips them without error and excludes them from
   the complexity results.

3. **Given** a repository with a cached complexity analysis at the same
   commit SHA, **When** the user requests complexity data, **Then** the
   system returns the cached result immediately with `cached: true`.

---

### User Story 2 - Hot Files and Repository-Level Aggregation (Priority: P1)

A developer wants to quickly identify the most complex parts of a codebase.
The system ranks files by complexity density (cyclomatic complexity / lines
of code) and provides repository-level aggregate statistics including mean,
median, p90 complexity, total function count, and a complexity distribution
histogram.

**Why this priority**: Aggregated metrics and hot-file ranking are the
primary decision-making outputs — they tell developers where to focus
refactoring effort. Without aggregation, raw per-file data requires manual
analysis.

**Independent Test**: Can be tested by requesting complexity for a
repository with known complex files and verifying the `hot_files` list
is ordered by complexity density, and the repository-level aggregates
(mean, median, p90, histogram) are present and mathematically consistent.

**Acceptance Scenarios**:

1. **Given** a repository with multiple source files of varying complexity,
   **When** the user requests `GET /api/:owner/:repo/complexity`,
   **Then** the response includes a `hot_files` list of up to 20 files
   ranked by descending complexity density (cyclomatic / LOC).

2. **Given** the same response, **Then** repository-level fields include
   `mean_complexity`, `median_complexity`, `p90_complexity`, and
   `total_function_count` computed across all analyzed files.

3. **Given** the same response, **Then** a `distribution` object contains
   counts for four buckets: Low (average function complexity 1–5),
   Medium (6–10), High (11–20), and Critical (21+).

---

### User Story 3 - Async Complexity Job Triggered After Core Analysis (Priority: P2)

When core analysis completes for a repository, the system automatically
enqueues a complexity analysis job. The complexity results become available
alongside the core results once processing finishes.

**Why this priority**: Auto-triggering ensures users get complexity data
without a separate request, but it depends on the core analysis pipeline
already functioning correctly.

**Independent Test**: Can be tested by requesting core analysis for a
repository, waiting for completion, then verifying that complexity data
is available at `GET /api/:owner/:repo/complexity` and a summary is
embedded in the main `GET /api/:owner/:repo` response.

**Acceptance Scenarios**:

1. **Given** a core analysis job completes for a repository, **When** the
   worker finishes the core pillar, **Then** the system automatically
   publishes a complexity job to the NATS job queue.

2. **Given** a completed complexity analysis, **When** the user requests
   `GET /api/:owner/:repo`, **Then** the response includes a
   `complexity_summary` field with key aggregate metrics (mean complexity,
   total functions, hot file count).

3. **Given** a complexity job is in progress, **When** the user requests
   `GET /api/:owner/:repo/complexity`, **Then** the system returns HTTP
   202 with a job status indicating the complexity analysis is running.

---

### User Story 4 - Parallel File Parsing (Priority: P2)

The system parses files in parallel using a bounded worker pool to ensure
complexity analysis completes within acceptable time limits even for large
repositories.

**Why this priority**: Parallel parsing is essential for performance but
is an optimization of the core parsing logic from User Story 1.

**Independent Test**: Can be tested by analyzing a repository with many
files and verifying that the analysis completes within the expected time
budget and that results are identical to a sequential run.

**Acceptance Scenarios**:

1. **Given** a repository with 1,000+ supported-language files, **When**
   complexity analysis runs, **Then** files are parsed concurrently using
   a worker pool bounded by the number of available CPU cores.

2. **Given** a parsing error in one file, **When** the worker pool
   encounters the error, **Then** it logs the error and continues
   processing remaining files without aborting the entire analysis.

---

### Edge Cases

- What happens when a repository has zero supported-language files?
  The system MUST return a valid complexity result with empty file list,
  zero aggregate metrics, and an empty hot_files list.
- What happens when a file has zero functions?
  The system MUST report cyclomatic complexity of 0, function count of 0,
  and average/max function complexity of 0.
- What happens when a file has syntax errors that prevent AST parsing?
  The system MUST skip the file, log a warning with the file path and
  error, and continue processing remaining files.
- What happens when the repository is extremely large (>50k files)?
  The system MUST respect the 5-minute analysis timeout from the core
  pillar and return partial complexity results for files that were
  analyzed before the timeout.
- What happens when complexity analysis is requested but core analysis
  has not yet completed?
  The system MUST return HTTP 404 with a message indicating core
  analysis must complete first.
- What happens when the same repository is re-analyzed at a new commit?
  The old complexity cache entry is superseded by the new one; the
  cache key includes the commit SHA.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST parse source files using AST-based parsing to
  extract structural complexity metrics (not regex or line-based heuristics).
- **FR-002**: System MUST support AST parsing for the following languages:
  Go, TypeScript, JavaScript, Python, Rust, Java, C, C++, and Ruby.
- **FR-003**: System MUST gracefully skip files in unsupported languages
  without producing errors in the response.
- **FR-004**: System MUST compute the following metrics per file:
  cyclomatic complexity, cognitive complexity, function count, average
  function complexity (total cyclomatic / function count), and max
  single-function complexity.
- **FR-005**: Cyclomatic complexity MUST be calculated by counting
  decision points (if, else if, for, while, switch cases, logical AND,
  logical OR, ternary) plus 1 per function.
- **FR-006**: Cognitive complexity MUST be calculated using a weighted
  nesting score where each nesting level multiplies the base complexity
  of decision points within it.
- **FR-007**: System MUST produce a `hot_files` ranking of up to 20 files
  ordered by descending complexity density (cyclomatic complexity / lines
  of code).
- **FR-008**: System MUST compute repository-level aggregate statistics:
  mean complexity, median complexity, 90th-percentile complexity, and
  total function count across all analyzed files.
- **FR-009**: System MUST produce a complexity distribution histogram
  with four buckets based on average function complexity: Low (1–5),
  Medium (6–10), High (11–20), Critical (21+).
- **FR-010**: System MUST parse files in parallel using a worker pool
  bounded by the number of available CPU cores.
- **FR-011**: System MUST publish complexity jobs to a dedicated NATS
  subject (`grit.jobs.complexity`) and process them asynchronously.
- **FR-012**: System MUST automatically trigger a complexity analysis job
  when core analysis completes for a repository.
- **FR-013**: System MUST cache complexity results in Redis with key
  format `{owner}/{repo}:{sha}:complexity` and a 24-hour TTL.
- **FR-014**: System MUST expose a dedicated endpoint at
  `GET /api/:owner/:repo/complexity` returning the full complexity data.
- **FR-015**: System MUST embed a `complexity_summary` in the main
  `GET /api/:owner/:repo` response once complexity analysis is available.
- **FR-016**: System MUST handle AST parse errors per file gracefully:
  log the error, skip the file, and continue with remaining files.
- **FR-017**: System MUST respect the existing 5-minute analysis timeout
  and return partial results if the timeout is reached.
- **FR-018**: System MUST reuse the cloned repository from the core
  analysis pillar (retained in the temporary directory with 1-hour TTL)
  rather than cloning again.

### Key Entities

- **FileComplexity**: Per-file complexity breakdown. Key attributes: file
  path, language, cyclomatic complexity, cognitive complexity, function
  count, average function complexity, max function complexity, lines of
  code, complexity density (cyclomatic / LOC).
- **FunctionComplexity**: Per-function complexity detail. Key attributes:
  function name, start line, end line, cyclomatic complexity, cognitive
  complexity.
- **ComplexityResult**: The complete complexity analysis output for a
  repository at a specific commit. Key attributes: file complexities,
  hot files list, repository-level aggregates (mean, median, p90,
  total functions), distribution histogram, analyzed_at timestamp,
  cache status.
- **ComplexityDistribution**: Histogram of file complexity. Key
  attributes: low count (1–5), medium count (6–10), high count (11–20),
  critical count (21+).
- **ComplexitySummary**: Abbreviated complexity data embedded in the main
  analysis response. Key attributes: mean complexity, total function
  count, hot file count, analysis status (pending/complete).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Complexity analysis of a repository with up to 5,000
  supported-language files completes within 3 minutes end-to-end.
- **SC-002**: Cyclomatic complexity values for Go files match
  `gocyclo`-equivalent calculations within ±1 for a set of 5 reference
  repositories.
- **SC-003**: The hot_files ranking correctly identifies the top 20 most
  complex files by density for a known reference repository.
- **SC-004**: Repository-level aggregates (mean, median, p90) are
  mathematically correct — verifiable by independent calculation from
  the per-file data in the same response.
- **SC-005**: Files in unsupported languages are excluded from results
  without errors — verified by analyzing a polyglot repository.
- **SC-006**: AST parse errors in individual files do not abort the
  entire analysis — verified by including a malformed file in a test
  repository.
- **SC-007**: The complexity endpoint returns valid JSON matching the
  documented schema for every request (success and error cases).
- **SC-008**: Cached complexity results are returned within 50ms (p95).
- **SC-009**: Complexity analysis automatically triggers after core
  analysis completes, without requiring a separate user request.
- **SC-010**: The worker pool processes files concurrently — verified by
  observing reduced wall-clock time compared to sequential processing
  on a multi-core system.
