-- ============================================================
-- M-PESA CONFIGURATIONS
--
-- Branch configuration:
--   Used for customer payments to a Cyber Owner.
--
-- Platform configuration:
--   Used only for CyberSaaS subscription payments.
--
-- Secrets are encrypted by the application before storage.
-- ============================================================

CREATE TABLE mpesa_configurations (
    id UUID PRIMARY KEY,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    branch_id UUID NOT NULL
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    provider VARCHAR(50) NOT NULL DEFAULT 'daraja',

    environment VARCHAR(20) NOT NULL DEFAULT 'sandbox',

    business_short_code VARCHAR(50),
    till_number VARCHAR(50),
    paybill_number VARCHAR(50),

    consumer_key_encrypted TEXT,
    consumer_secret_encrypted TEXT,
    passkey_encrypted TEXT,

    account_reference VARCHAR(100),

    callback_url TEXT,

    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(branch_id, provider),

    CHECK (
        environment IN ('sandbox', 'production')
    )
);

CREATE INDEX idx_mpesa_config_tenant
    ON mpesa_configurations(tenant_id);

CREATE INDEX idx_mpesa_config_branch
    ON mpesa_configurations(branch_id);

CREATE INDEX idx_mpesa_config_active
    ON mpesa_configurations(branch_id, active);


-- ============================================================
-- PLATFORM / SAAS M-PESA CONFIGURATION
--
-- This is NOT associated with a tenant or branch.
--
-- It is used for:
--   - SaaS subscriptions
--   - SaaS plan payments
--   - Lifetime payments
-- ============================================================

CREATE TABLE platform_mpesa_configurations (
    id UUID PRIMARY KEY,

    provider VARCHAR(50) NOT NULL DEFAULT 'daraja',

    environment VARCHAR(20) NOT NULL DEFAULT 'sandbox',

    business_short_code VARCHAR(50),
    till_number VARCHAR(50),
    paybill_number VARCHAR(50),

    consumer_key_encrypted TEXT,
    consumer_secret_encrypted TEXT,
    passkey_encrypted TEXT,

    account_reference VARCHAR(100),

    callback_url TEXT,

    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(provider),

    CHECK (
        environment IN ('sandbox', 'production')
    )
);

CREATE INDEX idx_platform_mpesa_active
    ON platform_mpesa_configurations(active);