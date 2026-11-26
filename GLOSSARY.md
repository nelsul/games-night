# Glossary

Terms used in code, issues and PRs. The site shows the Portuguese word in parentheses.

## Hub

- **Hub**: the part of Games Night shared by every game: sign-in, profiles, rooms, the lobby and live updates.
- **Game**: one entry in the catalog (`api/games.go`), like Backdoor. Has a name and a minimum and maximum number of players.
- **Room** (*sala*): where a group meets to play one game. Identified by its **code**.
- **Code** (*código*): 4 characters from an alphabet without look-alikes (no 0/O, 1/I/L). Unique only among open rooms.
- **Host** (*host*): the person who controls the room (starts the match). The first to join after the host leaves becomes host.
- **Player**: a person in a room. Shown as **online** while they have the room open on some device.
- **Lobby**: a room's state before a match starts (*quem está na sala?*).
- **Room status**: `lobby` → `playing` → `finished`; `closed` when everyone left.
- **Match** (*partida*): one play-through of a game inside a room.
- **Snapshot**: the full state sent over the WebSocket. Clients never merge diffs.

## Backdoor

- **Cudia**: the made-up software company the players work for. **Samgung**: its made-up client.
- **Devs** (*Time Devs*): the majority team. Wins by delivering 3 sprints.
- **Hackers** (*Time Hackers*): the hidden minority, about a third of the table. They know each other. Win with 3 hacked sprints or 5 failed meetings in a row.
- **Badge** (*crachá*): a player's secret role card.
- **Roles**: **Dev**; **Security Analyst** (*Analista de Segurança*, a dev who knows the hackers); **Hacker**; **Social Engineer** (*Engenharia Social*, a hacker who gets one last guess at the Security Analyst if the devs win).
- **Sprint**: one of up to 5 rounds. Ends **delivered** (*deploy*) or **hacked** (*invadido*).
- **Tech lead** (*tech lead da vez*): the player proposing the squad this round. Rotates around the table.
- **Squad**: the players the tech lead proposes for a sprint, shown as a pull request.
- **Review**: everyone's open vote on the squad: **approve** (*Approve*/LGTM) or **request changes**.
- **Failed meeting** (*reunião sem decisão*): a squad that was not approved. Five in a row and the hackers win.
- **Push**: each squad member's secret choice: a clean `git push`, or (hackers only) `git push --backdoor`. One backdoor hacks the sprint.
- **Post-mortem**: the end screen that reveals every badge.
