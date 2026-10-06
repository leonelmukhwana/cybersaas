-- ============================================================
-- SUBSCRIPTION PLANS
-- ============================================================

CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(50) NOT NULL UNIQUE,

    included_branches INTEGER NOT NULL DEFAULT 1,
    included_terminals INTEGER NOT NULL DEFAULT 5,

    extra_branch_rate NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    extra_terminal_rate NUMERIC(12,2) NOT NULL DEFAULT 0.00,

    monthly_price NUMERIC(12,2) NOT NULL DEFAULT 0.00,

    is_lifetime BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subscription_plans_branches_check
        CHECK (included_branches >= 0),

    CONSTRAINT subscription_plans_terminals_check
        CHECK (included_terminals >= 0),

    CONSTRAINT subscription_plans_branch_rate_check
        CHECK (extra_branch_rate >= 0),

    CONSTRAINT subscription_plans_terminal_rate_check
        CHECK (extra_terminal_rate >= 0),

    CONSTRAINT subscription_plans_monthly_price_check
        CHECK (monthly_price >= 0)
);


-- ============================================================
-- DEFAULT PLANS
-- ============================================================

INSERT INTO subscription_plans (
    name,
    included_branches,
    included_terminals,
    extra_branch_rate,
    extra_terminal_rate,
    monthly_price,
    is_lifetime,
    is_active
)
VALUES
(
    'Starter',
    1,
    5,
    1500.00,
    100.00,
    1500.00,
    FALSE,
    TRUE
),
(
    'Pro',
    3,
    20,
    1000.00,
    75.00,
    5000.00,
    FALSE,
    TRUE
),
(
    'Lifetime Enterprise',
    999,
    999,
    0.00,
    0.00,
    150000.00,
    TRUE,
    TRUE
)
ON CONFLICT (name) DO NOTHING;


-- ============================================================
-- SUBSCRIPTIONS
-- ============================================================

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    plan_id UUID NOT NULL
        REFERENCES subscription_plans(id)
        ON DELETE RESTRICT,

    status VARCHAR(20) NOT NULL DEFAULT 'active',

    is_lifetime BOOLEAN NOT NULL DEFAULT FALSE,

    account_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,

    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ,

    peak_branches INTEGER NOT NULL DEFAULT 0,
    peak_terminals INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subscriptions_status_check
        CHECK (
            status IN (
                'active',
                'past_due',
                'expired',
                'cancelled'
            )
        ),

    CONSTRAINT subscriptions_balance_check
        CHECK (account_balance >= 0),

    CONSTRAINT subscriptions_peak_branches_check
        CHECK (peak_branches >= 0),

    CONSTRAINT subscriptions_peak_terminals_check
        CHECK (peak_terminals >= 0),

    CONSTRAINT subscriptions_period_check
        CHECK (
            current_period_end IS NULL
            OR current_period_end > current_period_start
            OR is_lifetime = TRUE
        ),

    CONSTRAINT unique_tenant_subscription
        UNIQUE (tenant_id)
);


CREATE INDEX idx_subscriptions_tenant_status
    ON subscriptions(tenant_id, status);

CREATE INDEX idx_subscriptions_status
    ON subscriptions(status);

CREATE INDEX idx_subscriptions_period_end
    ON subscriptions(current_period_end);


-- ============================================================
-- SUBSCRIPTION PAYMENTS
--
-- Separate from the cyber's normal customer payments.
-- ============================================================

CREATE TABLE subscription_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    subscription_id UUID NOT NULL
        REFERENCES subscriptions(id)
        ON DELETE RESTRICT,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    amount NUMERIC(12,2) NOT NULL,

    phone_number VARCHAR(15) NOT NULL,

    payment_method VARCHAR(20) NOT NULL DEFAULT 'mpesa_stk',

    mpesa_checkout_request_id VARCHAR(100),

    mpesa_receipt_number VARCHAR(100),

    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    failure_reason VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subscription_payments_amount_check
        CHECK (amount > 0),

    CONSTRAINT subscription_payments_method_check
        CHECK (
            payment_method IN (
                'mpesa_stk',
                'manual',
                'bank',
                'other'
            )
        ),

    CONSTRAINT subscription_payments_status_check
        CHECK (
            status IN (
                'pending',
                'completed',
                'failed',
                'cancelled'
            )
        )
);


CREATE UNIQUE INDEX uq_subscription_payment_checkout
    ON subscription_payments(mpesa_checkout_request_id)
    WHERE mpesa_checkout_request_id IS NOT NULL;


CREATE UNIQUE INDEX uq_subscription_payment_receipt
    ON subscription_payments(mpesa_receipt_number)
    WHERE mpesa_receipt_number IS NOT NULL;


CREATE INDEX idx_subscription_payments_tenant
    ON subscription_payments(tenant_id, created_at DESC);

CREATE INDEX idx_subscription_payments_subscription
    ON subscription_payments(subscription_id, created_at DESC);

CREATE INDEX idx_subscription_payments_status
    ON subscription_payments(status);


-- ============================================================
-- SUBSCRIPTION LEDGER
--
-- Financial history.
-- Entries are immutable.
-- ============================================================

