# Glossary

Terms used in code, issues and PRs. The site shows the Portuguese word in parentheses.

## Hub

- **Handle** (*@*): a person's unique name for being found, 3 to 20 lower-case letters, digits, `.` or `_`. Required before using the site.
- **Friend** (*amigo*): someone who accepted your **friend request** (*pedido de amizade*), or whose request you accepted. Only friends can be invited to rooms.
- **Invite** (*convite*): a friend asking you to join their room. Joining the room accepts it; it disappears once the match starts.
- **Inbox**: your pending friend requests and invites. The site keeps it fresh over your own WebSocket (`/api/me/ws`), which only says "refetch".
- **Hub**: the part of Games Night shared by every game: sign-in, profiles, rooms, the lobby and live updates.
- **Game**: one entry in the catalog (`api/games.go`), like Backdoor. Has a name and a minimum and maximum number of players.
- **Room** (*sala*): where a group meets to play one game. Identified by its **code**.
- **Code** (*código*): 4 characters from an alphabet without look-alikes (no 0/O, 1/I/L). Unique only among open rooms.
- **Host** (*host*): the person who controls the room (starts the match). The first to join after the host leaves becomes host.
- **Player**: a person in a room. Shown as **online** while they have the room open on some device.
- **Lobby**: a room's state before a match starts (*quem está na sala?*).
- **Room status**: `lobby` → `playing` → `finished`; `closed` when everyone left.
- **Match** (*partida*): one play-through of a game inside a room.
- **Room chat** (*chat da sala*): messages everyone in the room sees, in the lobby and during matches (`room_messages`).
- **Snapshot**: the full state sent over the WebSocket. Clients never merge diffs.
- **Feedback** (*feedback*): a message someone sends from the site, of one of three kinds: **feedback**, **bug** or **game idea** (*ideia de jogo*). Sent with the person's name or **anonymously** (*anônimo*); anonymous feedback stores no user at all. Only admins read it.
- **Admin**: a person with `users.is_admin`. For now admins can only see feedback, mark it **done** (*concluído*) or delete it. The first admin is @nelsul (migration `007_feedback.sql`).

## Backdoor

- **Cudia**: the made-up software company the players work for. **Samgung**: its made-up client.
- **Devs** (*Time Devs*): the majority team. Wins by delivering 3 sprints, unless the Social Engineer then finds the Security Analyst.
- **Hackers** (*Time Hackers*): the hidden minority, about 40% of the table (2 with 5–6 players, 3 with 7–9, 4 with 10–11, 5 with 12–13, 6 with 14–15, 7 with 16). They know each other (except the Freelancer). Win with 3 rollbacks, 5 failed meetings in a row, or the Social Engineer's correct guess.
- **Badge** (*crachá*): a player's secret role card.
- **Roles**, always in play: **Dev**; **Security Analyst** (*Analista de Segurança*, a dev who sees the hackers except the Zero-day); **Hacker**; **Social Engineer** (*Engenharia Social*, a hacker who gets one last guess at the Security Analyst if the devs ship 3 sprints).
- **Optional roles** (host options): **Product Owner** (a dev who sees the Security Analyst, and the Deepfake as a second candidate when it's in play); **Zero-day** (a hacker the Security Analyst can't see, 7+ players); **Deepfake** (a hacker the Product Owner sees as a possible Security Analyst; only with the Product Owner); **Freelancer** (a hacker who doesn't know the other hackers and whom they don't know, 7+ players). Special hackers replace regular ones, and every role that is on is always dealt: a match only starts when the table has a hacker seat for each hacker role that is on (the Social Engineer always takes one).
- **Sprint**: one of up to 5 rounds. Ends **Deploy** or **Rollback**. With 7+ players, sprint 4 needs 2 backdoors to roll back; with 13+, sprint 5 too.
- **Scrum Master**: the player proposing the squad this round. Rotates around the table.
- **Squad**: the players the Scrum Master proposes for a sprint, shown as a pull request.
- **Discussion timer**: optional advisory countdown while the Scrum Master builds the squad.
- **Review**: everyone's open vote on the squad: **approve** (*Approve*/LGTM) or **request changes**.
- **Failed meeting** (*reunião sem decisão*): a squad that was not approved. Five in a row and the hackers win.
- **Push**: each squad member's secret choice: a clean `git push`, or (hackers only) `git push --backdoor`. One backdoor rolls the sprint back.
- **Reveal** (*conferir os commits*): after the pushes, the Scrum Master opens the sprint's commits one at a time, shuffled, so the table sees each clean push or backdoor but never who made it. The sprint is recorded when the Scrum Master closes it.
- **Match** data lives in the game's own Postgres schema (`backdoor.matches`, `backdoor.actions`), never in the hub's tables.
- **Post-mortem**: the end screen that reveals every badge.
