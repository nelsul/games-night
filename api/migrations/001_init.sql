-- Games Night initial schema: people, login, and the rooms where games are played.

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email       citext NOT NULL UNIQUE,
    name        text NOT NULL DEFAULT '',
    avatar_key  text,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE login_codes (
    email       citext PRIMARY KEY,
    code_hash   text NOT NULL,
    attempts    int NOT NULL DEFAULT 0,
    sent_at     timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL
);

CREATE TABLE sessions (
    token_hash  text PRIMARY KEY,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL
);
CREATE INDEX sessions_user_idx ON sessions(user_id);

-- A room is where a group meets to play one game. Its code is what people type to join.
-- Codes are unique only among open rooms, so finished or closed rooms free their code.
CREATE TABLE rooms (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code        text NOT NULL CHECK (code ~ '^[A-Z2-9]{4}$'),
    game        text NOT NULL,
    host_id     uuid NOT NULL REFERENCES users(id),
    status      text NOT NULL DEFAULT 'lobby' CHECK (status IN ('lobby', 'playing', 'finished', 'closed')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX rooms_open_code_idx ON rooms(code) WHERE status IN ('lobby', 'playing');

CREATE TABLE room_players (
    room_id     uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);
CREATE INDEX room_players_user_idx ON room_players(user_id);
