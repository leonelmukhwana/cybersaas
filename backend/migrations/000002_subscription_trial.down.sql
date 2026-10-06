DROP INDEX IF EXISTS idx_subscriptions_trial;

ALTER TABLE subscriptions
DROP COLUMN IF EXISTS is_trial;