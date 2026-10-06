ALTER TABLE sessions
ADD COLUMN IF NOT EXISTS paused_at TIMESTAMPTZ NULL;

ALTER TABLE sessions
ADD COLUMN IF NOT EXISTS total_paused_seconds BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_sessions_active_paused
ON sessions (terminal_id, status, paused_at);