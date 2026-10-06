package terminal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ============================================================
// SECRET HELPERS
// ============================================================

func generateSecret(bytesLength int) (string, error) {
	buf := make([]byte, bytesLength)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func hashSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// ============================================================
// BRANCH VALIDATION
// ============================================================

func (r *Repository) VerifyBranchBelongsToTenant(
	ctx context.Context,
	tenantID string,
	branchID string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE id = $1
			  AND tenant_id = $2
			  AND status = 'active'
		)
		`,
		branchID,
		tenantID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("branch not found or inactive")
	}

	return nil
}

// ============================================================
// LICENCE KEYS
// ============================================================

func (r *Repository) CreateLicenceKey(
	ctx context.Context,
	tenantID string,
	branchID string,
	createdBy string,
	expiresAt time.Time,
) (*LicenceKey, string, error) {
	plainKey, err := generateSecret(24)
	if err != nil {
		return nil, "", err
	}

	hash := hashSecret(plainKey)
	id := uuid.New().String()

	var key LicenceKey

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO licence_keys (
			id,
			tenant_id,
			branch_id,
			licence_key_hash,
			created_by,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			tenant_id,
			branch_id,
			created_by,
			expires_at,
			used_at,
			revoked_at,
			created_at
		`,
		id,
		tenantID,
		branchID,
		hash,
		createdBy,
		expiresAt,
	).Scan(
		&key.ID,
		&key.TenantID,
		&key.BranchID,
		&key.CreatedBy,
		&key.ExpiresAt,
		&key.UsedAt,
		&key.RevokedAt,
		&key.CreatedAt,
	)

	if err != nil {
		return nil, "", err
	}

	return &key, plainKey, nil
}

func (r *Repository) ListLicenceKeys(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]LicenceKey, int64, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			lk.id,
			lk.tenant_id,
			lk.branch_id,
			b.name,
			lk.created_by,
			lk.expires_at,
			lk.used_at,
			lk.revoked_at,
			lk.created_at
		FROM licence_keys lk
		INNER JOIN branches b
			ON b.id = lk.branch_id
		   AND b.tenant_id = lk.tenant_id
		WHERE lk.tenant_id = $1
		ORDER BY lk.created_at DESC
		LIMIT $2 OFFSET $3
		`,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	keys := make([]LicenceKey, 0)

	for rows.Next() {
		var key LicenceKey

		if err := rows.Scan(
			&key.ID,
			&key.TenantID,
			&key.BranchID,
			&key.BranchName,
			&key.CreatedBy,
			&key.ExpiresAt,
			&key.UsedAt,
			&key.RevokedAt,
			&key.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64

	err = r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM licence_keys
		WHERE tenant_id = $1
		`,
		tenantID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *Repository) RevokeLicenceKey(
	ctx context.Context,
	tenantID string,
	keyID string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE licence_keys
		SET revoked_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		  AND used_at IS NULL
		  AND revoked_at IS NULL
		`,
		keyID,
		tenantID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New(
			"licence key not found, already used, or already revoked",
		)
	}

	return nil
}

// ============================================================
// TERMINALS
// ============================================================

func (r *Repository) GetTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
) (*Terminal, error) {
	var terminal Terminal

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			t.id,
			t.tenant_id,
			t.branch_id,
			b.name,
			t.terminal_code,
			t.machine_name,
			t.device_identifier,
			t.status,
			t.registered_at,
			t.last_seen_at,
			t.disabled_at,
			t.created_at,
			t.updated_at
		FROM terminals t
		INNER JOIN branches b
			ON b.id = t.branch_id
		   AND b.tenant_id = t.tenant_id
		WHERE t.id = $1
		  AND t.tenant_id = $2
		`,
		terminalID,
		tenantID,
	).Scan(
		&terminal.ID,
		&terminal.TenantID,
		&terminal.BranchID,
		&terminal.BranchName,
		&terminal.TerminalCode,
		&terminal.MachineName,
		&terminal.DeviceID,
		&terminal.Status,
		&terminal.RegisteredAt,
		&terminal.LastSeenAt,
		&terminal.DisabledAt,
		&terminal.CreatedAt,
		&terminal.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("terminal not found")
		}

		return nil, err
	}

	return &terminal, nil
}

