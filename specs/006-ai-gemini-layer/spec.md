# Feature Specification: AI Layer (Gemini)

**Feature Branch**: `006-ai-gemini-layer`
**Created**: 2026-04-15
**Status**: Draft
**Input**: User description: "Add AI-powered intelligence to GRIT using Google Gemini (free tier)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI Codebase Summary (Priority: P1)

A user visits a repository's analysis page and requests an AI-generated summary. The system streams a structured analysis of the codebase in real-time via Server-Sent Events, explaining what the project does, its architecture, tech stack, potential red flags, and suggested entry points for new contributors. The result is cached for subsequent visitors.

**Why this priority**: This is the flagship differentiator from existing tools like ghloc-web. A natural-language explanation of a codebase delivers immediate value and showcases the AI integration end-to-end, including context construction, streaming, caching, and error handling.

**Independent Test**: Can be fully tested by sending a POST request to the summary endpoint for a repository with a completed core analysis and verifying that structured SSE chunks arrive, the final result is cached, and subsequent requests return the cached version.

**Acceptance Scenarios**:

1. **Given** a repository with completed core analysis and a configured AI API key, **When** the user requests an AI summary, **Then** the system streams a structured response containing project description, architecture pattern, tech stack, red flags, and contributor entry points.
2. **Given** a cached AI summary exists for the current commit SHA, **When** the user requests the summary again, **Then** the system returns the cached result immediately with an `X-Cache: HIT` header.
3. **Given** the AI API key is not configured, **When** the user requests an AI summary, **Then** the system returns a 503 error with a clear message.
4. **Given** the core analysis has not yet completed, **When** the user requests an AI summary, **Then** the system returns a 409 error indicating analysis is pending.
5. **Given** the AI provider rate-limits the request after all retry attempts, **When** the user requests a summary, **Then** the system returns a structured error with a `retry_after` value and never exposes raw provider error messages.

---

### User Story 2 - Ask This Repo Chat (Priority: P2)

A user opens a conversational interface to ask free-form questions about a repository's codebase. The system prepends repository context to the conversation and streams AI responses in real-time. The conversation is stateless on the server; the client sends the full message history with each request.

**Why this priority**: Interactive Q&A is the most engaging AI feature and extends the summary into a dynamic exploration tool. It builds on the same context construction and streaming infrastructure as US1.

**Independent Test**: Can be fully tested by sending a POST request with a message history, verifying SSE streaming, confirming context is prepended, and validating rate limiting (10 requests per IP per minute).

**Acceptance Scenarios**:

1. **Given** a repository with completed core analysis and valid message history, **When** the user sends a chat message, **Then** the system streams an AI response grounded in the repository context.
2. **Given** a conversation exceeding 20 turns, **When** the user sends a new message, **Then** the system truncates the oldest messages to stay within 20 turns before sending to the AI provider.
3. **Given** a single IP has sent 10 chat requests within one minute, **When** the same IP sends another request, **Then** the system returns a 429 rate-limited response.
4. **Given** the AI API key is not configured, **When** the user sends a chat message, **Then** the system returns a 503 error.
5. **Given** the core analysis has not yet completed, **When** the user sends a chat message, **Then** the system returns a 409 error.

---

### User Story 3 - AI Health Score (Priority: P3)

A user requests an AI-driven health assessment of a repository. The system sends repository context to the AI provider with a structured JSON output requirement and returns a health score across five categories: README quality, contributing guide, code documentation, test coverage signals, and project hygiene. The response includes an overall score (0-100), per-category scores with notes, and top improvement suggestions.

**Why this priority**: The health score provides actionable, quantified feedback. It is non-streaming and simpler to implement, but depends on the same context construction and client infrastructure.

**Independent Test**: Can be fully tested by sending a GET request to the health endpoint, verifying the response is valid JSON matching the expected schema, and confirming caching with a 6-hour TTL.

**Acceptance Scenarios**:

1. **Given** a repository with completed core analysis, **When** the user requests a health score, **Then** the system returns a JSON object with an overall score (0-100), five category scores with notes, and a list of top improvements.
2. **Given** a cached health score exists for the current commit SHA, **When** the user requests the score again, **Then** the system returns the cached result with `X-Cache: HIT`.
3. **Given** the AI provider returns malformed JSON on the first attempt, **When** the system retries with a stricter prompt, **Then** it returns valid structured JSON.
4. **Given** JSON parsing fails on both attempts, **When** the system has exhausted retries, **Then** it returns a structured error to the client.

---

### User Story 4 - Shared Context Construction (Priority: P1)

Before any AI request, the system assembles a rich context payload from cached analysis data. This context includes repository metadata, file tree, README content, package manifest contents, top complex files, and directory structure. The total context is kept under 800,000 tokens with intelligent truncation prioritizing README and manifests.

**Why this priority**: Context construction is the foundation all three AI features depend on. Without well-structured context, AI responses are generic and low-value.

**Independent Test**: Can be tested by constructing context for a known repository and verifying the output includes all expected sections, respects token limits, and truncates appropriately.

**Acceptance Scenarios**:

1. **Given** a repository with core and complexity analysis cached, **When** context is constructed, **Then** it includes metadata, file tree, README (up to 8,000 tokens), manifest contents (up to 2,000 tokens each), top 5 complex files (first 150 lines each), and directory structure summary.
2. **Given** a repository with a very large file tree, **When** context is constructed, **Then** the total stays under 800,000 tokens with file content truncated before README or manifests.
3. **Given** a repository with no README, **When** context is constructed, **Then** the README section is omitted gracefully and other sections fill the available budget.

