# GRIT — Git Repo Intelligence Tool

**Deep-dive analytics for any public GitHub repository.** GRIT clones, parses, and analyzes codebases to surface complexity hotspots, churn risk, contributor ownership, temporal trends, and AI-powered insights — all through a clean web UI.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)
![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-6.0-3178C6?style=flat-square&logo=typescript)
![License](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey?style=flat-square)

---

## Features

- **Repository Overview** — Stars, forks, languages, commit heatmap, health signals
- **Complexity Analysis** — Cyclomatic complexity per file, distribution breakdown, hot file detection
- **Churn Risk Matrix** — D3 scatter plot of churn vs complexity, P75 risk zones, stale file detection
- **Contributor Insights** — Bus factor gauge, code ownership pie chart, key people, active/inactive status
- **Temporal Trends** — LOC over time, weekly velocity (additions/deletions), refactor period detection
- **AI-Powered** — Gemini-based repo summaries, health scoring, interactive chat about the codebase
- **Badge Generation** — Shields.io badges for LOC, files, languages, complexity, bus factor

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│   Go API     │────▶│    Redis    │
│  React SPA   │     │   (Chi)      │     │   (Cache)    │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────▼───────┐     ┌──────────────┐
                     │   Workers    │────▶│    NATS      │
                     │  (Analysis)  │     │ (JetStream)  │
                     └──────────────┘     └──────────────┘
```

| Component | Stack |
|-----------|-------|
| **Backend** | Go, Chi router, NATS JetStream, Redis, Prometheus metrics |
| **Frontend** | React 19, TypeScript, TailwindCSS v4, TanStack Query, Recharts, D3 |
| **AI** | Google Gemini API (streaming SSE) |
| **Infrastructure** | Docker Compose, Caddy reverse proxy |

## API Endpoints

All endpoints are prefixed with `/api/{owner}/{repo}`.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Core analysis (triggers analysis if not cached) |
| `GET` | `/status` | Job status polling |
| `GET` | `/complexity` | Complexity metrics per file |
| `GET` | `/churn-matrix` | Churn + complexity risk matrix |
| `GET` | `/contributors` | Contributor ownership and bus factor |
| `GET` | `/contributors/files` | Per-file contributor breakdown |
| `GET` | `/temporal` | LOC over time, velocity, refactor periods |
| `POST` | `/ai/summary` | AI-generated repo summary (SSE stream) |
| `POST` | `/ai/chat` | Multi-turn AI chat about the repo (SSE stream) |
| `GET` | `/ai/health` | AI-scored health assessment |
| `GET` | `/badge` | SVG badge endpoint |
| `DELETE` | `/cache` | Invalidate cached analysis |
| `GET` | `/metrics` | Prometheus metrics (root-level) |

## Quick Start

### Prerequisites

- **Go 1.25+**
- **Node.js 20+**
- **Redis 7+**
- **NATS 2+** (with JetStream enabled)
- **Gemini API key** (optional, for AI features) — [get one free](https://aistudio.google.com/apikey)

### 1. Clone & configure

```bash
git clone https://github.com/Eng-Elias/GRIT.git
cd GRIT
cp .env.example .env
# Edit .env with your GITHUB_TOKEN and GEMINI_API_KEY
```

### 2. Start infrastructure

```bash
docker compose up redis nats -d
```

### 3. Run the backend

```bash
go run ./cmd/grit/
```

The API starts on `http://localhost:8080`.

### 4. Run the frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`. The Vite dev server proxies `/api` to the Go backend.

## Docker Compose (Full Stack)

```bash
cp .env.example .env
# Fill in GITHUB_TOKEN and GEMINI_API_KEY
docker compose up --build
```

This starts:
- **api** — Go backend on port 8080
- **worker** — Background analysis workers
- **redis** — Cache with healthcheck
- **nats** — JetStream message broker
- **caddy** — Reverse proxy on ports 80/443

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GITHUB_TOKEN` | — | GitHub PAT (60 req/h without, 5000 with) |
| `GEMINI_API_KEY` | — | Google Gemini API key for AI features |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `NATS_URL` | `nats://localhost:4222` | NATS connection URL |
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log verbosity |
| `CLONE_DIR` | `/tmp/grit-clones` | Disk clone directory |
| `CLONE_SIZE_THRESHOLD_KB` | `51200` | Max repo size for in-memory clone |
| `CACHE_TTL_CORE` | `24` | Core analysis cache TTL (hours) |
| `CACHE_TTL_COMPLEXITY` | `24` | Complexity cache TTL (hours) |
| `CACHE_TTL_TEMPORAL` | `12` | Temporal data cache TTL (hours) |
| `CACHE_TTL_CONTRIBUTOR` | `48` | Contributor data cache TTL (hours) |
| `CACHE_TTL_AI_SUMMARY` | `1` | AI summary cache TTL (hours) |
| `CACHE_TTL_AI_HEALTH` | `6` | AI health score cache TTL (hours) |
| `MAX_COMMIT_HISTORY` | `5000` | Max commits to analyze |
| `MAX_REPO_SIZE_MB` | `500` | Max repository size |
| `ANALYSIS_TIMEOUT_MINUTES` | `10` | Analysis job timeout |

## Project Structure

```
grit/
├── cmd/grit/           # Application entry point
├── internal/
│   ├── ai/             # Gemini AI client, rate limiter
│   ├── analysis/       # Core, complexity, churn, blame analyzers
│   ├── cache/          # Redis cache layer
│   ├── clone/          # Git clone (in-memory & disk)
│   ├── config/         # Environment config loader
│   ├── github/         # GitHub API client
│   ├── handler/        # HTTP handlers (Chi)
│   ├── job/            # NATS JetStream workers & publisher
│   ├── metrics/        # Prometheus metrics
│   └── models/         # Shared data models
├── frontend/           # React TypeScript SPA
│   ├── src/
│   │   ├── api/        # Typed fetch client & endpoints
│   │   ├── components/ # UI components by feature
│   │   ├── hooks/      # React Query & SSE hooks
│   │   ├── pages/      # HomePage, RepoPage
│   │   └── types/      # TypeScript type definitions
│   └── vite.config.ts
├── Dockerfile
├── docker-compose.yml
├── Caddyfile
└── .env.example
```

## Development

```bash
# Run Go tests
go test ./...

# Type-check frontend
cd frontend && npx tsc -b --noEmit

# Lint frontend
cd frontend && npm run lint

# Production build
cd frontend && npm run build
```

## License

This project is licensed under the [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-nc-sa/4.0/legalcode) license.

Copyright (c) 2025 Eng. Elias Owis

You are free to share and adapt this material for non-commercial purposes, with attribution, under the same license terms.
