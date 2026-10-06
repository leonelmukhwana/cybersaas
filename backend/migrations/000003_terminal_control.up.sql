-- Terminal lock/unlock state is independent of SaaS subscription status.
-- A terminal keeps its last known state when the SaaS subscription expires
-- or when the terminal temporarily loses internet connectivity.

CREATE TYPE terminal_lock_state AS ENUM (
    'unlocked',
    'locked'
);

CREATE TABLE terminal_control_state (
    terminal_id UUID PRIMARY KEY
        REFERENCES terminals(id) ON DELETE CASCADE,

    desired_state terminal_lock_state NOT NULL DEFAULT 'unlocked',

    command_version BIGINT NOT NULL DEFAULT 1,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_terminal_control_state_updated_at
    ON terminal_control_state(updated_at);

CREATE OR REPLACE FUNCTION create_terminal_control_state()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO terminal_control_state (
        terminal_id,
        desired_state,
        command_version
    )
    VALUES (
        NEW.id,
        'unlocked',
        1
    )
    ON CONFLICT (terminal_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_terminals_create_control_state
AFTER INSERT ON terminals
FOR EACH ROW
EXECUTE FUNCTION create_terminal_control_state();