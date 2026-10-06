DROP TRIGGER IF EXISTS trg_terminals_create_control_state
ON terminals;

DROP FUNCTION IF EXISTS create_terminal_control_state();

DROP TABLE IF EXISTS terminal_control_state;

DROP TYPE IF EXISTS terminal_lock_state;