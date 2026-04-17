# Feature Specification: GRIT Frontend SPA

**Feature Branch**: `007-frontend-spa`  
**Created**: 2026-04-17  
**Status**: Draft  
**Input**: User description: "Build the GRIT frontend — a React TypeScript SPA with dark theme, Recharts, D3, TanStack Query, and full coverage of all analysis pillars + AI features."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search and Discover a Repository (Priority: P1)

A developer visits GRIT, types a repository name (e.g., `facebook/react`) into the search bar, and is taken to the repository page. The system triggers analysis if no cached data exists, and the user sees a live progress banner showing which sub-jobs are running. Once core analysis completes, the Overview tab renders with stats cards, language bar, and commit heatmap. Previously searched repos appear in a "Recent Searches" list on the home page for quick re-access.

**Why this priority**: This is the entry point to the entire product. Without search → analysis → overview, no other feature is reachable.

**Independent Test**: Navigate to `/`, type `facebook/react`, press Enter. Verify redirect to `/repo/facebook/react`, analysis status banner appears, and Overview tab populates once data arrives.

**Acceptance Scenarios**:

1. **Given** the user is on the home page, **When** they type `facebook/react` and press Enter, **Then** they are navigated to `/repo/facebook/react` and analysis is triggered.
2. **Given** analysis is in progress, **When** the user views the repository page, **Then** a progress banner shows per-sub-job status (metadata, clone, file walk, etc.) that auto-updates via polling.
3. **Given** core analysis is cached, **When** the user navigates to the repository, **Then** the Overview tab renders immediately with stats cards, LOC language bar, and 52-week commit heatmap.
4. **Given** the user has previously searched repos, **When** they return to the home page, **Then** the last 5 searches are shown as clickable links.
5. **Given** the user enters an invalid or non-existent repository, **When** they submit the search, **Then** a clear error message is displayed ("Repository not found" or "Invalid repository name").

---

### User Story 2 - Explore Code Quality (Complexity & Churn) (Priority: P2)

A tech lead navigates to the Complexity tab to see mean complexity, p90, distribution, and hot files. They then switch to the Churn tab to view the scatter plot of churn vs. complexity, identify risk-zone files, and browse stale files. The scatter plot uses color-coded dots by language and shows a red-shaded quadrant for high-risk files.

**Why this priority**: Complexity and churn analysis are the core differentiators — the primary value proposition for engineering teams evaluating code health.

**Independent Test**: Navigate to `/repo/facebook/react`, click the Complexity tab, verify summary cards and hot files table render. Click the Churn tab, verify the scatter plot renders with hover tooltips and the risk zone list below.

**Acceptance Scenarios**:

1. **Given** complexity data is cached, **When** the user clicks the Complexity tab, **Then** summary cards (mean, p90, total functions), distribution bar, and hot files table render.
2. **Given** the user views the hot files table, **When** they click a column header, **Then** the table sorts by that column. **When** they select a language filter, **Then** only files in that language are shown.
3. **Given** churn data is cached, **When** the user clicks the Churn tab, **Then** the scatter plot renders with dots colored by language, p75 reference lines, and a red-shaded risk quadrant.
4. **Given** the user hovers over a dot in the scatter plot, **When** the tooltip appears, **Then** it shows file path, churn count, complexity score, LOC, and risk level.
5. **Given** churn data includes stale files, **When** the user views the Churn tab, **Then** a collapsible "Stale Files" section is visible below the risk list.

---

### User Story 3 - Understand Contributor Landscape (Priority: P3)

A team lead navigates to the Contributors tab to understand the bus factor, identify key people who own 80% of the codebase, and review the full contributors table with activity badges.

**Why this priority**: Bus factor and contributor analysis help teams identify knowledge concentration risks — critical for team planning but dependent on the more complex blame analysis.

**Independent Test**: Navigate to `/repo/golang/go`, click the Contributors tab, verify bus factor number, key people list, top-10 bar chart, and full contributors table render.

**Acceptance Scenarios**:

1. **Given** contributor data is cached, **When** the user clicks the Contributors tab, **Then** the bus factor is displayed as a large number with an explanation.
2. **Given** contributor data is available, **When** the user views the tab, **Then** a "Key People" section lists the authors who own 80%+ of the codebase.
3. **Given** contributor data is available, **When** the user views the tab, **Then** a top-10 contributors bar chart by LOC owned is visible.
4. **Given** the full contributors table is rendered, **When** the user views it, **Then** columns include name, lines owned, percentage, files touched, first commit, last commit, and an active/inactive badge.