func (r *Repository) ListTerminals(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]Terminal, int64, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			t.id,
			t.tenant_id,
			t.branch_id,
			b.name,
			t.terminal_code,
			t.machine_name,
			t.device_identifier,
			t.status,
			t.registered_at,
			t.last_seen_at,
			t.disabled_at,
			t.created_at,
			t.updated_at
		FROM terminals t
		INNER JOIN branches b
			ON b.id = t.branch_id
		   AND b.tenant_id = t.tenant_id
		WHERE t.tenant_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
		`,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	terminals := make([]Terminal, 0)

	for rows.Next() {
		var terminal Terminal

		if err := rows.Scan(
			&terminal.ID,
			&terminal.TenantID,
			&terminal.BranchID,
			&terminal.BranchName,
			&terminal.TerminalCode,
			&terminal.MachineName,
			&terminal.DeviceID,
			&terminal.Status,
			&terminal.RegisteredAt,
			&terminal.LastSeenAt,
			&terminal.DisabledAt,
			&terminal.CreatedAt,
			&terminal.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		terminals = append(terminals, terminal)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64

	err = r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM terminals
		WHERE tenant_id = $1
		`,
		tenantID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	return terminals, total, nil
}

func (r *Repository) UpdateTerminalName(
	ctx context.Context,
	tenantID string,
	terminalID string,
	machineName string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE terminals
		SET machine_name = $1,
		    updated_at = NOW()
		WHERE id = $2
		  AND tenant_id = $3
		`,
		machineName,
		terminalID,
		tenantID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("terminal not found")
	}

	return nil
}

func (r *Repository) ChangeTerminalStatus(
	ctx context.Context,
	tenantID string,
	terminalID string,
	status string,
) error {
	var disabledAt interface{}

	if status == "disabled" {
		disabledAt = time.Now()
	} else {
		disabledAt = nil
	}

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE terminals
		SET status = $1::terminal_status,
		    disabled_at = $2,
		    updated_at = NOW()
		WHERE id = $3
		  AND tenant_id = $4
		`,
		status,
		disabledAt,
		terminalID,
		tenantID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("terminal not found")
	}

	return nil
}

func (r *Repository) MoveTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Verify destination branch belongs to the same tenant
	// and is active.
	var branchExists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE id = $1
			  AND tenant_id = $2
			  AND status = 'active'
		)
		`,
		branchID,
		tenantID,
	).Scan(&branchExists)

	if err != nil {
		return err
	}

	if !branchExists {
		return errors.New("destination branch not found or inactive")
	}

	// Move terminal only within the same tenant.
	result, err := tx.Exec(
		ctx,
		`
		UPDATE terminals
		SET
			branch_id = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND tenant_id = $3
		`,
		branchID,
		terminalID,
		tenantID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("terminal not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// ============================================================
// SELF REGISTRATION
// ============================================================

// RegisterTerminal consumes a licence key and creates:
//
// 1. terminal
// 2. permanent terminal credential
//
// Everything happens inside one database transaction.
//
// If anything fails, the licence key remains unused.
func (r *Repository) RegisterTerminal(
	ctx context.Context,
	req RegisterTerminalRequest,
) (*RegisterTerminalResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	keyHash := hashSecret(req.LicenceKey)

	var (
		licenceID string
		tenantID  string
		branchID  string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			lk.id,
			lk.tenant_id,
			lk.branch_id
		FROM licence_keys lk
		INNER JOIN branches b
			ON b.id = lk.branch_id
		   AND b.tenant_id = lk.tenant_id
		WHERE lk.licence_key_hash = $1
		  AND lk.used_at IS NULL
		  AND lk.revoked_at IS NULL
		  AND (lk.expires_at IS NULL OR lk.expires_at > NOW())
		  AND b.status = 'active'
		FOR UPDATE
		`,
		keyHash,
	).Scan(
		&licenceID,
		&tenantID,
		&branchID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(
				"invalid, expired, revoked, or already used licence key",
			)
		}

		return nil, err
	}

	// Generate a server-side terminal code.
	terminalCodeSuffix, err := generateSecret(6)
	if err != nil {
		return nil, err
	}

	terminalCode := "TERM-" + strings.ToUpper(terminalCodeSuffix)

	terminalID := uuid.New().String()

	credential, err := generateSecret(32)
	if err != nil {
		return nil, err
	}

	credentialHash := hashSecret(credential)

	// Prevent duplicate physical device registration.
	var existingTerminalID string

	err = tx.QueryRow(
		ctx,
		`
		SELECT id
		FROM terminals
		WHERE device_identifier = $1
		`,
		req.DeviceID,
	).Scan(&existingTerminalID)

	if err == nil {
		return nil, errors.New(
			"this computer is already registered",
		)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO terminals (
			id,
			tenant_id,
			branch_id,
			terminal_code,
			machine_name,
			device_identifier,
			status,
			registered_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			'active',
			NOW()
		)
		`,
		terminalID,
		tenantID,
		branchID,
		terminalCode,
		req.MachineName,
		req.DeviceID,
	)

	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO terminal_credentials (
			id,
			terminal_id,
			credential_hash
		)
		VALUES ($1, $2, $3)
		`,
		uuid.New().String(),
		terminalID,
		credentialHash,
	)

	if err != nil {
		return nil, err
	}

	// The licence key is now permanently consumed.
	_, err = tx.Exec(
		ctx,
		`
		UPDATE licence_keys
		SET used_at = NOW()
		WHERE id = $1
		`,
		licenceID,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RegisterTerminalResponse{
		TerminalID:   terminalID,
		TenantID:     tenantID,
		BranchID:     branchID,
		TerminalCode: terminalCode,
		Credential:   credential,
	}, nil
}