CREATE TABLE subscription_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    subscription_id UUID NOT NULL
        REFERENCES subscriptions(id)
        ON DELETE RESTRICT,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    payment_id UUID
        REFERENCES subscription_payments(id)
        ON DELETE RESTRICT,

    entry_type VARCHAR(10) NOT NULL,

    amount NUMERIC(12,2) NOT NULL,

    balance_after NUMERIC(12,2) NOT NULL,

    description VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subscription_ledger_entry_type_check
        CHECK (entry_type IN ('CREDIT', 'DEBIT')),

    CONSTRAINT subscription_ledger_amount_check
        CHECK (amount > 0),

    CONSTRAINT subscription_ledger_balance_check
        CHECK (balance_after >= 0)
);


CREATE INDEX idx_subscription_ledger_subscription
    ON subscription_ledger(subscription_id, created_at DESC);

CREATE INDEX idx_subscription_ledger_tenant
    ON subscription_ledger(tenant_id, created_at DESC);

CREATE INDEX idx_subscription_ledger_payment
    ON subscription_ledger(payment_id);


-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================

CREATE TRIGGER trg_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


CREATE TRIGGER trg_subscription_payments_updated_at
BEFORE UPDATE ON subscription_payments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- ============================================================
-- UPDATE SUBSCRIPTION USAGE PEAKS
--
-- Branch creation/update affects branch peak.
-- Terminal creation/update affects terminal peak.
-- ============================================================

CREATE OR REPLACE FUNCTION update_subscription_usage_peaks()
RETURNS TRIGGER AS $$
DECLARE
    v_tenant_id UUID;
    v_branch_count INTEGER;
    v_terminal_count INTEGER;
BEGIN

    IF TG_TABLE_NAME = 'branches' THEN
        v_tenant_id := NEW.tenant_id;

    ELSIF TG_TABLE_NAME = 'terminals' THEN

        SELECT tenant_id
        INTO v_tenant_id
        FROM branches
        WHERE id = NEW.branch_id;

    END IF;


    IF v_tenant_id IS NULL THEN
        RETURN NEW;
    END IF;


    SELECT COUNT(*)
    INTO v_branch_count
    FROM branches
    WHERE tenant_id = v_tenant_id
      AND status = 'active';


    SELECT COUNT(*)
    INTO v_terminal_count
    FROM terminals t
    INNER JOIN branches b
        ON b.id = t.branch_id
    WHERE b.tenant_id = v_tenant_id
      AND t.status = 'active';


    UPDATE subscriptions
    SET
        peak_branches = GREATEST(
            peak_branches,
            v_branch_count
        ),

        peak_terminals = GREATEST(
            peak_terminals,
            v_terminal_count
        ),

        updated_at = NOW()

    WHERE tenant_id = v_tenant_id;


    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- ============================================================
-- USAGE PEAK TRIGGERS
-- ============================================================

CREATE TRIGGER trg_branches_subscription_usage
AFTER INSERT OR UPDATE OF status
ON branches
FOR EACH ROW
EXECUTE FUNCTION update_subscription_usage_peaks();


CREATE TRIGGER trg_terminals_subscription_usage
AFTER INSERT OR UPDATE OF status
ON terminals
FOR EACH ROW
EXECUTE FUNCTION update_subscription_usage_peaks();


-- ============================================================
-- PREVENT LEDGER MODIFICATION
--
-- Subscription financial ledger entries cannot be edited
-- or deleted after creation.
-- ============================================================

CREATE OR REPLACE FUNCTION prevent_subscription_ledger_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION
        'Subscription ledger entries are immutable';
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_subscription_ledger_immutable
BEFORE UPDATE OR DELETE
ON subscription_ledger
FOR EACH ROW
EXECUTE FUNCTION prevent_subscription_ledger_mutation();


-- ============================================================
-- PROCESS SUBSCRIPTION LEDGER ENTRY
--
-- CREDIT:
--   Adds money to the subscription account balance.
--
-- DEBIT:
--   Deducts money from the subscription account balance.
--
-- The row is locked during the operation to prevent
-- concurrent balance corruption.
-- ============================================================

CREATE OR REPLACE FUNCTION process_subscription_ledger_entry()
RETURNS TRIGGER AS $$
DECLARE
    v_balance NUMERIC(12,2);
BEGIN

    SELECT account_balance
    INTO v_balance
    FROM subscriptions
    WHERE id = NEW.subscription_id
    FOR UPDATE;


    IF NOT FOUND THEN
        RAISE EXCEPTION
            'Subscription % does not exist',
            NEW.subscription_id;
    END IF;


    IF NEW.entry_type = 'CREDIT' THEN

        v_balance := v_balance + NEW.amount;

    ELSIF NEW.entry_type = 'DEBIT' THEN

        IF v_balance < NEW.amount THEN
            RAISE EXCEPTION
                'Insufficient subscription balance';
        END IF;

        v_balance := v_balance - NEW.amount;

    ELSE

        RAISE EXCEPTION
            'Invalid subscription ledger entry type';

    END IF;


    NEW.balance_after := v_balance;


    UPDATE subscriptions
    SET
        account_balance = v_balance,
        updated_at = NOW()
    WHERE id = NEW.subscription_id;


    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_process_subscription_ledger
BEFORE INSERT
ON subscription_ledger
FOR EACH ROW
EXECUTE FUNCTION process_subscription_ledger_entry();


-- ============================================================
-- END SUBSCRIPTIONS
-- ============================================================