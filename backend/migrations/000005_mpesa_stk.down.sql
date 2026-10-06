DROP TRIGGER IF EXISTS trg_mpesa_stk_requests_updated_at
ON mpesa_stk_requests;

DROP FUNCTION IF EXISTS update_mpesa_stk_requests_updated_at();

DROP TABLE IF EXISTS mpesa_stk_requests;