CREATE TABLE mpesa_stk_requests (
    id UUID PRIMARY KEY,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id) ON DELETE RESTRICT,

    branch_id UUID NOT NULL
        REFERENCES branches(id) ON DELETE RESTRICT,

    sale_id UUID NOT NULL
        REFERENCES sales(id) ON DELETE RESTRICT,

    payment_id UUID
        REFERENCES payments(id) ON DELETE RESTRICT,

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

    CONSTRAINT chk_mpesa_stk_status
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

    CONSTRAINT chk_mpesa_stk_amount
        CHECK (amount > 0),

    CONSTRAINT uq_mpesa_stk_checkout_request
        UNIQUE (checkout_request_id)
);

CREATE INDEX idx_mpesa_stk_tenant
    ON mpesa_stk_requests(tenant_id);

CREATE INDEX idx_mpesa_stk_branch
    ON mpesa_stk_requests(branch_id);

CREATE INDEX idx_mpesa_stk_sale
    ON mpesa_stk_requests(sale_id);

CREATE INDEX idx_mpesa_stk_payment
    ON mpesa_stk_requests(payment_id);

CREATE INDEX idx_mpesa_stk_status
    ON mpesa_stk_requests(status);

CREATE INDEX idx_mpesa_stk_created_at
    ON mpesa_stk_requests(created_at);

CREATE INDEX idx_mpesa_stk_merchant_request
    ON mpesa_stk_requests(merchant_request_id);

CREATE OR REPLACE FUNCTION update_mpesa_stk_requests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mpesa_stk_requests_updated_at
BEFORE UPDATE ON mpesa_stk_requests
FOR EACH ROW
EXECUTE FUNCTION update_mpesa_stk_requests_updated_at();