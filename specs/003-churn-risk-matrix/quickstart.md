# Quickstart: Churn Matrix API

**Feature**: 003-churn-risk-matrix

## Prerequisites

- Docker Compose running (`docker compose up -d`)
- A repository has been analyzed (core + complexity analysis completed)

## Usage

### 1. Trigger analysis (if not already done)

```bash
curl -s http://localhost:8080/api/facebook/react | jq '.complexity_summary.status'
# Should return "complete" — complexity must finish before churn can run
```

Churn analysis is auto-triggered after complexity completes. No manual trigger needed.

### 2. Check churn analysis status

```bash
curl -s -w "\nHTTP %{http_code}\n" http://localhost:8080/api/facebook/react/churn-matrix
```

- **202**: Churn analysis is in progress — poll again in a few seconds
- **200**: Churn data is ready

### 3. Get churn matrix results

```bash
curl -s http://localhost:8080/api/facebook/react/churn-matrix | jq .
```

Expected response includes:
- `churn`: per-file commit counts
- `risk_matrix`: files classified by risk level
- `risk_zone`: critical files sorted by churn × complexity
- `thresholds`: p50/p75/p90 percentiles for churn and complexity
- `stale_files`: potentially unused files

### 4. View risk zone (critical files)

```bash
curl -s http://localhost:8080/api/facebook/react/churn-matrix | jq '.risk_zone[] | {path, churn, complexity_cyclomatic, risk_level}'
```

### 5. View stale files

```bash
curl -s http://localhost:8080/api/facebook/react/churn-matrix | jq '.stale_files[] | {path, months_inactive}'
```

### 6. Check churn summary in main analysis response

```bash
curl -s http://localhost:8080/api/facebook/react | jq '.churn_summary'
```

Expected:
```json
{
  "status": "complete",
  "total_files": 856,
  "critical_count": 12,
  "stale_count": 3,
  "churn_matrix_url": "/api/facebook/react/churn-matrix"
}
```

### 7. Clear cache (includes churn data)

```bash
curl -s -X DELETE http://localhost:8080/api/facebook/react/cache
```

This clears core, complexity, AND churn cache entries.

## Verification Checklist

- [ ] `GET /api/:owner/:repo/churn-matrix` returns 200 with `X-Cache: HIT`
- [ ] Response contains `churn`, `risk_matrix`, `risk_zone`, `thresholds`, `stale_files`
- [ ] `risk_zone` contains only critical-level files
- [ ] `risk_zone` is sorted by churn × complexity descending
- [ ] `thresholds` includes all 6 percentile values
- [ ] `stale_files` entries have `months_inactive >= 6`
- [ ] Main analysis response includes `churn_summary`
- [ ] Cache deletion clears churn data
- [ ] Second request returns `X-Cache: HIT`
