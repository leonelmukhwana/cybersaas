-- ============================================================
-- TERMINAL HEALTH
-- Latest health state reported by each terminal.
-- ============================================================

CREATE TABLE terminal_health (
    terminal_id UUID PRIMARY KEY
        REFERENCES terminals(id)
        ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL DEFAULT 'healthy',

    network_adapter_available BOOLEAN NOT NULL DEFAULT FALSE,

    internet_available BOOLEAN NOT NULL DEFAULT FALSE,

    dns_available BOOLEAN NOT NULL DEFAULT FALSE,

    api_available BOOLEAN NOT NULL DEFAULT FALSE,

    database_available BOOLEAN NOT NULL DEFAULT FALSE,

    disk_available BOOLEAN NOT NULL DEFAULT FALSE,

    memory_available BOOLEAN NOT NULL DEFAULT FALSE,

    printer_available BOOLEAN NOT NULL DEFAULT FALSE,

    offline_queue_count INTEGER NOT NULL DEFAULT 0,

    uptime_seconds BIGINT NOT NULL DEFAULT 0,

    network_issue TEXT NOT NULL DEFAULT '',

    issues JSONB NOT NULL DEFAULT '[]'::jsonb,

    last_health_check TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_terminal_health_status
        CHECK (
            status IN (
                'healthy',
                'attention',
                'offline'
            )
        ),

    CONSTRAINT chk_terminal_health_queue
        CHECK (offline_queue_count >= 0),

    CONSTRAINT chk_terminal_health_uptime
        CHECK (uptime_seconds >= 0)
);


-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_terminal_health_status
    ON terminal_health(status);

CREATE INDEX idx_terminal_health_last_seen
    ON terminal_health(last_seen_at);


-- ============================================================
-- UPDATED_AT TRIGGER
-- ============================================================

CREATE TRIGGER trg_terminal_health_updated_at
BEFORE UPDATE ON terminal_health
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();