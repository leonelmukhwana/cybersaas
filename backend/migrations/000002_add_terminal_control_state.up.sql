DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type
        WHERE typname = 'terminal_lock_state'
    ) THEN
        CREATE TYPE terminal_lock_state AS ENUM (
            'locked',
            'unlocked'
        );
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS terminal_control_state (
    terminal_id UUID PRIMARY KEY
        REFERENCES terminals(id)
        ON DELETE CASCADE,

    desired_state terminal_lock_state NOT NULL
        DEFAULT 'unlocked',

    command_version BIGINT NOT NULL
        DEFAULT 1,

    updated_at TIMESTAMPTZ NOT NULL
        DEFAULT NOW()
);