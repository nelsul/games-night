# Contributing

Thanks for helping build Games Night! 🎲

## Workflow

`main` is protected: nobody pushes to it directly. Every change goes through a pull request, **Nelson reviews and approves every pull request**, and **merging into `main` deploys to https://gamesnight.nelsul.com automatically**. Nothing merges by itself.

1. Create a branch from `main`: `git switch -c <game-or-hub>/short-name`. Fork the repo first if you don't have write access.
2. Make your change and run it locally (see [Run it locally](#run-it-locally)).
3. Run the checks for what you touched:
   - API: `gofmt -l .` (must print nothing), `go vet ./...` and `go test ./...` in `api/`. The database isolation tests run when `TEST_DATABASE_URL` points at a Postgres where you may create databases, for example the dev one: `TEST_DATABASE_URL=postgres://gamesnight:gamesnight@127.0.0.1:5435/gamesnight go test ./...`
   - website: `npm run check` and `npm run build` in `web/`
4. Open a pull request with a short description in English, following the template. Add screenshots for anything visual, at phone size (about 390 px wide).
5. CI runs two required checks: **`check`** (format, vet, tests, builds) and **`scope`** (below). Nelson reviews; once approved and green, he merges, and the Deploy workflow ships it. Watch it in the Actions tab.

**Pull requests merge by rebase** (the only option; `main` keeps a linear history). Your commits land on `main` one by one, as you wrote them, so:

- make each commit a small, meaningful step that builds and passes the tests, with a clear message in English;
- tidy up before asking for review (`git rebase -i main` to squash "fix typo" commits into the ones they fix);
- keep the branch up to date with `git fetch && git rebase origin/main` (not merge commits), then `git push --force-with-lease`.

## Two kinds of pull requests

Every pull request is exactly one of these. The **`scope`** check works it out from the files you changed and fails anything in between.

| Kind | Title starts with | May change |
|---|---|---|
| **Hub** | `hub: ` | anything outside the games' folders: rooms, lobby, profiles, friends, chat, the site shell, CI, docs |
| **One game** | `<game id>: ` (e.g. `backdoor: `) | only that game's folders (next section) |

A game pull request may not touch the hub or any other game, and a hub pull request may not touch a game. If a change really has to cross that line (for example, changing the interface every game implements), say so in the PR: Nelson can allow it with the `hub+games` label.

`scope` also fails when:

- a migration already on `main` is edited, renamed or deleted (add a new numbered file instead);
- a game's SQL names anything outside its own schema (other than `REFERENCES public.rooms(id)` / `public.users(id)`), uses an unqualified table name, or uses functions, triggers, `DO` blocks, grants or roles;
- hub SQL touches a game's schema;
- a game imports another game's code (Go or web), or the hub's Go package;
- a game doesn't name its creators (next section).

## Adding a game

A game lives entirely in its own folders, named after its **id** (lower case, letters, digits and `_`, e.g. `quiz`). The id is also its Postgres schema and the prefix of its pull request titles.

| Folder | What goes there |
|---|---|
| `api/games/<id>/` | the Go package: rules engine with unit tests, plus the `games.Engine` implementation (see `api/games/games.go`) |
| `api/games/<id>/migrations/` | `001_init.sql`, `002_….sql`…: tables in schema `<id>`, e.g. `CREATE TABLE <id>.matches (…)` |
| `api/games/all/<id>.go` | one line, `import _ "github.com/nelsul/games-night/api/games/<id>"`, so the hub includes the game |
| `web/src/lib/games/<id>/` | the screens; `index.ts` exports the game module (`GameModule` in `web/src/lib/games.ts`): info (with its creators), title, poster art, match screen, optional lobby options |
| `web/static/games/<id>/` | images, sounds and other files, if any |
| `docs/games/<id>/` | rules and notes, if any |

**Creators.** `info.creators` lists the GitHub username of everyone who made the game, without the @: `creators: ['nelsul', 'your-username']`. It is required, takes one or more people, and the game page shows them with a link to their GitHub. CI fails if the list is missing or empty, has a malformed or repeated name, or names an account that doesn't exist on GitHub. Whoever adds real work to a game later can add themselves in a PR for that game.

The hub finds the game by itself (Go `init()` registration and Vite's `import.meta.glob`), so adding one never touches hub files. Backdoor (`api/games/backdoor`, `web/src/lib/games/backdoor`) is the reference: copy its structure.

**How a game's data stays isolated.** The hub creates a Postgres role `game_<id>` that owns schema `<id>` and nothing else, runs the game's migrations as that role, and switches to it around every call into the game. The role can only point foreign keys at `public.rooms(id)` and `public.users(id)`. It can't read the hub's tables (not even emails) or touch another game's schema: the database refuses, whatever the code does. `api/isolation_test.go` proves it in CI.

**What a game gets from the hub:** the room and its players (with names, photos, online status and bots), the host, lobby options stored as JSON (validated by the game), live updates (every player gets their own view of the match from `View`), the room chat, and bots (implement `games.BotPlayer`).

**Rules for games:** the server decides everything and `View` only sends each player what they may see; original names, art and text (familiar mechanics are fine, copied names, characters or art aren't); Portuguese for everything people see; mobile first.

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

- **The hub** (`api/*.go`, `web/src/routes/`, `web/src/lib/` outside `games/<id>/`) owns people, rooms, the lobby, chat, bots and live updates. It never names a specific game.
- **Games** plug in as described in [Adding a game](#adding-a-game).
- **Live updates:** each room has a WebSocket (`/api/rooms/{code}/ws`). The server always sends a full snapshot of the room, never a diff, so a phone that wakes up or reconnects is up to date at once.
- **Secrets stay on the server.** When games have hidden information (like who the hackers are), the server only ever sends each player what that player may see.

## Guidelines

- **Mobile first.** People play in the living room, each on their own phone, on Safari and Chrome. Check 390 px width; nothing should scroll sideways. Buttons at least 44 px. Inputs at least 16 px so iOS doesn't zoom.
- **Portuguese UI** (pt-BR) for everything people see, including API error messages and `aria-label`s. Code, comments, commits, issues, PRs and docs are in English.
- **Two looks:** the hub has a streaming-style look (near-black, red accent, Bebas Neue titles), with tokens in `web/src/app.css`. Each game has its own look; Backdoor's is inspired by phone settings screens and code review tools (dark, green and red status colors, monospace details). Don't use real companies' names, logos or exact brand colors: the game's companies are the made-up **Cudia** and **Samgung**.
- **Original games only.** Mechanics can be familiar, but names, characters, art and text are ours.
- **Database changes** are new files in `api/migrations/` for the hub, or in `api/games/<id>/migrations/` for a game (`002_….sql`, `003_….sql`…). They are embedded in the API and applied in order when it starts. Never edit a migration that is already on `main`: production has run it.
- **Accessibility:** visible focus, labels on inputs, `aria-label` on icon-only buttons, and nothing communicated by color alone.

## Never commit secrets 🔒

- `.env` files are git-ignored. Keep it that way. `.env.example` only has placeholders.
- No real API keys, passwords, tokens, server addresses or anyone's email address in code, tests, screenshots, issues or PRs. Test people use `@example.test`.
- Production secrets live only in GitHub (environment `production`, used by `main` only) and on the server.
- If you leaked something by accident, tell Nelson right away. The key has to be rotated; deleting the commit isn't enough.
