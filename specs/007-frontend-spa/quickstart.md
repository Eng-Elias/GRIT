# Quickstart: GRIT Frontend SPA

## Prerequisites

- Node.js 20+ and npm 10+
- Go backend running on `localhost:8080` (with Redis and NATS)

## Development Setup

```bash
# From repo root
cd frontend

# Install dependencies
npm install

# Start dev server (proxies /api/* to localhost:8080)
npm run dev
# → http://localhost:5173
```

## Build for Production

```bash
cd frontend
npm run build
# → Output: frontend/dist/
```

The Go backend serves `frontend/dist/` at `/*` after `/api/*` routes.

## Verify Setup

1. **Home page loads**: Navigate to `http://localhost:5173` — see search bar, tagline, example repos.
2. **API proxy works**: Search for `facebook/react` — repo page loads, status banner appears.
3. **Tabs render**: Click each tab — skeleton loaders appear for pending data, real data renders when cached.
4. **Dark theme**: All backgrounds should be dark (zinc-900/950), text white/gray.

## Key Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Start Vite dev server with HMR |
| `npm run build` | Production build to `dist/` |
| `npm run preview` | Preview production build locally |
| `npm run lint` | Run ESLint |
| `npm run test` | Run Vitest |
| `npm run test:watch` | Run Vitest in watch mode |

## Environment

No frontend-specific environment variables. The API base URL is always same-origin:
- **Dev**: Vite proxy routes `/api/*` → `http://localhost:8080`
- **Prod**: Go backend serves both API and static files

## Architecture Notes

- **No separate frontend server in production** — Go binary serves everything.
- **Dark theme only** — `class="dark"` is hardcoded on `<html>`.
- **Tab code splitting** — Each tab component is lazy-loaded via `React.lazy()`.
- **SSE for AI** — AI summary and chat use POST-based SSE via `fetch` + `ReadableStream`.
- **Status polling** — TanStack Query `refetchInterval: 3000` with auto-disable on completion.
