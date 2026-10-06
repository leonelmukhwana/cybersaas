DROP TRIGGER IF EXISTS trg_process_subscription_ledger
ON subscription_ledger;

DROP TRIGGER IF EXISTS trg_subscription_ledger_immutable
ON subscription_ledger;

DROP TRIGGER IF EXISTS trg_branches_subscription_usage
ON branches;

DROP TRIGGER IF EXISTS trg_terminals_subscription_usage
ON terminals;

DROP TRIGGER IF EXISTS trg_subscription_payments_updated_at
ON subscription_payments;

DROP TRIGGER IF EXISTS trg_subscriptions_updated_at
ON subscriptions;


DROP FUNCTION IF EXISTS process_subscription_ledger_entry();
DROP FUNCTION IF EXISTS prevent_subscription_ledger_mutation();
DROP FUNCTION IF EXISTS update_subscription_usage_peaks();


DROP TABLE IF EXISTS subscription_ledger;
DROP TABLE IF EXISTS subscription_payments;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS subscription_plans;