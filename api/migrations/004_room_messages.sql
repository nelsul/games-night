-- The room chat, shared by the lobby and every match played in the room.
CREATE TABLE room_messages (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    room_id     uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body        text NOT NULL CHECK (length(body) BETWEEN 1 AND 300),
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX room_messages_room_idx ON room_messages(room_id, id);
