-- ============================================================
-- REPORTING / CA COMPLIANCE
-- Migration: 000006
-- ============================================================

-- ------------------------------------------------------------
-- COMPLIANCE REPORT SNAPSHOTS
--
-- Stores an immutable snapshot of a generated compliance
-- report. The source operational records remain in their
-- original tables.
--
-- retention_until is calculated when the snapshot is created
-- and MUST NOT be extended merely because the report is viewed
-- or regenerated.
-- ------------------------------------------------------------

CREATE TABLE compliance_report_snapshots (
    id UUID PRIMARY KEY,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    branch_id UUID
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    generated_by UUID
        REFERENCES users(id)
        ON DELETE RESTRICT,

    report_type VARCHAR(50) NOT NULL,

    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,

    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    retention_until TIMESTAMPTZ NOT NULL,

    snapshot JSONB NOT NULL,

    snapshot_hash VARCHAR(128),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (period_end > period_start),

    CHECK (retention_until >= generated_at)
);

CREATE INDEX idx_compliance_snapshots_tenant
    ON compliance_report_snapshots(tenant_id);

CREATE INDEX idx_compliance_snapshots_branch
    ON compliance_report_snapshots(branch_id);

CREATE INDEX idx_compliance_snapshots_generated_by
    ON compliance_report_snapshots(generated_by);

CREATE INDEX idx_compliance_snapshots_type
    ON compliance_report_snapshots(report_type);

CREATE INDEX idx_compliance_snapshots_period
    ON compliance_report_snapshots(period_start, period_end);

CREATE INDEX idx_compliance_snapshots_retention
    ON compliance_report_snapshots(retention_until);


-- ------------------------------------------------------------
-- PREVENT ACCIDENTAL MODIFICATION OF SNAPSHOT CONTENT
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION prevent_compliance_snapshot_update()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.branch_id IS DISTINCT FROM OLD.branch_id
       OR NEW.generated_by IS DISTINCT FROM OLD.generated_by
       OR NEW.report_type IS DISTINCT FROM OLD.report_type
       OR NEW.period_start IS DISTINCT FROM OLD.period_start
       OR NEW.period_end IS DISTINCT FROM OLD.period_end
       OR NEW.generated_at IS DISTINCT FROM OLD.generated_at
       OR NEW.retention_until IS DISTINCT FROM OLD.retention_until
       OR NEW.snapshot IS DISTINCT FROM OLD.snapshot
       OR NEW.snapshot_hash IS DISTINCT FROM OLD.snapshot_hash
    THEN
        RAISE EXCEPTION
            'Compliance report snapshots are immutable';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_compliance_snapshot_immutable
BEFORE UPDATE ON compliance_report_snapshots
FOR EACH ROW
EXECUTE FUNCTION prevent_compliance_snapshot_update();


-- ------------------------------------------------------------
-- PREVENT DELETING COMPLIANCE SNAPSHOTS BEFORE RETENTION
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION prevent_compliance_snapshot_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.retention_until > NOW() THEN
        RAISE EXCEPTION
            'Compliance report snapshot is retained until %',
            OLD.retention_until;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_compliance_snapshot_retention
BEFORE DELETE ON compliance_report_snapshots
FOR EACH ROW
EXECUTE FUNCTION prevent_compliance_snapshot_delete();


-- ------------------------------------------------------------
-- REPORT GENERATION AUDIT
-- ------------------------------------------------------------

CREATE TABLE report_generation_logs (
    id UUID PRIMARY KEY,

    tenant_id UUID
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    branch_id UUID
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    generated_by UUID
        REFERENCES users(id)
        ON DELETE RESTRICT,

    user_role user_role,

    report_type VARCHAR(50) NOT NULL,

    period_start TIMESTAMPTZ NOT NULL,

    period_end TIMESTAMPTZ NOT NULL,

    format VARCHAR(20) NOT NULL DEFAULT 'json',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (period_end > period_start)
);

CREATE INDEX idx_report_generation_tenant
    ON report_generation_logs(tenant_id);

CREATE INDEX idx_report_generation_branch
    ON report_generation_logs(branch_id);

CREATE INDEX idx_report_generation_user
    ON report_generation_logs(generated_by);

CREATE INDEX idx_report_generation_created
    ON report_generation_logs(created_at);