---

### User Story 4 - Explore Timeline (Priority: P4)

A developer navigates to the Timeline tab to view LOC growth over time as a stacked area chart by language, refactor periods as shaded bands, weekly velocity as a dual bar chart, and commit cadence.

**Why this priority**: Temporal intelligence provides historical context that enriches all other analyses but is not required for the core experience.

**Independent Test**: Navigate to `/repo/torvalds/linux`, click the Timeline tab, verify stacked area chart renders with period selector (1y/2y/3y) and velocity chart below.

**Acceptance Scenarios**:

1. **Given** temporal data is cached, **When** the user clicks the Timeline tab, **Then** a stacked area chart of LOC over time by language renders with animation.
2. **Given** the stacked area chart is visible, **When** the user selects a different period (1y, 2y, 3y), **Then** the chart re-renders with the selected time range.
3. **Given** refactor periods exist in the data, **When** the timeline renders, **Then** refactor periods are overlaid as shaded vertical bands.
4. **Given** temporal data is available, **When** the user scrolls below the area chart, **Then** a weekly velocity dual bar chart (green additions, red deletions) with commit cadence line is visible.

---

### User Story 5 - AI Insights (Summary, Health, Chat) (Priority: P5)

A developer navigates to the AI tab and clicks "Generate Summary" to stream an AI codebase summary via SSE with typewriter effect. They then click "Analyze Health" to get a 0-100 health gauge with per-category breakdown. Finally, they use the chat interface to ask questions about the repository, seeing streaming responses and example question chips.

**Why this priority**: AI features are the "wow factor" but depend on all other analysis pillars being available. They are gated behind an optional API key and are the most expensive to run.

**Independent Test**: Navigate to `/repo/facebook/react`, click the AI tab, click "Generate Summary", verify SSE streaming typewriter display. Click "Analyze Health", verify gauge and categories render. Type a question in chat, verify streaming response.

**Acceptance Scenarios**:

1. **Given** the user is on the AI tab and core analysis is cached, **When** they click "Generate Summary", **Then** a POST request triggers SSE streaming and text appears with a typewriter effect.
2. **Given** an AI summary is already cached, **When** the user views the AI tab, **Then** the cached summary is displayed immediately and a "Regenerate" button is available.
3. **Given** the user clicks "Analyze Health", **When** the response arrives, **Then** a score gauge (0-100), per-category breakdown (5 categories), and improvement suggestions render.
4. **Given** the user types a question in the chat input, **When** they send it, **Then** the response streams in via SSE and is appended to the message history.
5. **Given** the AI API key is not configured on the backend, **When** the user clicks any AI action, **Then** a clear message states "AI features are not available" and no request is sent.
6. **Given** the user has sent 10 chat messages in one minute, **When** they try to send another, **Then** a rate limit notice is displayed.

---

### User Story 6 - Badge Generation (Priority: P6)

A developer clicks the gear icon in the repository page header to open a panel showing a shields.io badge URL for the repository, with a one-click copy-to-clipboard button.

**Why this priority**: Badges are a nice-to-have for README integration — low effort, low dependency, but adds shareability.

**Independent Test**: Navigate to any repository page, click the gear icon, verify badge URL is shown and clipboard copy works.

**Acceptance Scenarios**:

1. **Given** the user is on a repository page, **When** they click the gear icon in the header, **Then** a panel opens showing the shields.io badge URL.
2. **Given** the badge panel is open, **When** the user clicks the copy button, **Then** the badge URL is copied to the clipboard and a "Copied!" confirmation appears.

---

### Edge Cases

