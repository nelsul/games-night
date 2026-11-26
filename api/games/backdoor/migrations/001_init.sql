-- Backdoor's own schema. It only points at the hub's rooms and users.

CREATE TABLE backdoor.matches (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     uuid NOT NULL REFERENCES public.rooms(id) ON DELETE CASCADE,
    state       jsonb NOT NULL,          -- the engine's State, secrets included
    version     int NOT NULL DEFAULT 0,  -- number of actions applied
    winner      text CHECK (winner IN ('devs', 'hackers')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    closed_at   timestamptz              -- set when the room moves on (play again, or ended early)
);
-- A room has at most one current match.
CREATE UNIQUE INDEX matches_current_idx ON backdoor.matches(room_id) WHERE closed_at IS NULL;

-- Every accepted action, in order: enough to replay or audit a match.
CREATE TABLE backdoor.actions (
    match_id    uuid NOT NULL REFERENCES backdoor.matches(id) ON DELETE CASCADE,
    seq         int NOT NULL,
    user_id     uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    action      jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, seq)
);
