# Data Model: AI Layer (Gemini)

## Entities

### AISummary

Structured AI-generated codebase analysis. Cached in Redis with 1h TTL.

| Field | Type | Description |
|-------|------|-------------|
| Repository | Repository | Owner, name, full_name, SHA |
| Description | string | 2-3 paragraph plain-English explanation of the project |
| Architecture | string | Inferred architecture pattern (e.g., MVC, microservices, monolith) |
| TechStack | []string | Detected technologies and frameworks |
| RedFlags | []string | Top 3 potential concerns or risks |
| EntryPoints | []string | Suggested starting points for new contributors |
| GeneratedAt | time.Time | Timestamp of AI generation |
| Model | string | Model used for generation (flash or flash-lite) |
| Cached | bool | Whether this was served from cache |
| CachedAt | *time.Time | When it was cached (nil if not cached) |

**Cache key**: `{owner}/{repo}:{sha}:ai_summary`
**TTL**: 1 hour

---

### ChatMessage

A single message in a conversation.

| Field | Type | Description |
|-------|------|-------------|
| Role | string | "user" or "model" |
| Content | string | Message text |

---

### ChatRequest

Client-sent payload for the chat endpoint.

| Field | Type | Description |
|-------|------|-------------|
| Messages | []ChatMessage | Full conversation history |

**Validation rules**:
- Messages must not be empty
- Maximum 20 messages; oldest truncated if exceeded
- Last message must have role "user"
- Content must not be empty or whitespace-only

---

### HealthScore

Structured AI-generated repository health assessment. Cached with 6h TTL.

| Field | Type | Description |
|-------|------|-------------|
| Repository | Repository | Owner, name, full_name, SHA |
| OverallScore | int | 0-100 composite score |
| Categories | HealthCategories | Five category assessments |
| TopImprovements | []string | Top 3 improvement suggestions |
| GeneratedAt | time.Time | Timestamp of AI generation |
| Model | string | Model used for generation |
| Cached | bool | Whether served from cache |
| CachedAt | *time.Time | When cached |

**Cache key**: `{owner}/{repo}:{sha}:ai_health`
**TTL**: 6 hours

---

### HealthCategories

Container for the five health assessment categories.

| Field | Type | Description |
|-------|------|-------------|
| ReadmeQuality | CategoryScore | README completeness and quality |
| ContributingGuide | CategoryScore | CONTRIBUTING.md presence and quality |
| CodeDocumentation | CategoryScore | Inline docs and comments quality |
| TestCoverageSignals | CategoryScore | Evidence of testing practices |
| ProjectHygiene | CategoryScore | License, CI, gitignore, etc. |

---

### CategoryScore

A single health category assessment.

| Field | Type | Description |
|-------|------|-------------|
| Score | int | 0-100 score for this category |
| Notes | string | Brief explanation of the score |

---

### RepoContext

Assembled context payload passed to the AI provider. Not persisted.

| Field | Type | Description |
|-------|------|-------------|
| Metadata | string | Repo name, description, stars, language, topics, license |
| FileTree | string | Full list of file paths |
| README | string | README.md content (up to 8,000 tokens) |
| Manifests | map[string]string | Package manifest contents (up to 2,000 tokens each) |
| ComplexFiles | []ComplexFileSnippet | Top 5 complex files, first 150 lines each |
| DirSummary | string | Top-level directories with file counts |

---

### ComplexFileSnippet

A snippet from a high-complexity file included in AI context.

| Field | Type | Description |
|-------|------|-------------|
| Path | string | File path |
| Complexity | float64 | Complexity score from analysis |
| Content | string | First 150 lines of the file |

## Relationships

```
AISummary ──> Repository (embedded)
HealthScore ──> Repository (embedded)
HealthScore ──> HealthCategories ──> CategoryScore (5x)
ChatRequest ──> ChatMessage (1..20)
RepoContext ──> ComplexFileSnippet (0..5)
```

## State Transitions

### AI Summary Request
```
[No cache] → Generate (stream) → Cache result → [Cached]
[Cached] → Return cached → [Cached]
[Cache expired] → [No cache]
```

### Health Score Request
```
[No cache] → Generate (sync) → Parse JSON → Cache result → [Cached]
[Parse fail] → Retry with strict prompt → Parse JSON → Cache or Error
[Cached] → Return cached → [Cached]
```

### Chat Request
```
Validate messages → Build context → Prepend system message → Stream response
(No caching)
```
