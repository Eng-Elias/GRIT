# Implementation Plan: GRIT Frontend SPA

**Branch**: `007-frontend-spa` | **Date**: 2026-04-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-frontend-spa/spec.md`

## Summary

Build the GRIT frontend as a React 18 + TypeScript SPA with Vite, TailwindCSS dark theme, Recharts for standard charts, D3 v7 for the churn scatter plot, TanStack Query for data fetching with polling, and a custom `useSSE` hook for AI streaming. The build output lives in `frontend/dist` and is served by the Go backend as a static file fallback after API routes.

## Technical Context

**Language/Version**: TypeScript 5.x, React 18, Node 20+ (build only)
**Primary Dependencies**: Vite, TailwindCSS v3, Recharts, D3 v7, TanStack Query v5, React Router v6, clsx, lucide-react
**Storage**: localStorage (recent searches), in-memory (chat history, tab state)
**Testing**: Vitest + React Testing Library
**Target Platform**: Modern browsers (Chrome 100+, Firefox 100+, Safari 16+, Edge 100+)
**Project Type**: Single-page application (frontend for existing Go web service)
**Performance Goals**: <500KB initial bundle, <1s home page load, <100ms interaction response, <2s scatter plot with 500 points
**Constraints**: Dark theme only, no SSR, no separate frontend server in production, all data from Go API
**Scale/Scope**: 2 pages (Home, Repository), 6 tabs, ~25 components, ~10 API hooks

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. API-First Design | ✅ PASS | Frontend is a pure API consumer. No server-side rendering. Go backend serves compiled static files. |
| II. Modular Analysis Pillars | ✅ PASS | Frontend does not modify pillar architecture — reads from pillar endpoints only. |
| III. Async-First Execution | ✅ PASS | Frontend polls `/api/{owner}/{repo}/status` for job progress. Does not block on analysis. |
| IV. Cache-First with Redis | ✅ PASS | Frontend delegates caching to backend; respects `X-Cache` headers for display. |
| V. Defensive AI Integration | ✅ PASS | AI features are on-demand only (button-triggered). Handles 503 (AI unavailable) and 429 (rate limited) gracefully. |
| VI. Self-Hostable by Default | ✅ PASS | Build output is static files served by Go binary. No separate frontend server needed. |
| VII. Clean Handler Separation | ✅ N/A | No backend changes in this feature. |
| VIII. Test Discipline | ✅ PASS | Vitest + React Testing Library for component and hook tests. |
| Technology Stack | ✅ PASS | React 18 + TypeScript, Recharts, D3 v7, TailwindCSS — all mandated by constitution. |

**Gate Result**: ALL PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/007-frontend-spa/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (frontend data contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.ts
├── postcss.config.js
├── src/
│   ├── main.tsx                  # React entry point
│   ├── App.tsx                   # Router setup
│   ├── index.css                 # Tailwind base + custom styles
│   ├── api/
│   │   ├── client.ts             # Base fetch wrapper (error handling, base URL)
│   │   └── endpoints.ts          # Typed API endpoint functions
│   ├── hooks/
│   │   ├── useAnalysis.ts        # TanStack Query hook for core analysis
│   │   ├── useComplexity.ts      # TanStack Query hook for complexity
│   │   ├── useChurn.ts           # TanStack Query hook for churn matrix
│   │   ├── useContributors.ts    # TanStack Query hook for contributors
│   │   ├── useTemporal.ts        # TanStack Query hook for temporal data
│   │   ├── useStatus.ts          # TanStack Query hook with refetchInterval:3000
│   │   ├── useAISummary.ts       # Mutation + SSE streaming for summary
│   │   ├── useAIHealth.ts        # TanStack Query hook for health score
│   │   ├── useAIChat.ts          # SSE streaming hook for chat
│   │   ├── useSSE.ts             # Custom EventSource hook (shared)
│   │   ├── useRecentSearches.ts  # localStorage management
│   │   └── useBadge.ts           # Badge URL generation + clipboard
│   ├── pages/
│   │   ├── HomePage.tsx
│   │   └── RepoPage.tsx
│   ├── components/
│   │   ├── layout/
│   │   │   ├── Header.tsx
│   │   │   └── TabNav.tsx
│   │   ├── home/
│   │   │   ├── SearchBar.tsx
│   │   │   ├── RecentSearches.tsx
│   │   │   └── ExampleRepos.tsx
│   │   ├── repo/
│   │   │   ├── RepoHeader.tsx
│   │   │   ├── LanguageBar.tsx
│   │   │   ├── AnalysisStatus.tsx
│   │   │   └── BadgePanel.tsx
│   │   ├── overview/
│   │   │   ├── OverviewTab.tsx
│   │   │   ├── StatsCards.tsx
│   │   │   ├── CommitHeatmap.tsx
│   │   │   └── HealthSignals.tsx
│   │   ├── complexity/
│   │   │   ├── ComplexityTab.tsx
│   │   │   ├── ComplexitySummary.tsx
│   │   │   ├── DistributionBar.tsx
│   │   │   └── HotFilesTable.tsx
│   │   ├── churn/
│   │   │   ├── ChurnTab.tsx
│   │   │   ├── ScatterPlot.tsx     # D3 v7 SVG
│   │   │   ├── RiskZoneList.tsx
│   │   │   └── StaleFiles.tsx
│   │   ├── contributors/
│   │   │   ├── ContributorsTab.tsx
│   │   │   ├── BusFactor.tsx
│   │   │   ├── KeyPeople.tsx
│   │   │   ├── TopContributorsChart.tsx
│   │   │   └── ContributorsTable.tsx
│   │   ├── timeline/
│   │   │   ├── TimelineTab.tsx
│   │   │   ├── LOCAreaChart.tsx
│   │   │   ├── VelocityChart.tsx
│   │   │   └── PeriodSelector.tsx
│   │   ├── ai/
│   │   │   ├── AITab.tsx
│   │   │   ├── AISummary.tsx
│   │   │   ├── AIHealthGauge.tsx
│   │   │   ├── AIChat.tsx
│   │   │   └── ChatMessage.tsx
│   │   └── shared/
│   │       ├── Skeleton.tsx
│   │       ├── ErrorBanner.tsx
│   │       ├── EmptyState.tsx
│   │       └── RiskBadge.tsx
│   └── types/
│       ├── analysis.ts           # Core analysis response types
│       ├── complexity.ts         # Complexity response types
│       ├── churn.ts              # Churn response types
│       ├── contributors.ts       # Contributor response types
│       ├── temporal.ts           # Temporal response types
│       ├── ai.ts                 # AI response types (summary, health, chat)
│       └── status.ts             # Job status types
└── dist/                         # Build output (served by Go backend)
```

**Structure Decision**: Frontend-only addition to the existing Go monorepo. The `frontend/` directory is a self-contained Vite project. Build output in `frontend/dist` is served by the Go backend at `/*` after `/api/*` routes. No backend source changes are required except adding a static file handler (which will be a separate task).

## Complexity Tracking

> No violations — table not needed.
