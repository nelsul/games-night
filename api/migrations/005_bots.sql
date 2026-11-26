-- Bots: house players that fill a room. They can't sign in (no email) and are never
-- friends, searchable or hosts. Named after pioneers of computing.
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ADD COLUMN is_bot boolean NOT NULL DEFAULT false;
ALTER TABLE users ADD CONSTRAINT users_bot_or_email CHECK (is_bot OR email IS NOT NULL);

INSERT INTO users (name, is_bot) VALUES
    ('Ada', true), ('Linus', true), ('Grace', true), ('Alan', true), ('Margaret', true), ('Dennis', true),
    ('Barbara', true), ('Ken', true), ('Radia', true), ('Guido', true), ('Hedy', true);
