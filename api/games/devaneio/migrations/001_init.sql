-- Devaneio's own schema. It only points at the hub's rooms and users.

CREATE TABLE devaneio.matches (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     uuid NOT NULL REFERENCES public.rooms(id) ON DELETE CASCADE,
    state       jsonb NOT NULL,          -- the engine's State, the card on the table included
    version     int NOT NULL DEFAULT 0,  -- number of actions applied
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    closed_at   timestamptz
);
CREATE UNIQUE INDEX matches_current_idx ON devaneio.matches(room_id) WHERE closed_at IS NULL;

CREATE TABLE devaneio.actions (
    match_id    uuid NOT NULL REFERENCES devaneio.matches(id) ON DELETE CASCADE,
    seq         int NOT NULL,
    user_id     uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    action      jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, seq)
);