---

### Edge Cases

- What happens when the AI provider is completely unreachable (network failure)? The system returns a 503 after exhausting retries with exponential backoff.
- What happens when the cached core analysis exists but the complexity analysis does not? Context construction proceeds without the "top complex files" section.
- What happens when a repository has no recognized package manifests? The manifests section is omitted from the context.
- What happens when the AI returns a valid response but with unexpected structure? The system validates required fields and returns an error if critical fields are missing.
- What happens when multiple concurrent requests arrive for the same uncached summary? The first request triggers AI generation; concurrent requests receive a 409 or wait for cache population.
- What happens when the chat message body is empty or contains only whitespace? The system returns a 400 bad request.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a Gemini client using the `google.golang.org/genai` SDK with `gemini-2.5-flash` as the primary model.
- **FR-002**: System MUST fall back to `gemini-2.5-flash-lite` on the third and final retry attempt.
- **FR-003**: System MUST implement retry logic with 3 attempts using exponential backoff with jitter (base 1 second, maximum 30 seconds).
- **FR-004**: System MUST return `{ "error": "rate_limited", "retry_after": 60 }` on final failure and MUST never expose raw AI provider error messages to the client.
- **FR-005**: System MUST return 503 if the `GEMINI_API_KEY` environment variable is not configured.
- **FR-006**: System MUST return 409 `{ "error": "analysis_pending" }` if core analysis results are not yet cached for the requested repository.
- **FR-007**: System MUST construct a shared context payload for all AI features containing: repository metadata, full file tree (paths only), README content (up to 8,000 tokens), package manifest contents (up to 2,000 tokens each for package.json, go.mod, Cargo.toml, requirements.txt, pyproject.toml, pom.xml, build.gradle), top 5 most complex files (first 150 lines each), and directory structure summary with file counts.
- **FR-008**: System MUST keep total AI context under 800,000 tokens, truncating file content first while prioritizing README and manifests.
- **FR-009**: System MUST provide an AI summary endpoint that streams a structured analysis (project description, architecture pattern, tech stack assessment, top 3 red flags, suggested entry points) via Server-Sent Events.
- **FR-010**: System MUST cache completed AI summaries with the key pattern `{owner}/{repo}:{sha}:ai_summary` and a TTL of 1 hour.
- **FR-011**: System MUST provide a chat endpoint accepting a message history and streaming AI responses via Server-Sent Events.
- **FR-012**: System MUST prepend repository context as a system-level message in chat interactions.
- **FR-013**: System MUST enforce a maximum of 20 conversation turns, truncating the oldest messages if exceeded.
- **FR-014**: System MUST rate-limit chat requests to 10 per IP per minute using an in-memory token bucket.
- **FR-015**: Chat responses MUST NOT be cached (each conversation is unique).
- **FR-016**: System MUST provide a health score endpoint returning a structured JSON response with an overall score (0-100), five category scores (readme_quality, contributing_guide, code_documentation, test_coverage_signals, project_hygiene) each with score and notes, and a list of top improvements.
- **FR-017**: System MUST configure the AI provider to return JSON output for health score requests.
- **FR-018**: System MUST retry health score generation once with a stricter prompt if JSON parsing fails.
- **FR-019**: System MUST cache health scores with the key pattern `{owner}/{repo}:{sha}:ai_health` and a TTL of 6 hours.

### Key Entities

- **GeminiClient**: Wraps the AI SDK with retry logic, model fallback, and error masking. Holds API key configuration and model references.
- **RepoContext**: Assembled context payload containing metadata, file tree, README, manifests, complex files, and directory summary. Shared input for all three AI features.
- **AISummary**: Structured AI-generated analysis containing project description, architecture pattern, tech stack, red flags, and entry points. Cached with 1-hour TTL.
- **ChatMessage**: A single message in a conversation with a role (user or model) and content string.
- **ChatRequest**: A collection of chat messages representing the full conversation history sent by the client.
- **HealthScore**: Structured JSON output with overall score, five category assessments, and top improvement suggestions. Cached with 6-hour TTL.
- **RateLimiter**: In-memory per-IP token bucket enforcing chat request limits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: AI summary requests for repositories with cached core analysis MUST begin streaming within 3 seconds.
- **SC-002**: Cached AI summaries MUST be returned within 100 milliseconds.
- **SC-003**: Chat responses MUST begin streaming within 3 seconds for repositories with cached core analysis.
- **SC-004**: Health score requests MUST return a complete valid JSON response within 30 seconds.
- **SC-005**: The system MUST successfully handle AI provider rate limiting without exposing internal error details, returning a structured retry response 100% of the time.
- **SC-006**: Context construction MUST complete within 2 seconds for repositories with up to 10,000 files.
- **SC-007**: The chat rate limiter MUST correctly enforce the 10-requests-per-minute-per-IP limit with zero false positives under normal load.
- **SC-008**: All three AI features MUST degrade gracefully when the AI API key is missing, returning informative 503 errors.

## Assumptions

- The Gemini free tier provides sufficient quota for development and moderate production use; the system does not implement billing or usage tracking.
- Token counting uses a simple character-based approximation (4 characters per token) rather than exact tokenizer counts, which is sufficient for context budget enforcement.
- The in-memory rate limiter state is local to each process and is lost on restart; this is acceptable for the initial single-instance deployment model.
- SSE streaming requires clients to handle `text/event-stream` content type; no WebSocket fallback is provided.
