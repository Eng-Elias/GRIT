# Quickstart: Core Analysis Engine

**Branch**: `001-core-analysis`
**Date**: 2026-03-23

## Prerequisites

- Go 1.22+
- Docker and Docker Compose
- A GitHub personal access token (optional, for higher rate limits)

## 1. Clone and Setup

```bash
git clone <repo-url> grit
cd grit
git checkout 001-core-analysis
cp .env.example .env
```

Edit `.env` to set your GitHub token (optional):

```env
GITHUB_TOKEN=ghp_your_token_here
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
PORT=8080
CLONE_DIR=/tmp/grit-clones
CLONE_SIZE_THRESHOLD_KB=51200
```

## 2. Start Infrastructure

```bash
docker compose up -d redis nats
```

This starts Redis 7 on port 6379 and NATS with JetStream on port 4222.

## 3. Run the Backend

```bash
go mod download
go run ./cmd/grit/
```

The server starts on `http://localhost:8080`.

## 4. Analyze a Repository

### Request analysis:

```bash
curl -s http://localhost:8080/api/sindresorhus/is
```

Expected response (first request — cache miss):

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "poll_url": "/api/sindresorhus/is/status"
}
```

### Poll for status:

```bash
curl -s http://localhost:8080/api/sindresorhus/is/status
```

### Get completed result:

Once status shows `"completed"`, fetch the full result:

```bash
curl -s http://localhost:8080/api/sindresorhus/is | head -50
```

### Get a badge:

```bash
curl -s http://localhost:8080/api/sindresorhus/is/badge
```

### Bust the cache:

```bash
curl -X DELETE http://localhost:8080/api/sindresorhus/is/cache
```

## 5. Run with Docker Compose (full stack)

```bash
docker compose up --build
```

This builds the Go binary and starts grit + Redis + NATS together.
The API is available at `http://localhost:8080`.

## 6. Run Tests

```bash
go test ./... -v
```

Integration tests require Redis and NATS to be running:

```bash
docker compose up -d redis nats
go test ./internal/... -v -tags=integration
```

## Verification Checklist

- [ ] `GET /api/sindresorhus/is` returns 202 on first request
- [ ] `GET /api/sindresorhus/is/status` shows sub-job progress
- [ ] Second `GET /api/sindresorhus/is` returns 200 with cached result
- [ ] Response includes `X-Cache: HIT` header on cached request
- [ ] `GET /api/sindresorhus/is/badge` returns valid shields.io JSON
- [ ] `DELETE /api/sindresorhus/is/cache` returns 204
- [ ] `GET /api/nonexistent/repo` returns 404 with error envelope
- [ ] `GET /metrics` returns Prometheus metrics
