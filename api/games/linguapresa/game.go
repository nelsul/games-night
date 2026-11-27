package linguapresa

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nelsul/games-night/api/games"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// UserMessage makes rule errors reach players as they are.
func (e *RuleError) UserMessage() string { return e.Msg }

func init() { games.Register(New()) }

// Game plugs Língua Presa into the hub and keeps matches in the linguapresa schema.
type Game struct{}

func New() *Game { return &Game{} }

func (*Game) Info() games.Info {
	return games.Info{ID: "linguapresa", Name: "Língua Presa", MinPlayers: MinPlayers, MaxPlayers: MaxPlayers, Available: true}
}

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

func newRng() *rand.Rand { return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())) }

func (*Game) Start(ctx context.Context, tx pgx.Tx, roomID string, players []string, options json.RawMessage, now time.Time) error {
	o, err := parseOptions(options)
	if err != nil {
		return err
	}
	s, err := Setup(players, o, newRng(), now)
	if err != nil {
		return err
	}
	state, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE linguapresa.matches SET closed_at = now() WHERE room_id = $1 AND closed_at IS NULL`, roomID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO linguapresa.matches (room_id, state) VALUES ($1, $2)`, roomID, state)
	return err
}

func (*Game) Act(ctx context.Context, tx pgx.Tx, roomID, userID string, raw json.RawMessage, now time.Time) error {
	var a Action
	if err := json.Unmarshal(raw, &a); err != nil {
		return ruleErr("Ação inválida.")
	}
	m, err := load(ctx, tx, roomID)
	if err != nil {
		return err
	}
	if err := Apply(&m.state, userID, a, now, newRng()); err != nil {
		return err
	}
	return m.save(ctx, tx, userID, a)
}

// BotStep lets one bot that has something to do act, in random order.
func (*Game) BotStep(ctx context.Context, tx pgx.Tx, roomID string, bots []string, now time.Time) (bool, error) {
	m, err := load(ctx, tx, roomID)
	if errors.Is(err, errNoMatch) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	rng := newRng()
	order := slices.Clone(bots)
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	for _, bot := range order {
		a := BotAction(&m.state, bot, now, rng)
		if a == nil {
			continue
		}
		if err := Apply(&m.state, bot, *a, now, rng); err != nil {
			return false, fmt.Errorf("bot %s %s: %w", bot, a.Type, err)
		}
		return true, m.save(ctx, tx, bot, *a)
	}
	return false, nil
}

var errNoMatch = ruleErr("Não há partida nessa sala.")

type match struct {
	id      string
	version int
	state   State
}

func load(ctx context.Context, tx pgx.Tx, roomID string) (*match, error) {
	var m match
	var state []byte
	err := tx.QueryRow(ctx, `SELECT id, version, state FROM linguapresa.matches WHERE room_id = $1 AND closed_at IS NULL FOR UPDATE`, roomID).
		Scan(&m.id, &m.version, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errNoMatch
	} else if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(state, &m.state); err != nil {
		return nil, fmt.Errorf("decode match %s: %w", m.id, err)
	}
	return &m, nil
}

func (m *match) save(ctx context.Context, tx pgx.Tx, userID string, a Action) error {
	next, err := json.Marshal(&m.state)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE linguapresa.matches SET state = $2, version = version + 1, updated_at = now() WHERE id = $1`, m.id, next); err != nil {
		return err
	}
	action, _ := json.Marshal(a)
	_, err = tx.Exec(ctx, `INSERT INTO linguapresa.actions (match_id, seq, user_id, action) VALUES ($1, $2, $3, $4)`, m.id, m.version+1, userID, action)
	return err
}

func (*Game) View(ctx context.Context, q games.Querier, roomID, userID string) (any, error) {
	var state []byte
	err := q.QueryRow(ctx, `SELECT state FROM linguapresa.matches WHERE room_id = $1 AND closed_at IS NULL`, roomID).Scan(&state)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(state, &s); err != nil {
		return nil, err
	}
	return ViewFor(&s, userID, time.Now()), nil
}

func (*Game) Over(ctx context.Context, q games.Querier, roomID string) (bool, error) {
	var over bool
	err := q.QueryRow(ctx, `SELECT state->>'phase' = 'over' FROM linguapresa.matches WHERE room_id = $1 AND closed_at IS NULL`, roomID).Scan(&over)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return over, err
}

func (*Game) Stop(ctx context.Context, tx pgx.Tx, roomID string) error {
	_, err := tx.Exec(ctx, `UPDATE linguapresa.matches SET closed_at = now() WHERE room_id = $1 AND closed_at IS NULL`, roomID)
	return err
}
