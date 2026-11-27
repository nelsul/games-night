-- Feedback, bug reports and game ideas sent from the site, and the admins who read
-- them. For now admins can only see, close and delete feedback.
ALTER TABLE users ADD COLUMN is_admin boolean NOT NULL DEFAULT false;
UPDATE users SET is_admin = true WHERE handle = 'nelsul' AND NOT is_bot;

CREATE TABLE feedback (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind       text NOT NULL CHECK (kind IN ('feedback', 'bug', 'idea')),
    body       text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    -- NULL when sent anonymously: then nothing ties it to the person who sent it.
    user_id    uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    done_at    timestamptz,
    done_by    uuid REFERENCES users (id) ON DELETE SET NULL
);
CREATE INDEX feedback_by_date ON feedback (created_at DESC);
