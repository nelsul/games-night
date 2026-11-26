-- A room whose match just ended is "finished": people can still see the post-mortem,
-- leave, or play again. Its code stays reserved until the room closes.
DROP INDEX rooms_open_code_idx;
CREATE UNIQUE INDEX rooms_open_code_idx ON rooms(code) WHERE status IN ('lobby', 'playing', 'finished');
