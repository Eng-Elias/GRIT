# Tasks: GRIT Frontend SPA

**Input**: Design documents from `/specs/007-frontend-spa/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api-client.md

**Tests**: Not explicitly requested — test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Vite + React + TypeScript project initialization with TailwindCSS dark theme

- [ ] T001 Scaffold Vite React-TS project in frontend/ with package.json, tsconfig.json, vite.config.ts (include /api proxy to localhost:8080)
- [ ] T002 Install dependencies: react, react-dom, react-router-dom, @tanstack/react-query, recharts, d3, tailwindcss, postcss, autoprefixer, clsx, lucide-react
- [ ] T003 [P] Configure TailwindCSS: tailwind.config.ts with darkMode:'class', zinc palette; postcss.config.js; frontend/src/index.css with @tailwind directives and dark custom styles
- [ ] T004 [P] Set dark class on html element in frontend/index.html, add base HTML structure
- [ ] T005 [P] Create all TypeScript type files from data-model.md in frontend/src/types/analysis.ts, complexity.ts, churn.ts, contributors.ts, temporal.ts, ai.ts, status.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: API client, shared hooks, routing shell, shared components — MUST be complete before any user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Implement base fetch wrapper with error handling and typed responses in frontend/src/api/client.ts (handle 400/403/404/409/429/500/503 per contract)
- [ ] T007 Implement typed API endpoint functions for all 12 endpoints in frontend/src/api/endpoints.ts
- [ ] T008 [P] Implement useSSE custom hook wrapping fetch+ReadableStream for POST-based SSE in frontend/src/hooks/useSSE.ts (parse data: lines, handle event:done, return chunks/error/isStreaming)
- [ ] T009 [P] Create shared Skeleton component in frontend/src/components/shared/Skeleton.tsx (pulse animation, configurable height/width/rows)
- [ ] T010 [P] Create shared ErrorBanner component in frontend/src/components/shared/ErrorBanner.tsx (maps error codes to user-friendly messages per contract)
- [ ] T011 [P] Create shared EmptyState component in frontend/src/components/shared/EmptyState.tsx (icon, title, description)
- [ ] T012 [P] Create shared RiskBadge component in frontend/src/components/shared/RiskBadge.tsx (color-coded: low/medium/high/critical)
- [ ] T013 Create App.tsx with React Router v6 routes (/ → HomePage, /repo/:owner/:repo → RepoPage) and QueryClientProvider in frontend/src/App.tsx
- [ ] T014 Create main.tsx entry point rendering App in frontend/src/main.tsx
- [ ] T015 [P] Create Header component with GRIT logo/title and navigation in frontend/src/components/layout/Header.tsx

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Search and Discover a Repository (Priority: P1) 🎯 MVP

**Goal**: Home page with search, recent searches, example repos → navigate to repo page → status polling → Overview tab with stats, language bar, heatmap

**Independent Test**: Navigate to /, type facebook/react, press Enter. Verify redirect, status banner, Overview tab populates.

### Home Page

- [ ] T016 [P] [US1] Implement SearchBar component with owner/repo input validation and Enter/submit navigation in frontend/src/components/home/SearchBar.tsx
- [ ] T017 [P] [US1] Implement useRecentSearches hook: read/write last 5 searches to localStorage in frontend/src/hooks/useRecentSearches.ts
- [ ] T018 [P] [US1] Implement RecentSearches component displaying last 5 searches as clickable links in frontend/src/components/home/RecentSearches.tsx
- [ ] T019 [P] [US1] Implement ExampleRepos component with hardcoded links (facebook/react, golang/go, torvalds/linux) in frontend/src/components/home/ExampleRepos.tsx
- [ ] T020 [US1] Implement HomePage assembling SearchBar, tagline, RecentSearches, ExampleRepos in frontend/src/pages/HomePage.tsx

### Repository Page Shell

- [ ] T021 [US1] Implement useAnalysis hook with TanStack Query fetching GET /api/{owner}/{repo} in frontend/src/hooks/useAnalysis.ts
- [ ] T022 [P] [US1] Implement useStatus hook with refetchInterval:3000, auto-disable on completed/failed, invalidate analysis on complete in frontend/src/hooks/useStatus.ts
- [ ] T023 [P] [US1] Implement AnalysisStatus banner component showing per-sub-job progress (clone, file_walk, metadata, commits) in frontend/src/components/repo/AnalysisStatus.tsx
- [ ] T024 [P] [US1] Implement RepoHeader component: repo name, description, stars, forks, language pill, license, last pushed, GitHub link in frontend/src/components/repo/RepoHeader.tsx
- [ ] T025 [P] [US1] Implement LanguageBar horizontal proportional bar from languages array in frontend/src/components/repo/LanguageBar.tsx
- [ ] T026 [P] [US1] Implement TabNav component with 6 tabs (Overview, Complexity, Churn, Contributors, Timeline, AI) reflecting active tab in URL in frontend/src/components/layout/TabNav.tsx

### Overview Tab

- [ ] T027 [P] [US1] Implement StatsCards component: total LOC, total files, total commits, contributor count in frontend/src/components/overview/StatsCards.tsx
- [ ] T028 [P] [US1] Implement CommitHeatmap using Recharts rendering 52-week commit activity grid in frontend/src/components/overview/CommitHeatmap.tsx
- [ ] T029 [P] [US1] Implement HealthSignals component: README, license, contributing guide, code of conduct presence from metadata in frontend/src/components/overview/HealthSignals.tsx
- [ ] T030 [US1] Implement OverviewTab assembling StatsCards, CommitHeatmap, HealthSignals, bus factor callout in frontend/src/components/overview/OverviewTab.tsx

### Repo Page Assembly

- [ ] T031 [US1] Implement RepoPage assembling RepoHeader, LanguageBar, AnalysisStatus, TabNav, lazy-loaded tab content with React.lazy + Suspense in frontend/src/pages/RepoPage.tsx

**Checkpoint**: US1 complete — home page → search → repo page → overview tab fully functional

---

## Phase 4: User Story 2 — Explore Code Quality (Priority: P2)

**Goal**: Complexity tab with summary/distribution/hot files + Churn tab with D3 scatter plot, risk zone, stale files

**Independent Test**: Navigate to repo, click Complexity tab, verify cards + table. Click Churn tab, verify scatter plot + risk list.

### Complexity Tab

- [ ] T032 [P] [US2] Implement useComplexity hook fetching GET /api/{owner}/{repo}/complexity in frontend/src/hooks/useComplexity.ts
- [ ] T033 [P] [US2] Implement ComplexitySummary cards (mean, p90, total functions) in frontend/src/components/complexity/ComplexitySummary.tsx
- [ ] T034 [P] [US2] Implement DistributionBar showing Low/Medium/High/Critical counts as horizontal segments in frontend/src/components/complexity/DistributionBar.tsx
- [ ] T035 [P] [US2] Implement HotFilesTable: top 20 by complexity_density, sortable columns (path, cyclomatic, cognitive, functions, LOC, risk badge), language filter dropdown in frontend/src/components/complexity/HotFilesTable.tsx
- [ ] T036 [US2] Implement ComplexityTab assembling ComplexitySummary, DistributionBar, HotFilesTable with skeleton loading in frontend/src/components/complexity/ComplexityTab.tsx

### Churn Tab

- [ ] T037 [P] [US2] Implement useChurn hook fetching GET /api/{owner}/{repo}/churn-matrix in frontend/src/hooks/useChurn.ts
- [ ] T038 [US2] Implement ScatterPlot with D3 v7: SVG axes, language-colored dots, p75 reference lines from thresholds, red-shaded risk quadrant, hover tooltip (path, churn, complexity, LOC, risk_level) in frontend/src/components/churn/ScatterPlot.tsx
- [ ] T039 [P] [US2] Implement RiskZoneList table showing risk_zone entries with RiskBadge in frontend/src/components/churn/RiskZoneList.tsx
- [ ] T040 [P] [US2] Implement StaleFiles collapsible section listing stale files with months_inactive in frontend/src/components/churn/StaleFiles.tsx
- [ ] T041 [US2] Implement ChurnTab assembling ScatterPlot, RiskZoneList, StaleFiles with skeleton loading in frontend/src/components/churn/ChurnTab.tsx

**Checkpoint**: US2 complete — Complexity + Churn tabs render all data with interactive features

---

## Phase 5: User Story 3 — Understand Contributor Landscape (Priority: P3)

**Goal**: Contributors tab with bus factor, key people, top-10 chart, full contributors table

**Independent Test**: Navigate to repo, click Contributors tab, verify bus factor, key people, chart, table.

- [ ] T042 [P] [US3] Implement useContributors hook fetching GET /api/{owner}/{repo}/contributors in frontend/src/hooks/useContributors.ts
- [ ] T043 [P] [US3] Implement BusFactor component: large number display with explanation text in frontend/src/components/contributors/BusFactor.tsx
- [ ] T044 [P] [US3] Implement KeyPeople component: list of authors owning 80%+ with name and lines in frontend/src/components/contributors/KeyPeople.tsx
- [ ] T045 [P] [US3] Implement TopContributorsChart using Recharts horizontal BarChart: top 10 by total_lines_owned in frontend/src/components/contributors/TopContributorsChart.tsx
- [ ] T046 [P] [US3] Implement ContributorsTable: name, lines owned, %, files touched, first/last commit, active badge in frontend/src/components/contributors/ContributorsTable.tsx
- [ ] T047 [US3] Implement ContributorsTab assembling BusFactor, KeyPeople, TopContributorsChart, ContributorsTable with skeleton loading in frontend/src/components/contributors/ContributorsTab.tsx

**Checkpoint**: US3 complete — Contributors tab fully functional

---

## Phase 6: User Story 4 — Explore Timeline (Priority: P4)

**Goal**: Timeline tab with stacked area chart, period selector, refactor bands, velocity chart

**Independent Test**: Navigate to repo, click Timeline tab, verify stacked area chart with period selector and velocity chart.

- [ ] T048 [P] [US4] Implement useTemporal hook fetching GET /api/{owner}/{repo}/temporal with period query param in frontend/src/hooks/useTemporal.ts
- [ ] T049 [P] [US4] Implement PeriodSelector component with 1y/2y/3y toggle buttons in frontend/src/components/timeline/PeriodSelector.tsx
- [ ] T050 [US4] Implement LOCAreaChart using Recharts StackedArea: LOC by language over time, animated, with RefactorPeriod shaded vertical bands using ReferenceArea in frontend/src/components/timeline/LOCAreaChart.tsx
- [ ] T051 [US4] Implement VelocityChart using Recharts: dual BarChart (green additions, red deletions) with commit cadence Line overlay in frontend/src/components/timeline/VelocityChart.tsx
- [ ] T052 [US4] Implement TimelineTab assembling PeriodSelector, LOCAreaChart, VelocityChart with skeleton loading in frontend/src/components/timeline/TimelineTab.tsx

**Checkpoint**: US4 complete — Timeline tab fully functional

---

## Phase 7: User Story 5 — AI Insights (Priority: P5)

**Goal**: AI tab with SSE summary streaming, health score gauge, chat interface

**Independent Test**: Click AI tab → Generate Summary (typewriter), Analyze Health (gauge), Chat (streaming response).

### AI Summary

- [ ] T053 [P] [US5] Implement useAISummary hook: POST /api/{owner}/{repo}/ai/summary via useSSE, cache check, regenerate (bust cache) in frontend/src/hooks/useAISummary.ts
- [ ] T054 [US5] Implement AISummary component: "Generate Summary" button, typewriter SSE display, cached summary display, "Regenerate" button in frontend/src/components/ai/AISummary.tsx

### AI Health

- [ ] T055 [P] [US5] Implement useAIHealth hook fetching GET /api/{owner}/{repo}/ai/health in frontend/src/hooks/useAIHealth.ts
- [ ] T056 [US5] Implement AIHealthGauge: 0-100 gauge (SVG arc or Recharts RadialBar), 5 category breakdowns, improvement suggestions list in frontend/src/components/ai/AIHealthGauge.tsx

### AI Chat

- [ ] T057 [P] [US5] Implement useAIChat hook: POST /api/{owner}/{repo}/ai/chat via useSSE, manage message history, handle 429 in frontend/src/hooks/useAIChat.ts
- [ ] T058 [P] [US5] Implement ChatMessage component: user/assistant styling, markdown-friendly rendering in frontend/src/components/ai/ChatMessage.tsx
- [ ] T059 [US5] Implement AIChat component: example question chips, message list, input field, streaming display, rate limit notice in frontend/src/components/ai/AIChat.tsx

### AI Tab Assembly

- [ ] T060 [US5] Implement AITab assembling AISummary, AIHealthGauge, AIChat with 503 "AI unavailable" guard and status note in frontend/src/components/ai/AITab.tsx

**Checkpoint**: US5 complete — AI summary, health, and chat all functional

---

## Phase 8: User Story 6 — Badge Generation (Priority: P6)

**Goal**: Badge URL panel with copy-to-clipboard

**Independent Test**: Click gear icon in header, verify badge URL shown, copy works.

- [ ] T061 [P] [US6] Implement useBadge hook: generate shields.io badge URL from owner/repo, clipboard copy with feedback in frontend/src/hooks/useBadge.ts
- [ ] T062 [US6] Implement BadgePanel component: gear icon trigger, panel with badge URL, preview image, copy button with "Copied!" feedback in frontend/src/components/repo/BadgePanel.tsx
- [ ] T063 [US6] Wire BadgePanel into RepoHeader gear icon in frontend/src/components/repo/RepoHeader.tsx

**Checkpoint**: US6 complete — badge generation fully functional

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Error states, responsiveness, Go static serving, final validation

- [ ] T064 [P] Ensure all error states render correctly: 404 not found page, 403 private repo banner, 429 rate limit with countdown, 500 generic error, 503 AI unavailable, network error banner in frontend/src/components/shared/ErrorBanner.tsx and frontend/src/pages/RepoPage.tsx
- [ ] T065 [P] Add responsive breakpoints: charts reflow on <768px, tables get horizontal scroll wrapper, tab nav wraps in frontend/src/index.css and relevant components
- [ ] T066 Add Go backend static file handler serving frontend/dist/ at /* with SPA fallback to index.html (after /api/* routes) in cmd/grit/main.go
- [ ] T067 Verify production build: run npm run build in frontend/, confirm dist/ output <500KB, no TypeScript errors
- [ ] T068 Run quickstart.md validation: dev server loads home page, search triggers analysis, all 6 tabs render with cached data, AI 503 when no key

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — provides the page shell all other stories render within
- **US2–US6 (Phases 4–8)**: Depend on Phase 2 (foundational) + T031 (RepoPage shell with TabNav + lazy loading)
- **Polish (Phase 9)**: Depends on all stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2. Creates the RepoPage shell used by all other stories.
- **US2 (P2)**: Depends on Phase 2 + T031 (RepoPage). Can run parallel with US3–US6.
- **US3 (P3)**: Depends on Phase 2 + T031 (RepoPage). Can run parallel with US2, US4–US6.
- **US4 (P4)**: Depends on Phase 2 + T031 (RepoPage). Can run parallel with US2, US3, US5–US6.
- **US5 (P5)**: Depends on Phase 2 + T031 (RepoPage) + T008 (useSSE). Can run parallel with US2–US4, US6.
- **US6 (P6)**: Depends on Phase 2 + T024 (RepoHeader). Can run parallel with US2–US5.

### Within Each User Story

- Hooks before components that use them
- Child components before parent assembler (e.g., StatsCards before OverviewTab)
- [P] tasks within a story can run in parallel

### Parallel Opportunities

- T003, T004, T005 in Phase 1 (all different files)
- T008–T012, T015 in Phase 2 (all independent shared components/hooks)
- T016–T019 in US1 (all independent home components)
- T022–T026 in US1 (all independent repo shell components)
- T027–T029 in US1 (all independent overview components)
- T032–T035, T037, T039, T040 in US2 (all independent components)
- T042–T046 in US3 (all independent components)
- T048–T049 in US4 (hook + selector independent)
- T053, T055, T057, T058 in US5 (all independent hooks/components)
- T064, T065 in Phase 9 (different concerns)

---

## Parallel Example: User Story 2

```text
# All independent hook + component tasks in parallel:
T032: useComplexity hook
T033: ComplexitySummary cards
T034: DistributionBar
T035: HotFilesTable
T037: useChurn hook
T039: RiskZoneList
T040: StaleFiles

# Then sequential assembly:
T036: ComplexityTab (depends on T032–T035)
T038: ScatterPlot (depends on T037 for data shape)
T041: ChurnTab (depends on T037–T040)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T005)
2. Complete Phase 2: Foundational (T006–T015)
3. Complete Phase 3: User Story 1 (T016–T031)
4. **STOP and VALIDATE**: Home page search → repo page → Overview tab works end-to-end
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US1 → Search + Overview works → **MVP!**
3. US2 → Complexity + Churn tabs → quality analysis
4. US3 → Contributors tab → team insights
5. US4 → Timeline tab → historical context
6. US5 → AI tab → intelligent insights
7. US6 → Badge generation → shareability
8. Polish → error handling, responsiveness, Go serving

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable after Phase 2
- D3 v7 is ONLY used in T038 (ScatterPlot) — all other charts use Recharts
- SSE hook (T008) is shared infrastructure used by US5 (AI summary + chat)
- Go backend static file serving (T066) is a backend change in the polish phase
- Commit after each task or logical group
