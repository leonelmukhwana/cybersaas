-- ============================================================
-- TERMINAL COMMANDS
-- ============================================================

CREATE TYPE terminal_command_type AS ENUM (
    'restart',
    'shutdown'
);

CREATE TYPE terminal_command_status AS ENUM (
    'pending',
    'executed',
    'failed',
    'cancelled'
);

CREATE TABLE terminal_commands (
    id UUID PRIMARY KEY,
    terminal_id UUID NOT NULL REFERENCES terminals(id) ON DELETE CASCADE,

    command terminal_command_type NOT NULL,
    status terminal_command_status NOT NULL DEFAULT 'pending',

    command_version BIGINT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,

    error_message TEXT
);

CREATE INDEX idx_terminal_commands_pending
    ON terminal_commands (terminal_id, status, command_version);

CREATE INDEX idx_terminal_commands_created_at
    ON terminal_commands (created_at);