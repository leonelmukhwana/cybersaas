-- ============================================================
-- SUBSCRIPTION M-PESA STK REQUESTS
--
-- Used ONLY for CyberSaaS subscription payments.
-- This is deliberately separate from mpesa_stk_requests,
-- which belongs to cyber-customer sales.
-- ============================================================

CREATE TABLE subscription_mpesa_stk_requests (
    id UUID PRIMARY KEY,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    subscription_id UUID NOT NULL
        REFERENCES subscriptions(id)
        ON DELETE RESTRICT,

    subscription_payment_id UUID NOT NULL
        REFERENCES subscription_payments(id)
        ON DELETE RESTRICT,

    phone_number VARCHAR(20) NOT NULL,

    amount NUMERIC(12,2) NOT NULL,

    account_reference VARCHAR(100) NOT NULL,

    transaction_description VARCHAR(255) NOT NULL,

    merchant_request_id VARCHAR(100),

    checkout_request_id VARCHAR(100),

    response_code VARCHAR(20),

    response_description TEXT,

    customer_message TEXT,

    status VARCHAR(30) NOT NULL DEFAULT 'pending',

    result_code VARCHAR(20),

    result_description TEXT,

    mpesa_receipt_number VARCHAR(100),

    transaction_date TIMESTAMPTZ,

    callback_received_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_subscription_mpesa_stk_amount
        CHECK (amount > 0),

    CONSTRAINT chk_subscription_mpesa_stk_status
        CHECK (
            status IN (
                'pending',
                'accepted',
                'completed',
                'failed',
                'cancelled',
                'timeout'
            )
        ),

    CONSTRAINT uq_subscription_mpesa_stk_payment
        UNIQUE (subscription_payment_id),

    CONSTRAINT uq_subscription_mpesa_stk_checkout
        UNIQUE (checkout_request_id),

    CONSTRAINT uq_subscription_mpesa_stk_receipt
        UNIQUE (mpesa_receipt_number)
);

CREATE INDEX idx_subscription_mpesa_stk_tenant
    ON subscription_mpesa_stk_requests(tenant_id);

CREATE INDEX idx_subscription_mpesa_stk_subscription
    ON subscription_mpesa_stk_requests(subscription_id);

CREATE INDEX idx_subscription_mpesa_stk_payment
    ON subscription_mpesa_stk_requests(subscription_payment_id);

CREATE INDEX idx_subscription_mpesa_stk_status
    ON subscription_mpesa_stk_requests(status);

CREATE INDEX idx_subscription_mpesa_stk_created
    ON subscription_mpesa_stk_requests(created_at);

CREATE INDEX idx_subscription_mpesa_stk_merchant
    ON subscription_mpesa_stk_requests(merchant_request_id);

CREATE OR REPLACE FUNCTION update_subscription_mpesa_stk_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_subscription_mpesa_stk_updated_at
BEFORE UPDATE ON subscription_mpesa_stk_requests
FOR EACH ROW
EXECUTE FUNCTION update_subscription_mpesa_stk_updated_at();