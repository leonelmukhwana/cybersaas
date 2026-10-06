CREATE TABLE IF NOT EXISTS terminal_health (
    terminal_id UUID PRIMARY KEY,

    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,

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

    issues JSONB NOT NULL DEFAULT '[]'::jsonb,

    last_health_check TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_terminal_health_terminal
        FOREIGN KEY (terminal_id)
        REFERENCES terminals(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_terminal_health_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_terminal_health_branch
        FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE CASCADE,

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

CREATE INDEX IF NOT EXISTS idx_terminal_health_branch
    ON terminal_health(branch_id);

CREATE INDEX IF NOT EXISTS idx_terminal_health_tenant
    ON terminal_health(tenant_id);

CREATE INDEX IF NOT EXISTS idx_terminal_health_status
    ON terminal_health(status);

CREATE INDEX IF NOT EXISTS idx_terminal_health_last_seen
    ON terminal_health(last_seen_at);