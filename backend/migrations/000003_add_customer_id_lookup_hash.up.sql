
ALTER TABLE customers
ADD COLUMN IF NOT EXISTS id_number_lookup_hash CHAR(64);

CREATE INDEX IF NOT EXISTS idx_customers_branch_id_lookup_hash
ON customers (branch_id, id_number_lookup_hash);
