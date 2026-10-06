CREATE TABLE expenses (
    id UUID PRIMARY KEY,

    tenant_id UUID NOT NULL
        REFERENCES tenants(id)
        ON DELETE RESTRICT,

    branch_id UUID NOT NULL
        REFERENCES branches(id)
        ON DELETE RESTRICT,

    recorded_by UUID NOT NULL
        REFERENCES users(id)
        ON DELETE RESTRICT,

    category VARCHAR(100) NOT NULL,

    description TEXT,

    amount NUMERIC(14,2) NOT NULL
        CHECK (amount > 0),

    currency CHAR(3) NOT NULL DEFAULT 'KES',

    payment_method VARCHAR(50) NOT NULL
        CHECK (
            payment_method IN (
                'cash',
                'mpesa',
                'bank',
                'card',
                'other'
            )
        ),

    reference VARCHAR(255),

    expense_date DATE NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'recorded'
        CHECK (
            status IN (
                'recorded',
                'voided'
            )
        ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expenses_tenant
    ON expenses (tenant_id);

CREATE INDEX idx_expenses_branch
    ON expenses (branch_id);

CREATE INDEX idx_expenses_branch_date
    ON expenses (branch_id, expense_date);

CREATE INDEX idx_expenses_tenant_date
    ON expenses (tenant_id, expense_date);

CREATE INDEX idx_expenses_recorded_by
    ON expenses (recorded_by);