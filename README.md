# GitHub Recap 2025 (Wrapped-style) — `dennislee928`

A runnable repo skeleton that generates a **2025 GitHub Recap** (HTML report + shareable PNG cards), using:
- **GitHub GraphQL API** for contribution calendar + contribution totals (commits/PRs/issues/reviews)
- **GitHub Search (GraphQL)** for PR/Issue details (merge rate, average merge time, biggest PR, etc.)
- **GitHub REST API** for repo languages + star/fork growth (based on timestamped stargazers & forks where available)

## What you get

### A. Stable metrics (included)
- Total contributions (2025): commits / PRs / issues / reviews
- Contribution calendar heatmap (daily)
- Top repos by activity (commit/PR/issue counts per repo)
- PR stats: opened, merged, average time-to-merge
- Issue stats: opened (created in 2025), closed (closed in 2025) — for issues authored by you
- Languages: weighted by repo language bytes (aggregated across repos you contributed to)
- Stars / Forks gained in 2025 (for **your owned repos**):
  - Stars gained: derived from `stargazers` with `starred_at` (when available)
  - Forks gained: derived from forks list `created_at`

> Note: star/fork *growth* requires iterating stargazers/forks, which can be rate-limit heavy if you have large repos.
> You can disable it with `--skip-growth`.

### B. “Recap-style” extras (included)
- Longest streak (consecutive contribution days)
- Most productive day + most productive ISO week
- Biggest PR (additions + deletions)
- Review impact: total review contributions (from contributionsCollection)
- Time-of-day pattern: based on PR/Issue creation timestamps (commits are not available as a full-year event stream via public APIs)

## Requirements
- Go >= 1.22
- Node >= 18
- A **GitHub Personal Access Token** (Classic PAT recommended):
  - To include **private contributions**, the token must belong to `dennislee928` (you) and include `repo` scope.
  - Also include `read:user` (safe default).

## Quick start

### 1) Create a token
Create a GitHub PAT (Classic):
- Scopes: `repo`, `read:user`

### 2) Export env vars
```bash
cp .env.example .env
# edit .env and paste your token
set -a; source .env; set +a
```

### 3) Generate recap JSON
```bash
go run ./cmd/recap \
  --user dennislee928 \
  --year 2025 \
  --out ./web/recap_2025.json
```

### 4) Render HTML + PNG cards
```bash
cd web
npm install
# Install browser for Playwright
npx playwright install chromium
npm run render
```

Outputs:
- `web/out/report.html` (shareable report)
- `web/out/cards/card-01.png` ... (shareable cards)

## Useful flags
- `--skip-growth` : skip stars/forks gained calculation (faster, fewer API calls)
- `--max-search 1000` : cap GraphQL search results (GitHub search has practical limits)

## Troubleshooting
- If **completion counts look low**, ensure the token belongs to the same account and includes `repo` to access private items.
- If you hit rate limits, re-run with `--skip-growth`.

## License
MIT
