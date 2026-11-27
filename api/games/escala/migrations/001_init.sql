-- Escala's own schema. It only points at the hub's rooms and users.

CREATE TABLE escala.matches (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     uuid NOT NULL REFERENCES public.rooms(id) ON DELETE CASCADE,
    state       jsonb NOT NULL,          -- the engine's State, secret numbers included
    version     int NOT NULL DEFAULT 0,  -- number of actions applied
    score       int,                     -- rounds sorted right; null while playing (a team result)
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    closed_at   timestamptz
);
CREATE UNIQUE INDEX matches_current_idx ON escala.matches(room_id) WHERE closed_at IS NULL;

CREATE TABLE escala.actions (
    match_id    uuid NOT NULL REFERENCES escala.matches(id) ON DELETE CASCADE,
    seq         int NOT NULL,
    user_id     uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    action      jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, seq)
);
