CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    token_hash VARCHAR(64) NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_password_reset_token_hash
ON password_reset_tokens(token_hash);

CREATE INDEX idx_password_reset_tokens_user
ON password_reset_tokens(user_id);

CREATE INDEX idx_password_reset_tokens_expiry
ON password_reset_tokens(expires_at);