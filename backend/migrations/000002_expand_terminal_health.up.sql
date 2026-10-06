

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS cpu_usage_percent NUMERIC(5,2)
    NOT NULL DEFAULT 0.00;

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_cpu_usage
CHECK (
    cpu_usage_percent >= 0
    AND cpu_usage_percent <= 100
);


-- ============================================================
-- MEMORY
-- Values are stored in bytes.
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS memory_total_bytes BIGINT
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS memory_used_bytes BIGINT
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS memory_free_bytes BIGINT
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_memory_total
CHECK (memory_total_bytes >= 0);

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_memory_used
CHECK (memory_used_bytes >= 0);

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_memory_free
CHECK (memory_free_bytes >= 0);


-- ============================================================
-- DISK
-- Values are stored in bytes.
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS disk_total_bytes BIGINT
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS disk_free_bytes BIGINT
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_disk_total
CHECK (disk_total_bytes >= 0);

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_disk_free
CHECK (disk_free_bytes >= 0);


-- ============================================================
-- NETWORK
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS lan_available BOOLEAN
    NOT NULL DEFAULT FALSE;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS backend_latency_ms INTEGER
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_latency
CHECK (backend_latency_ms >= 0);


-- ============================================================
-- AUDIO / SOUND
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS sound_available BOOLEAN
    NOT NULL DEFAULT FALSE;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS sound_issue TEXT
    NOT NULL DEFAULT '';


-- ============================================================
-- DRIVERS
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS drivers_available BOOLEAN
    NOT NULL DEFAULT FALSE;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS driver_issue TEXT
    NOT NULL DEFAULT '';


-- ============================================================
-- WINDOWS SECURITY / ANTIVIRUS
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS security_available BOOLEAN
    NOT NULL DEFAULT FALSE;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS antivirus_enabled BOOLEAN
    NOT NULL DEFAULT FALSE;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS security_issue TEXT
    NOT NULL DEFAULT '';


-- ============================================================
-- OPERATING SYSTEM
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS os_name VARCHAR(100)
    NOT NULL DEFAULT '';

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS os_version VARCHAR(100)
    NOT NULL DEFAULT '';

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS os_architecture VARCHAR(50)
    NOT NULL DEFAULT '';


-- ============================================================
-- PC CLIENT
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS client_version VARCHAR(50)
    NOT NULL DEFAULT '';

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS client_status VARCHAR(20)
    NOT NULL DEFAULT 'running';

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_client_status
CHECK (
    client_status IN (
        'running',
        'stopped',
        'error',
        'unknown'
    )
);


-- ============================================================
-- SYNCHRONIZATION
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS sync_pending_count INTEGER
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS sync_failed_count INTEGER
    NOT NULL DEFAULT 0;

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS sync_status VARCHAR(20)
    NOT NULL DEFAULT 'synced';

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS last_sync_at TIMESTAMPTZ;

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_sync_pending
CHECK (sync_pending_count >= 0);

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_sync_failed
CHECK (sync_failed_count >= 0);

ALTER TABLE terminal_health
ADD CONSTRAINT chk_terminal_health_sync_status
CHECK (
    sync_status IN (
        'synced',
        'pending',
        'failed',
        'offline'
    )
);


-- ============================================================
-- HEALTH MESSAGE
-- Human-readable summary generated by the client/server.
-- ============================================================

ALTER TABLE terminal_health
ADD COLUMN IF NOT EXISTS health_message TEXT
    NOT NULL DEFAULT '';


-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_terminal_health_client_status
    ON terminal_health(client_status);

CREATE INDEX IF NOT EXISTS idx_terminal_health_sync_status
    ON terminal_health(sync_status);

CREATE INDEX IF NOT EXISTS idx_terminal_health_updated_at
    ON terminal_health(updated_at);