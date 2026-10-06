-- 1. Add trial columns with safe defaults
ALTER TABLE subscriptions 
ADD COLUMN IF NOT EXISTS is_trial BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP WITH TIME ZONE;

-- 2. Backfill existing subscriptions:
-- Set trial_ends_at = created_at + 7 days for historical tracking, 
-- but leave is_trial = FALSE so active paying users aren't disrupted.
UPDATE subscriptions 
SET trial_ends_at = created_at + INTERVAL '7 days' 
WHERE trial_ends_at IS NULL;

-- 3. Create a partial index for fast lookup of active trials (used by Go cron/workers)
CREATE INDEX IF NOT EXISTS idx_subscriptions_active_trials 
ON subscriptions (tenant_id, trial_ends_at) 
WHERE is_trial = TRUE;