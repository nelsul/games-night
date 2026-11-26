# AGENTS.md

Guide for AI coding agents working in this repo. Humans: start with [README.md](README.md) and [CONTRIBUTING.md](CONTRIBUTING.md); this file sums up what an agent needs.

## The project

Games Night is a hub of party games played with friends, each person on their own phone. A host creates a room, others join with a 4-letter code, and the game runs live. The first game is **Backdoor**, a social deduction game themed on a software company (the made-up Cudia, building an app for the made-up client Samgung). Live at https://gamesnight.nelsul.com. The owner is Nelson.

| Folder | What | Stack |
|---|---|---|
| `api/` | JSON API under `/api`, one WebSocket per room, and it serves the built web app | Go, chi, pgx, coder/websocket, PostgreSQL 17, S3-compatible storage (Cloudflare R2 in production) |
| `web/` | The site people use | SvelteKit 2 (static SPA), Svelte 5 runes, TypeScript |
| `.github/workflows/` | `ci.yml` (`check`) and `deploy.yml` | GitHub Actions, deploy over SSH to Docker on a VPS |

Key ideas:

- **The hub doesn't know game rules.** It owns people, rooms, the lobby and live updates. Games register in `api/games.go` and `web/src/lib/games.ts`.
- **The server is authoritative.** Phones send actions; the server validates them and decides outcomes.
- **Hidden information never leaves the server.** Each player only receives their own view of the game state. Never send another player's secret role "to hide it in the UI".
- **Every live message is a full snapshot**, so reconnecting (Safari drops sockets when the screen locks) needs no replay.
- The only emails sent are login codes. Photos are optional: without S3 settings the API runs with photos turned off.

Use the terms in [GLOSSARY.md](GLOSSARY.md) (room, host, lobby, match, squad, sprint…) in code, issues and PRs.

## Rules

### Secrets and privacy (most important)

- **Never commit or print secrets**: `.env`, R2 keys, SMTP credentials, the database password, server addresses, SSH keys. The repo is **public**.
- Never put real people's email addresses in code, tests, screenshots, commits, issues or PRs. Test people use `@example.test`.
- Before every push, check the diff and new files for anything secret-looking.

### Workflow

- `main` is protected: work on a branch and open a PR. Never push to `main`.
- **Merging a PR that touches `api/`, `web/`, `Dockerfile`, `docker-compose.yml` or the workflows deploys to production.** Treat a merge as a deploy.
- CI's `check` job must pass: `gofmt`, `go vet`, `go test`, `npm run check`, `npm run build`.
- Merge with squash. Commit messages, PR descriptions and issues are in English.

### Test locally first

- Run changes against the local Docker Postgres and local S3 (`docker-compose.dev.yml`, `./dev.sh`, `npm run dev` in `web/`).
- **Never send real emails while testing.** Without `SMTP_HOST` the login code is printed in the API's output.
- **Never run migrations or test writes against production.** Production applies `api/migrations/*.sql` by itself when the API starts. Add new numbered files; never edit one already on `main`.
- Multiplayer: test with several test people in private windows joined to the same room.

### Language and look

- Everything people see is in **Portuguese (pt-BR)**, API error messages included. Code, comments, commits and docs are in English.
- The hub's look: tokens in `web/src/app.css`, icons in `web/src/lib/components/Icon.svelte`, character art in `web/src/lib/art/`. No emoji in the UI.
- Games are original: no real companies' names, logos or exact brand colors, and no copied names, characters or text from existing games.
- Mobile first: check at ~390 px; nothing should scroll sideways. Buttons at least 44 px.

## Agent skills

### Issue tracker

Issues live in GitHub Issues for `nelsul/games-night`, managed with the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

The default five-role vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`), created as GitHub labels. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
