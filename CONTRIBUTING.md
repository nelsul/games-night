# Contributing

Thanks for helping build Games Night! 🎲

## Workflow

`main` is protected: nobody pushes to it directly, Nelson included. Every change goes through a pull request, and **merging into `main` deploys to https://gamesnight.nelsul.com automatically**.

1. Create a branch from `main`: `git switch -c feature/short-name`. Fork the repo first if you don't have write access.
2. Make your change and run it locally (see [Run it locally](#run-it-locally)).
3. Run the checks for what you touched:
   - API: `gofmt -l .` (must print nothing), `go vet ./...` and `go test ./...` in `api/`
   - website: `npm run check` and `npm run build` in `web/`

   CI runs all of them on your PR as the `check` job, and it must pass before the PR can be merged.
4. Open a pull request with a short description in English. Add screenshots for anything visual, at phone size (about 390 px wide).
5. Once it's merged, the Deploy workflow builds the image, ships it to the server and runs a smoke test. Watch it in the Actions tab.

Ideas and bugs go in [GitHub Issues](https://github.com/nelsul/games-night/issues).

## Run it locally

You need Docker, Go (the version in `api/go.mod`) and Node 22+.

```bash
docker compose -f docker-compose.dev.yml up -d   # Postgres on :5435, a local S3 (SeaweedFS) on :8334
./dev.sh                                          # API on :8090; migrations run on start
cd web && npm install && npm run dev              # site on http://localhost:5175
```

- Sign in with a test address such as `ana@example.test`. No email is sent in development: the login code is printed in the `./dev.sh` terminal.
- To play with several people, sign in as other test people in private windows (or other browsers) and join the same room by its code.
- Photos go to the local S3, never to the production bucket.
- The ports differ from cotinha's, so both projects can run side by side.

## How the code is organized

- **The hub** (`api/rooms.go`, `api/hub.go`, `web/src/routes/`) owns people, rooms, the lobby and live updates. It doesn't know any game's rules.
- **A game** is a Go package under `api/games/<game>/` implementing `games.Engine` (`api/games/games.go`): a pure rules engine with unit tests, plus storage in **its own Postgres schema** (`<game>.*`) with its own migrations in `api/games/<game>/migrations/`, tracked in `<game>.schema_migrations`. Register it in `api/games.go`. Its screens live in `web/src/lib/games/<game>/`, and its copy for the hub in `web/src/lib/games.ts`. The hub never reads a game's tables, and a game only points at the hub's `rooms` and `users`.
- **Live updates:** each room has a WebSocket (`/api/rooms/{code}/ws`). The server always sends a full snapshot of the room, never a diff, so a phone that wakes up or reconnects is up to date at once.
- **Secrets stay on the server.** When games have hidden information (like who the hackers are), the server only ever sends each player what that player may see.

## Guidelines

- **Mobile first.** People play in the living room, each on their own phone, on Safari and Chrome. Check 390 px width; nothing should scroll sideways. Buttons at least 44 px. Inputs at least 16 px so iOS doesn't zoom.
- **Portuguese UI** (pt-BR) for everything people see, including API error messages and `aria-label`s. Code, comments, commits, issues, PRs and docs are in English.
- **Two looks:** the hub has a streaming-style look (near-black, red accent, Bebas Neue titles), with tokens in `web/src/app.css`. Each game has its own look; Backdoor's is inspired by phone settings screens and code review tools (dark, green and red status colors, monospace details). Don't use real companies' names, logos or exact brand colors: the game's companies are the made-up **Cudia** and **Samgung**.
- **Original games only.** Mechanics can be familiar, but names, characters, art and text are ours.
- **Database changes** are new files in `api/migrations/` for the hub, or in `api/games/<game>/migrations/` for a game (`002_….sql`, `003_….sql`…). They are embedded in the API and applied in order when it starts. Never edit a migration that is already on `main`: production has run it.
- **Accessibility:** visible focus, labels on inputs, `aria-label` on icon-only buttons, and nothing communicated by color alone.

## Never commit secrets 🔒

- `.env` files are git-ignored. Keep it that way. `.env.example` only has placeholders.
- No real API keys, passwords, tokens, server addresses or anyone's email address in code, tests, screenshots, issues or PRs. Test people use `@example.test`.
- Production secrets live only in GitHub (environment `production`, used by `main` only) and on the server.
- If you leaked something by accident, tell Nelson right away. The key has to be rotated; deleting the commit isn't enough.
