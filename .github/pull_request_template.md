<!--
Title: start with "hub: " for a hub change, or "<game id>: " for a change to one game
(for example "backdoor: faster reveal"). The "scope" check enforces it.
-->

## Type

- [ ] **Hub** (rooms, lobby, profiles, friends, chat, site shell, CI, docs)
- [ ] **One game**: `<game id>` (only files in that game's folders)

## What

<!-- What changes for the people playing, in a few lines. -->

## Tested locally

<!-- How you ran it (local Postgres, test users at @example.test, bots), and screenshots at ~390 px for anything visual. -->

- [ ] `go test ./...` and `npm run check` pass
- [ ] Migrations are new files only (none edited)
- [ ] A new game lists its creators (GitHub usernames) in `info.creators`
