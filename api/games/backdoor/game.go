package backdoor

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nelsul/games-night/api/games"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// UserMessage makes rule errors reach players as they are.
func (e *RuleError) UserMessage() string { return e.Msg }

// Game plugs Backdoor into the hub and keeps matches in the backdoor schema.
type Game struct{}

func New() *Game { return &Game{} }

func (*Game) Info() games.Info {
	return games.Info{ID: "backdoor", Name: "Backdoor", MinPlayers: MinPlayers, MaxPlayers: MaxPlayers, Available: true}
}

func (*Game) Schema() string { return "backdoor" }

func (*Game) Migrations() fs.FS {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		panic(err)
	}
	return sub
}

func parseOptions(raw json.RawMessage) (Options, error) {
	var o Options
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &o); err != nil {
			return o, ruleErr("Opções inválidas.")
		}
	}
	return o, nil
}

func (*Game) CheckOptions(raw json.RawMessage) (json.RawMessage, error) {
	o, err := parseOptions(raw)
	if err != nil {
		return nil, err
	}
	if err := o.CheckLobby(); err != nil {
		return nil, err
	}
	return json.Marshal(o)
}

func (*Game) Start(ctx context.Context, tx pgx.Tx, roomID string, players []string, options json.RawMessage, now time.Time) error {
	o, err := parseOptions(options)
	if err != nil {
		return err
	}
	s, err := Setup(players, o, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), now)
	if err != nil {
		return err
	}
	state, err := json.Marshal(s)
	if err != nil {
		return err
	}
	// A previous match left open (finished or not) is closed first.
	if _, err := tx.Exec(ctx, `UPDATE backdoor.matches SET closed_at = now() WHERE room_id = $1 AND closed_at IS NULL`, roomID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO backdoor.matches (room_id, state) VALUES ($1, $2)`, roomID, state)
	return err
}

func (*Game) Act(ctx context.Context, tx pgx.Tx, roomID, userID string, raw json.RawMessage, now time.Time) error {
	var a Action
	if err := json.Unmarshal(raw, &a); err != nil {
		return ruleErr("Ação inválida.")
	}
	var id string
	var version int
	var state []byte
	err := tx.QueryRow(ctx, `SELECT id, version, state FROM backdoor.matches WHERE room_id = $1 AND closed_at IS NULL FOR UPDATE`, roomID).
		Scan(&id, &version, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return ruleErr("Não há partida nessa sala.")
	} else if err != nil {
		return err
	}
	var s State
	if err := json.Unmarshal(state, &s); err != nil {
		return fmt.Errorf("decode match %s: %w", id, err)
	}
	if err := Apply(&s, userID, a, now); err != nil {
		return err
	}
	next, err := json.Marshal(&s)
	if err != nil {
		return err
	}
	var winner *string
	if s.Winner != "" {
		winner = &s.Winner
	}
	if _, err := tx.Exec(ctx, `UPDATE backdoor.matches SET state = $2, version = version + 1, winner = $3, updated_at = now() WHERE id = $1`, id, next, winner); err != nil {
		return err
	}
	action, _ := json.Marshal(a)
	_, err = tx.Exec(ctx, `INSERT INTO backdoor.actions (match_id, seq, user_id, action) VALUES ($1, $2, $3, $4)`, id, version+1, userID, action)
	return err
}

func (*Game) View(ctx context.Context, q games.Querier, roomID, userID string) (any, error) {
	var state []byte
	err := q.QueryRow(ctx, `SELECT state FROM backdoor.matches WHERE room_id = $1 AND closed_at IS NULL`, roomID).Scan(&state)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(state, &s); err != nil {
		return nil, err
	}
	return ViewFor(&s, userID), nil
}

func (*Game) Stop(ctx context.Context, tx pgx.Tx, roomID string) error {
	_, err := tx.Exec(ctx, `UPDATE backdoor.matches SET closed_at = now() WHERE room_id = $1 AND closed_at IS NULL`, roomID)
	return err
}
