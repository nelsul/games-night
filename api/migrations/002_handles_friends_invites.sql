-- Unique @handles, friendships and inviting friends to a room.

ALTER TABLE users ADD COLUMN handle citext UNIQUE CHECK (handle ~ '^[a-z0-9._]{3,20}$');

-- Friendships are stored once per pair, with user_a < user_b.
CREATE TABLE friendships (
    user_a      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_a, user_b),
    CHECK (user_a < user_b)
);
CREATE INDEX friendships_b_idx ON friendships(user_b);

CREATE TABLE friend_requests (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    from_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_id         uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        text NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    responded_at  timestamptz,
    CHECK (from_id <> to_id)
);
CREATE UNIQUE INDEX friend_requests_pending_idx ON friend_requests(from_id, to_id) WHERE status = 'pending';
CREATE INDEX friend_requests_to_idx ON friend_requests(to_id, status);

-- A friend asking someone to join their room. Joining the room accepts it.
CREATE TABLE room_invites (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    from_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    CHECK (from_id <> to_id)
);
CREATE UNIQUE INDEX room_invites_pending_idx ON room_invites(room_id, to_id) WHERE status = 'pending';
CREATE INDEX room_invites_to_idx ON room_invites(to_id, status);
