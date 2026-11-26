-- Lobby options set by the host. The hub stores them as-is; the room's game validates them.
ALTER TABLE rooms ADD COLUMN options jsonb NOT NULL DEFAULT '{}';