// ============================================================
// TERMINAL AUTHENTICATION
// ============================================================

func (r *Repository) AuthenticateTerminal(
	ctx context.Context,
	credentialHash string,
) (string, string, string, error) {
	var terminalID string
	var tenantID string
	var branchID string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			t.id,
			t.tenant_id,
			t.branch_id
		FROM terminal_credentials tc
		INNER JOIN terminals t
			ON t.id = tc.terminal_id
		WHERE tc.credential_hash = $1
		  AND tc.revoked_at IS NULL
		  AND t.status = 'active'
		LIMIT 1
		`,
		credentialHash,
	).Scan(
		&terminalID,
		&tenantID,
		&branchID,
	)

	if err != nil {
		return "", "", "", err
	}

	return terminalID, tenantID, branchID, nil
}

// ============================================================
// HEARTBEAT
// ============================================================

func (r *Repository) Heartbeat(
	ctx context.Context,
	terminalID string,
	machineName *string,
	deviceIdentifier *string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE terminals
		SET
			machine_name = COALESCE($2, machine_name),
			device_identifier = COALESCE($3, device_identifier),
			last_seen_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		`,
		terminalID,
		machineName,
		deviceIdentifier,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("terminal not found")
	}

	return nil
}

// ============================================================
// TERMINAL CONTROL STATE
// ============================================================

// GetControlState returns the current desired lock state.
//
// Newly registered terminals may not have a control-state row yet.
// In that case we create the safe default:
//
//	unlocked
//
// This is important because heartbeat calls GetControlState()
// immediately after registration/authentication.
func (r *Repository) GetControlState(
	ctx context.Context,
	terminalID string,
) (*TerminalControlState, error) {
	var state TerminalControlState

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO terminal_control_state (
			terminal_id,
			desired_state,
			command_version
		)
		VALUES (
			$1,
			'unlocked'::terminal_lock_state,
			1
		)
		ON CONFLICT (terminal_id)
		DO UPDATE SET
			terminal_id = terminal_control_state.terminal_id
		RETURNING
			terminal_id,
			desired_state,
			command_version,
			updated_at
		`,
		terminalID,
	).Scan(
		&state.TerminalID,
		&state.DesiredState,
		&state.CommandVersion,
		&state.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (r *Repository) SetControlState(
	ctx context.Context,
	terminalID string,
	state string,
) (*TerminalControlState, error) {
	var result TerminalControlState

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO terminal_control_state (
			terminal_id,
			desired_state,
			command_version
		)
		VALUES (
			$1,
			$2::terminal_lock_state,
			1
		)
		ON CONFLICT (terminal_id)
		DO UPDATE SET
			desired_state = EXCLUDED.desired_state,
			command_version =
				terminal_control_state.command_version + 1,
			updated_at = NOW()
		RETURNING
			terminal_id,
			desired_state,
			command_version,
			updated_at
		`,
		terminalID,
		state,
	).Scan(
		&result.TerminalID,
		&result.DesiredState,
		&result.CommandVersion,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================
// TENANT VALIDATION
// ============================================================

func (r *Repository) VerifyTerminalBelongsToTenant(
	ctx context.Context,
	terminalID string,
	tenantID string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM terminals
			WHERE id = $1
			  AND tenant_id = $2
		)
		`,
		terminalID,
		tenantID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("terminal not found")
	}

	return nil
}

// ============================================================
// DEREGISTER TERMINAL
// ============================================================

func (r *Repository) DeregisterTerminal(
	ctx context.Context,
	tenantID string,
	terminalID string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var terminalExists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM terminals
			WHERE id = $1
			  AND tenant_id = $2
		)
		`,
		terminalID,
		tenantID,
	).Scan(&terminalExists)

	if err != nil {
		return err
	}

	if !terminalExists {
		return errors.New("terminal not found")
	}

	// Keep the terminal row for historical records.
	// Release the physical device identifier so the same
	// computer can be enrolled again later with a new key.
	_, err = tx.Exec(
		ctx,
		`
		UPDATE terminals
		SET
			status = 'disabled',
			device_identifier = NULL,
			disabled_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		`,
		terminalID,
		tenantID,
	)

	if err != nil {
		return err
	}

	// Revoke all credentials belonging to this terminal.
	_, err = tx.Exec(
		ctx,
		`
		UPDATE terminal_credentials
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE terminal_id = $1
		  AND revoked_at IS NULL
		`,
		terminalID,
	)

	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// The C# client must acknowledge the command after executing it.