- What happens when the backend is unreachable? The app MUST show a connection error banner and retry on the next user action.
- What happens when a repository is private and no token is provided? The app MUST display a clear "Private repository" error with instructions to provide a GitHub token.
- What happens when the user resizes the browser below 768px? Charts MUST reflow responsively; tables MUST become horizontally scrollable.
- What happens when analysis completes for some pillars but not all? The app MUST show available data immediately and display skeleton loaders for pending sections.
- What happens when SSE connection drops during AI streaming? The app MUST show what was already received and display a "Connection lost" notice with a retry option.
- What happens when the user navigates away during analysis and returns? The app MUST resume showing current status without re-triggering analysis.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The app MUST be a single-page application with client-side routing between Home and Repository pages.
- **FR-002**: The Home page MUST provide a search input that accepts `owner/repo` format and navigates to the repository page on submit.
- **FR-003**: The Home page MUST display up to 5 recent searches from local storage as clickable links.
- **FR-004**: The Home page MUST display example repository links (facebook/react, golang/go, torvalds/linux).
- **FR-005**: The Repository page header MUST display: repo name, description, stars, forks, primary language pill, license, last push date, and a link to GitHub.
- **FR-006**: The Repository page MUST display a horizontal proportional LOC language bar showing each language's percentage.
- **FR-007**: The app MUST poll the `/api/{owner}/{repo}/status` endpoint and display per-sub-job progress when analysis is running.
- **FR-008**: The Repository page MUST have tab navigation: Overview, Complexity, Churn, Contributors, Timeline, AI.
- **FR-009**: The Overview tab MUST display a 52-week commit activity heatmap, stats cards (total LOC, files, commits, contributor count), and GitHub health signals.
- **FR-010**: The Complexity tab MUST display summary cards (mean, p90, total functions), a distribution bar, and a sortable/filterable hot files table.
- **FR-011**: The Churn tab MUST display an interactive scatter plot (churn vs. complexity) with language-colored dots, p75 reference lines, risk quadrant shading, and hover tooltips.
- **FR-012**: The Churn tab MUST display a risk zone file list and a collapsible stale files section.
- **FR-013**: The Contributors tab MUST display bus factor, key people, top-10 bar chart, and full contributors table with active/inactive badges.
- **FR-014**: The Timeline tab MUST display a stacked area chart of LOC over time with a period selector (1y/2y/3y), refactor band overlays, and a weekly velocity chart.
- **FR-015**: The AI tab MUST provide "Generate Summary" and "Analyze Health" buttons that trigger on-demand AI requests.
- **FR-016**: AI summary generation MUST stream via SSE and render with a typewriter effect. A "Regenerate" button MUST bust the cache.
- **FR-017**: AI health score MUST display as a gauge (0-100) with 5 per-category breakdowns and improvement suggestions.
- **FR-018**: The AI chat interface MUST display example question chips, support message history, and stream responses via SSE.
- **FR-019**: The AI tab MUST display a rate limit notice ("Limited to 10 requests per minute") and show a user-friendly message when the limit is reached.
- **FR-020**: The app MUST display a 503 notice ("AI features not available") when the backend returns 503 for AI endpoints.
- **FR-021**: Badge generation MUST be accessible via a gear icon in the header and provide a one-click copy of the shields.io badge URL.
- **FR-022**: All data sections MUST show skeleton loaders during fetch and display partial data when some analysis pillars are complete while others are pending.
- **FR-023**: The app MUST display clear error messages for: not found (404), private repo (403), rate limited (429), server error (500), and AI unavailable (503).
- **FR-024**: The app MUST use a dark theme exclusively — no light theme toggle.
- **FR-025**: All charts and tables MUST be responsive and usable on viewports down to 768px.

### Key Entities

- **Repository View**: The assembled display of a single repository combining data from core analysis, complexity, churn, contributors, temporal, and AI endpoints. Central entity around which all pages and tabs are organized.
- **Analysis Status**: The current state of background analysis jobs (queued, running, completed, failed) with per-sub-job granularity, used to drive the progress banner.
- **Recent Search**: A locally persisted record of the last 5 `owner/repo` searches, stored in the browser and displayed on the home page.
- **Tab State**: The current active tab on the repository page, reflected in the URL for shareability and bookmarkability.
- **Chat History**: An in-memory ordered list of user and assistant messages for the AI chat, persisted only for the current session.
- **SSE Stream**: A server-pushed event stream used for AI summary and chat responses, consumed by the frontend to render progressive output.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can search for a repository and see the Overview tab render with data within 3 seconds of cached data being available.
- **SC-002**: All 6 tabs render their primary content correctly when backend data is available, with no console errors or broken layouts.
- **SC-003**: The scatter plot on the Churn tab renders up to 500 data points with hover tooltips in under 2 seconds.
- **SC-004**: AI summary streaming renders the first character within 1 second of clicking "Generate Summary" (given backend responds promptly).
- **SC-005**: The app displays meaningful partial content when 2+ analysis pillars are complete but others are still running.
- **SC-006**: All error states (404, 403, 429, 500, 503, network error) show user-friendly messages — no raw JSON or stack traces are ever visible.
- **SC-007**: The app loads and renders the home page in under 1 second on a typical broadband connection (< 500KB initial bundle).
- **SC-008**: All interactive elements (tabs, buttons, tooltips, filters, sorting) respond within 100ms of user interaction.
