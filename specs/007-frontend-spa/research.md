# Research: GRIT Frontend SPA

## R1: Vite + React 18 + TypeScript Project Setup

**Decision**: Vite 5 with `react-ts` template, React 18.3, TypeScript 5.x strict mode.

**Rationale**: Vite provides instant HMR, fast builds, and native ESM. The `react-ts` template provides sensible defaults. React 18 is the constitution-mandated version.

**Alternatives considered**:
- Create React App — deprecated, slow builds, no ESM.
- Next.js — SSR/SSG features are unnecessary; constitution mandates no server-side rendering.

## R2: TailwindCSS v3 Dark Theme Strategy

**Decision**: `darkMode: 'class'` in `tailwind.config.ts`, `class="dark"` permanently on `<html>` element in `index.html`. No light theme. Extend Tailwind's default palette with a custom dark color scheme using `zinc` as the base gray scale.

**Rationale**: Class-based dark mode allows full Tailwind utility usage without media queries. Fixing `dark` class on `<html>` ensures all `dark:` variants are always active. Zinc gray scale provides a modern, professional dark appearance similar to Vercel's dashboard.

**Alternatives considered**:
- `darkMode: 'media'` — would require OS dark mode and prevent fixed dark-only approach.
- CSS variables approach — more complex, no benefit since we only have one theme.

## R3: Data Fetching with TanStack Query v5

**Decision**: TanStack Query v5 (`@tanstack/react-query`) with `QueryClientProvider` at app root. Each API endpoint gets a dedicated hook. Status polling uses `refetchInterval: 3000` that auto-disables when status is `completed` or `failed`.

**Rationale**: TanStack Query handles caching, deduplication, background refetch, error retry, and loading states out of the box. The `refetchInterval` pattern is the recommended approach for polling.

**Alternatives considered**:
- SWR — similar capabilities but smaller community and less flexible polling.
- Raw `fetch` + `useEffect` — error-prone, no caching, manual loading states.

## R4: SSE Streaming with Custom useSSE Hook

**Decision**: Custom `useSSE` hook wrapping the browser `EventSource` API. For POST endpoints (AI summary, chat), use `fetch` with `ReadableStream` since `EventSource` only supports GET. Parse `data:` lines manually. Return `{ data, error, isStreaming }`.

**Rationale**: The native `EventSource` API only supports GET requests. AI summary and chat endpoints are POST. Using `fetch` + `ReadableStream` with manual SSE line parsing is the standard pattern for POST-based SSE.

**Alternatives considered**:
- `eventsource` npm package — adds dependency for minimal benefit, still GET-only.
- `sse.js` — thin wrapper but POST support is inconsistent.
- WebSocket — requires backend changes, SSE is already implemented.

## R5: Recharts vs D3 Division of Labor

**Decision**: Recharts for all standard charts (bar charts, area charts, heatmaps, line charts). D3 v7 for the churn scatter plot only (SVG-based with custom interactions).

**Rationale**: Recharts is React-native, declarative, and handles 90% of chart needs. The scatter plot requires custom quadrant shading, p75 reference lines, language-colored dots, and rich hover tooltips — D3's imperative SVG API is better suited for this level of customization.

**Alternatives considered**:
- All D3 — higher implementation cost for standard charts, poor React integration.
- All Recharts — ScatterChart exists but lacks custom SVG overlays (quadrant shading, reference lines).

## R6: Go Backend Static File Serving

**Decision**: Add a catch-all handler in `main.go` that serves files from `frontend/dist/` for any non-`/api/*` and non-`/metrics` path. Falls back to `index.html` for client-side routing (SPA fallback).

**Rationale**: Constitution mandates "The backend MUST serve the compiled React frontend as static files — no separate frontend server in production." This is the standard SPA serving pattern.

**Alternatives considered**:
- Caddy serves frontend separately — adds complexity, violates single-binary goal.
- Embed files in Go binary — possible but complicates dev workflow; `frontend/dist` directory serving is simpler and more standard.

## R7: Vite Dev Proxy for API Calls

**Decision**: Configure `vite.config.ts` with `server.proxy: { '/api': 'http://localhost:8080' }` to proxy API calls to the Go backend during development.

**Rationale**: Avoids CORS issues during development without requiring CORS headers on the backend. Standard Vite pattern.

**Alternatives considered**:
- CORS headers on Go backend — works but leaks development concerns into production code.
- Separate API base URL env var — more complex, proxy is simpler.

## R8: Bundle Size Strategy

**Decision**: Target <500KB initial bundle. Lazy-load tab components via `React.lazy()` + `Suspense`. D3 is only imported in `ScatterPlot.tsx` (tree-shaken). Recharts components are individually imported.

**Rationale**: Keeping the initial bundle small ensures fast first load. Tab-level code splitting means users only download the code for tabs they visit.

**Alternatives considered**:
- No code splitting — simpler but risks >1MB bundle with D3 + Recharts.
- Route-level splitting only — tabs within RepoPage are the more impactful split points.