func (r *Repository) QueueCommand(
	ctx context.Context,
	terminalID string,
	command TerminalCommand,
) (*TerminalCommandResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Lock the terminal row so command_version remains
	// sequential even when multiple commands are queued concurrently.
	var lockedID string

	err = tx.QueryRow(
		ctx,
		`
		SELECT id
		FROM terminals
		WHERE id = $1
		FOR UPDATE
		`,
		terminalID,
	).Scan(&lockedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("terminal not found")
		}
		return nil, err
	}

	var commandVersion int64

	err = tx.QueryRow(
		ctx,
		`
		SELECT COALESCE(MAX(command_version), 0) + 1
		FROM terminal_commands
		WHERE terminal_id = $1
		`,
		terminalID,
	).Scan(&commandVersion)

	if err != nil {
		return nil, err
	}

	var result TerminalCommandResponse

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO terminal_commands (
			id,
			terminal_id,
			command,
			status,
			command_version
		)
		VALUES (
			$1,
			$2,
			$3::terminal_command_type,
			'pending'::terminal_command_status,
			$4
		)
		RETURNING
			id,
			terminal_id,
			command,
			status,
			command_version,
			created_at,
			executed_at,
			acknowledged_at,
			error_message
		`,
		uuid.New().String(),
		terminalID,
		command,
		commandVersion,
	).Scan(
		&result.ID,
		&result.TerminalID,
		&result.Command,
		&result.Status,
		&result.CommandVersion,
		&result.CreatedAt,
		&result.ExecutedAt,
		&result.AcknowledgedAt,
		&result.ErrorMessage,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &result, nil
}

// get pending commands
func (r *Repository) GetPendingCommand(
	ctx context.Context,
	terminalID string,
) (*TerminalCommandResponse, error) {
	var result TerminalCommandResponse

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			terminal_id,
			command,
			status,
			command_version,
			created_at,
			executed_at,
			acknowledged_at,
			error_message
		FROM terminal_commands
		WHERE terminal_id = $1
		  AND status = 'pending'::terminal_command_status
		ORDER BY command_version ASC
		LIMIT 1
		`,
		terminalID,
	).Scan(
		&result.ID,
		&result.TerminalID,
		&result.Command,
		&result.Status,
		&result.CommandVersion,
		&result.CreatedAt,
		&result.ExecutedAt,
		&result.AcknowledgedAt,
		&result.ErrorMessage,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

// acvknowledge commands
func (r *Repository) AcknowledgeCommand(
	ctx context.Context,
	terminalID string,
	commandID string,
	status string,
	errorMessage *string,
) (*TerminalCommandResponse, error) {
	var result TerminalCommandResponse

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE terminal_commands
		SET
			status = $1::terminal_command_status,
			executed_at = CASE
				WHEN $1 = 'executed'
				THEN COALESCE(executed_at, NOW())
				ELSE executed_at
			END,
			acknowledged_at = NOW(),
			error_message = $2
		WHERE id = $3
		  AND terminal_id = $4
		  AND status = 'pending'::terminal_command_status
		RETURNING
			id,
			terminal_id,
			command,
			status,
			command_version,
			created_at,
			executed_at,
			acknowledged_at,
			error_message
		`,
		status,
		errorMessage,
		commandID,
		terminalID,
	).Scan(
		&result.ID,
		&result.TerminalID,
		&result.Command,
		&result.Status,
		&result.CommandVersion,
		&result.CreatedAt,
		&result.ExecutedAt,
		&result.AcknowledgedAt,
		&result.ErrorMessage,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("pending command not found")
		}
		return nil, err
	}

	return &result, nil
}

func (r *Repository) ListAttendantTerminals(
	ctx context.Context,
	tenantID string,
	userID string,
	limit int,
	offset int,
) ([]Terminal, int64, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			t.id,
			t.tenant_id,
			t.branch_id,
			b.name,
			t.terminal_code,
			t.machine_name,
			t.device_identifier,
			t.status,
			t.registered_at,
			t.last_seen_at,
			t.disabled_at,
			t.created_at,
			t.updated_at
		FROM terminals t
		INNER JOIN branches b
			ON b.id = t.branch_id
		   AND b.tenant_id = t.tenant_id
		INNER JOIN attendant_branch_assignments aba
			ON aba.branch_id = t.branch_id
		   AND aba.tenant_id = t.tenant_id
		   AND aba.user_id = $2
		   AND aba.unassigned_at IS NULL
		WHERE t.tenant_id = $1
		ORDER BY t.created_at DESC
		LIMIT $3 OFFSET $4
		`,
		tenantID,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	terminals := make([]Terminal, 0)

	for rows.Next() {
		var terminal Terminal

		if err := rows.Scan(
			&terminal.ID,
			&terminal.TenantID,
			&terminal.BranchID,
			&terminal.BranchName,
			&terminal.TerminalCode,
			&terminal.MachineName,
			&terminal.DeviceID,
			&terminal.Status,
			&terminal.RegisteredAt,
			&terminal.LastSeenAt,
			&terminal.DisabledAt,
			&terminal.CreatedAt,
			&terminal.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		terminals = append(terminals, terminal)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64

	err = r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM terminals t
		INNER JOIN attendant_branch_assignments aba
			ON aba.branch_id = t.branch_id
		   AND aba.tenant_id = t.tenant_id
		   AND aba.user_id = $2
		   AND aba.unassigned_at IS NULL
		WHERE t.tenant_id = $1
		`,
		tenantID,
		userID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	return terminals, total, nil
}

func (r *Repository) GetAttendantTerminal(
	ctx context.Context,
	tenantID string,
	userID string,
	terminalID string,
) (*Terminal, error) {
	var terminal Terminal

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			t.id,
			t.tenant_id,
			t.branch_id,
			b.name,
			t.terminal_code,
			t.machine_name,
			t.device_identifier,
			t.status,
			t.registered_at,
			t.last_seen_at,
			t.disabled_at,
			t.created_at,
			t.updated_at
		FROM terminals t
		INNER JOIN branches b
			ON b.id = t.branch_id
		   AND b.tenant_id = t.tenant_id
		INNER JOIN attendant_branch_assignments aba
			ON aba.branch_id = t.branch_id
		   AND aba.tenant_id = t.tenant_id
		   AND aba.user_id = $2
		   AND aba.unassigned_at IS NULL
		WHERE t.id = $3
		  AND t.tenant_id = $1
		`,
		tenantID,
		userID,
		terminalID,
	).Scan(
		&terminal.ID,
		&terminal.TenantID,
		&terminal.BranchID,
		&terminal.BranchName,
		&terminal.TerminalCode,
		&terminal.MachineName,
		&terminal.DeviceID,
		&terminal.Status,
		&terminal.RegisteredAt,
		&terminal.LastSeenAt,
		&terminal.DisabledAt,
		&terminal.CreatedAt,
		&terminal.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("terminal not found")
		}

		return nil, err
	}

	return &terminal, nil
}

// SetAttendantControlState allows an attendant to lock or unlock
// only a terminal belonging to their currently assigned branch.
func (r *Repository) SetAttendantControlState(
	ctx context.Context,
	tenantID string,
	userID string,
	terminalID string,
	state string,
) (*TerminalControlState, error) {
	var result TerminalControlState

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO terminal_control_state (
			terminal_id,
			desired_state,
			command_version
		)
		SELECT
			t.id,
			$4::terminal_lock_state,
			1
		FROM terminals t
		INNER JOIN attendant_branch_assignments aba
			ON aba.branch_id = t.branch_id
		   AND aba.tenant_id = t.tenant_id
		   AND aba.user_id = $2
		   AND aba.unassigned_at IS NULL
		WHERE t.id = $3
		  AND t.tenant_id = $1
		ON CONFLICT (terminal_id)
		DO UPDATE SET
			desired_state = EXCLUDED.desired_state,
			command_version =
				terminal_control_state.command_version + 1,
			updated_at = NOW()
		RETURNING
			terminal_id,
			desired_state,
			command_version,
			updated_at
		`,
		tenantID,
		userID,
		terminalID,
		state,
	).Scan(
		&result.TerminalID,
		&result.DesiredState,
		&result.CommandVersion,
		&result.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("terminal not found or not assigned to attendant branch")
		}

		return nil, err
	}

	return &result, nil
}

func (r *Repository) QueueAttendantCommand(
	ctx context.Context,
	tenantID string,
	userID string,
	terminalID string,
	command TerminalCommand,
) (*TerminalCommandResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Lock only if the terminal belongs to the attendant's
	// currently assigned branch.
	var lockedID string

	err = tx.QueryRow(
		ctx,
		`
		SELECT t.id
		FROM terminals t
		INNER JOIN attendant_branch_assignments aba
			ON aba.branch_id = t.branch_id
		   AND aba.tenant_id = t.tenant_id
		   AND aba.user_id = $2
		   AND aba.unassigned_at IS NULL
		WHERE t.id = $3
		  AND t.tenant_id = $1
		FOR UPDATE
		`,
		tenantID,
		userID,
		terminalID,
	).Scan(&lockedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(
				"terminal not found or not assigned to attendant branch",
			)
		}

		return nil, err
	}

	var commandVersion int64

	err = tx.QueryRow(
		ctx,
		`
		SELECT COALESCE(MAX(command_version), 0) + 1
		FROM terminal_commands
		WHERE terminal_id = $1
		`,
		terminalID,
	).Scan(&commandVersion)

	if err != nil {
		return nil, err
	}

	var result TerminalCommandResponse

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO terminal_commands (
			id,
			terminal_id,
			command,
			status,
			command_version
		)
		VALUES (
			$1,
			$2,
			$3::terminal_command_type,
			'pending'::terminal_command_status,
			$4
		)
		RETURNING
			id,
			terminal_id,
			command,
			status,
			command_version,
			created_at,
			executed_at,
			acknowledged_at,
			error_message
		`,
		uuid.New().String(),
		terminalID,
		command,
		commandVersion,
	).Scan(
		&result.ID,
		&result.TerminalID,
		&result.Command,
		&result.Status,
		&result.CommandVersion,
		&result.CreatedAt,
		&result.ExecutedAt,
		&result.AcknowledgedAt,
		&result.ErrorMessage,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &result, nil
}

// ============================================================
// AUDIT LOG
// ============================================================

func (r *Repository) CreateAuditLog(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	action string,
	entityType string,
	entityID string,
	oldValue []byte,
	newValue []byte,
	reason string,
	ipAddress *string,
	deviceID *string,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO audit_logs (
			id,
			tenant_id,
			user_id,
			user_role,
			action,
			entity_type,
			entity_id,
			old_value,
			new_value,
			reason,
			ip_address,
			device_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4::user_role,
			$5,
			$6,
			$7,
			$8::jsonb,
			$9::jsonb,
			$10,
			$11,
			$12
		)
		`,
		uuid.New().String(),
		tenantID,
		userID,
		userRole,
		action,
		entityType,
		entityID,
		jsonOrNull(oldValue),
		jsonOrNull(newValue),
		reason,
		ipAddress,
		deviceID,
	)

	return err
}

func jsonOrNull(value []byte) interface{} {
	if len(value) == 0 {
		return nil
	}

	return string(value)
}

// ============================================================
// ERROR HELPERS
// ============================================================

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}

func normalizeMachineName(value string) string {
	return strings.TrimSpace(value)
}

func validateTerminalInput(req RegisterTerminalRequest) error {
	if strings.TrimSpace(req.LicenceKey) == "" {
		return errors.New("licence key is required")
	}

	if normalizeMachineName(req.MachineName) == "" {
		return errors.New("machine name is required")
	}

	if strings.TrimSpace(req.DeviceID) == "" {
		return errors.New("device identifier is required")
	}

	if len(req.DeviceID) > 255 {
		return errors.New("device identifier is too long")
	}

	if len(req.MachineName) > 200 {
		return errors.New("machine name is too long")
	}

	return nil
}

func _unusedFmtReference() {
	_ = fmt.Sprintf
}
