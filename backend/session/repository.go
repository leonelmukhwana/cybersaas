package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrBranchAccessDenied  = errors.New("branch access denied")
	ErrTerminalNotFound    = errors.New("terminal not found")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrActiveSessionExists = errors.New("terminal already has an active session")
	ErrDuplicateOperation  = errors.New("client operation already exists")
	ErrInvalidSessionState = errors.New("invalid session state")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

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
		return ErrBranchAccessDenied
	}

	return nil
}

func (r *Repository) VerifyAttendantBranchAccess(
	ctx context.Context,
	userID string,
	tenantID string,
	branchID string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM attendant_branch_assignments
			WHERE user_id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
			  AND unassigned_at IS NULL
		)
		`,
		userID,
		tenantID,
		branchID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return ErrBranchAccessDenied
	}

	return nil
}

func (r *Repository) VerifyTerminal(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
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
			  AND branch_id = $3
			  AND status = 'active'
		)
		`,
		terminalID,
		tenantID,
		branchID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return ErrTerminalNotFound
	}

	return nil
}

func (r *Repository) VerifyCustomer(
	ctx context.Context,
	tenantID string,
	branchID string,
	customerID string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM customers
			WHERE id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
		)
		`,
		customerID,
		tenantID,
		branchID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return ErrCustomerNotFound
	}

	return nil
}

func (r *Repository) GetBillingConfig(
	ctx context.Context,
	tenantID string,
	branchID string,
) (
	rate string,
	minimum string,
	interval int,
	rounding string,
	currency string,
	err error,
) {
	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			rate_per_minute::text,
			minimum_charge::text,
			billing_interval_minutes,
			rounding_mode,
			currency
		FROM billing_configs
		WHERE tenant_id = $1
		  AND branch_id = $2
		`,
		tenantID,
		branchID,
	).Scan(
		&rate,
		&minimum,
		&interval,
		&rounding,
		&currency,
	)

	return
}

func (r *Repository) HasActiveSession(
	ctx context.Context,
	tenantID string,
	terminalID string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM sessions
			WHERE tenant_id = $1
			  AND terminal_id = $2
			  AND status = 'active'
		)
		`,
		tenantID,
		terminalID,
	).Scan(&exists)

	return exists, err
}

func (r *Repository) Create(
	ctx context.Context,
	session Session,
) (*Session, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var terminalExists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM terminals
			WHERE id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
			  AND status = 'active'
		)
		`,
		session.TerminalID,
		session.TenantID,
		session.BranchID,
	).Scan(&terminalExists)

	if err != nil {
		return nil, err
	}

	if !terminalExists {
		return nil, ErrTerminalNotFound
	}

	_, err = tx.Exec(
		ctx,
		`
		SELECT id
		FROM terminals
		WHERE id = $1
		FOR UPDATE
		`,
		session.TerminalID,
	)
	if err != nil {
		return nil, err
	}

	var active bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM sessions
			WHERE tenant_id = $1
			  AND terminal_id = $2
			  AND status = 'active'
		)
		`,
		session.TenantID,
		session.TerminalID,
	).Scan(&active)

	if err != nil {
		return nil, err
	}

	if active {
		return nil, ErrActiveSessionExists
	}

	var existing Session

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			terminal_id,
			customer_id,
			attendant_id,
			session_type,
			status,
			started_at,
			ended_at,
			paused_at,
			total_paused_seconds,
			prepaid_amount,
			allowed_minutes,
			rate_per_minute,
			minimum_charge,
			final_amount,
			client_operation_id,
			created_at,
			updated_at
		FROM sessions
		WHERE client_operation_id = $1
		  AND started_at = $2
		`,
		session.ClientOperationID,
		session.StartedAt,
	).Scan(
		&existing.ID,
		&existing.TenantID,
		&existing.BranchID,
		&existing.TerminalID,
		&existing.CustomerID,
		&existing.AttendantID,
		&existing.SessionType,
		&existing.Status,
		&existing.StartedAt,
		&existing.EndedAt,
		&existing.PausedAt,
		&existing.TotalPausedSeconds,
		&existing.PrepaidAmount,
		&existing.AllowedMinutes,
		&existing.RatePerMinute,
		&existing.MinimumCharge,
		&existing.FinalAmount,
		&existing.ClientOperationID,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)

	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		return r.populateDisplayFields(ctx, &existing)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO sessions (
			id,
			tenant_id,
			branch_id,
			terminal_id,
			customer_id,
			attendant_id,
			session_type,
			status,
			started_at,
			prepaid_amount,
			allowed_minutes,
			rate_per_minute,
			minimum_charge,
			client_operation_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7::session_type,
			'active',
			$8,
			$9,
			$10,
			$11,
			$12,
			$13
		)
		`,
		session.ID,
		session.TenantID,
		session.BranchID,
		session.TerminalID,
		session.CustomerID,
		session.AttendantID,
		session.SessionType,
		session.StartedAt,
		session.PrepaidAmount,
		session.AllowedMinutes,
		session.RatePerMinute,
		session.MinimumCharge,
		session.ClientOperationID,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.populateDisplayFields(ctx, &session)
}

func (r *Repository) Get(
	ctx context.Context,
	tenantID string,
	branchID string,
	sessionID string,
	startedAt time.Time,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			s.id,
			s.tenant_id,
			s.branch_id,
			s.terminal_id,
			s.customer_id,
			s.attendant_id,
			s.session_type,
			s.status,
			s.started_at,
			s.ended_at,
			s.paused_at,
			s.total_paused_seconds,
			s.prepaid_amount,
			s.allowed_minutes,
			s.rate_per_minute,
			s.minimum_charge,
			s.final_amount,
			s.client_operation_id,
			s.created_at,
			s.updated_at,
			c.full_name,
			c.customer_type,
			c.parent_name,
			t.machine_name
		FROM sessions s
		INNER JOIN customers c
			ON c.id = s.customer_id
		   AND c.tenant_id = s.tenant_id
		   AND c.branch_id = s.branch_id
		INNER JOIN terminals t
			ON t.id = s.terminal_id
		   AND t.tenant_id = s.tenant_id
		   AND t.branch_id = s.branch_id
		WHERE s.id = $1
		  AND s.started_at = $2
		  AND s.tenant_id = $3
		  AND s.branch_id = $4
		`,
		sessionID,
		startedAt,
		tenantID,
		branchID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CustomerName,
		&result.CustomerType,
		&result.ParentName,
		&result.TerminalName,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) GetLatestForID(
	ctx context.Context,
	tenantID string,
	branchID string,
	sessionID string,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			s.id,
			s.tenant_id,
			s.branch_id,
			s.terminal_id,
			s.customer_id,
			s.attendant_id,
			s.session_type,
			s.status,
			s.started_at,
			s.ended_at,
			s.paused_at,
			s.total_paused_seconds,
			s.prepaid_amount,
			s.allowed_minutes,
			s.rate_per_minute,
			s.minimum_charge,
			s.final_amount,
			s.client_operation_id,
			s.created_at,
			s.updated_at,
			c.full_name,
			c.customer_type,
			c.parent_name,
			t.machine_name
		FROM sessions s
		INNER JOIN customers c
			ON c.id = s.customer_id
		   AND c.tenant_id = s.tenant_id
		   AND c.branch_id = s.branch_id
		INNER JOIN terminals t
			ON t.id = s.terminal_id
		   AND t.tenant_id = s.tenant_id
		   AND t.branch_id = s.branch_id
		WHERE s.id = $1
		  AND s.tenant_id = $2
		  AND s.branch_id = $3
		ORDER BY s.started_at DESC
		LIMIT 1
		`,
		sessionID,
		tenantID,
		branchID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CustomerName,
		&result.CustomerType,
		&result.ParentName,
		&result.TerminalName,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) End(
	ctx context.Context,
	tenantID string,
	branchID string,
	sessionID string,
	startedAt time.Time,
	endedAt time.Time,
	finalAmount string,
	status string,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE sessions
		SET
			ended_at = $5,
			final_amount = $6,
			status = $7::session_status,
			updated_at = NOW()
		WHERE id = $1
		  AND started_at = $2
		  AND tenant_id = $3
		  AND branch_id = $4
		  AND status = 'active'
		RETURNING
			id,
			tenant_id,
			branch_id,
			terminal_id,
			customer_id,
			attendant_id,
			session_type,
			status,
			started_at,
			ended_at,
			paused_at,
			total_paused_seconds,
			prepaid_amount,
			allowed_minutes,
			rate_per_minute,
			minimum_charge,
			final_amount,
			client_operation_id,
			created_at,
			updated_at
		`,
		sessionID,
		startedAt,
		tenantID,
		branchID,
		endedAt,
		finalAmount,
		status,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	return r.populateDisplayFields(ctx, &result)
}

func (r *Repository) Cancel(
	ctx context.Context,
	tenantID string,
	branchID string,
	sessionID string,
	startedAt time.Time,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE sessions
		SET
			ended_at = COALESCE(ended_at, NOW()),
			status = 'cancelled',
			updated_at = NOW()
		WHERE id = $1
		  AND started_at = $2
		  AND tenant_id = $3
		  AND branch_id = $4
		  AND status = 'active'
		RETURNING
			id,
			tenant_id,
			branch_id,
			terminal_id,
			customer_id,
			attendant_id,
			session_type,
			status,
			started_at,
			ended_at,
			paused_at,
			total_paused_seconds,
			prepaid_amount,
			allowed_minutes,
			rate_per_minute,
			minimum_charge,
			final_amount,
			client_operation_id,
			created_at,
			updated_at
		`,
		sessionID,
		startedAt,
		tenantID,
		branchID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	return r.populateDisplayFields(ctx, &result)
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	branchID string,
	status string,
	startTime *time.Time,
	endTime *time.Time,
	limit int,
	offset int,
) ([]Session, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			s.id,
			s.tenant_id,
			s.branch_id,
			s.terminal_id,
			s.customer_id,
			s.attendant_id,
			s.session_type,
			s.status,
			s.started_at,
			s.ended_at,
			s.paused_at,
			s.total_paused_seconds,
			s.prepaid_amount,
			s.allowed_minutes,
			s.rate_per_minute,
			s.minimum_charge,
			s.final_amount,
			s.client_operation_id,
			s.created_at,
			s.updated_at,
			c.full_name,
			c.customer_type,
			c.parent_name,
			t.machine_name
		FROM sessions s
		INNER JOIN customers c
			ON c.id = s.customer_id
		   AND c.tenant_id = s.tenant_id
		   AND c.branch_id = s.branch_id
		INNER JOIN terminals t
			ON t.id = s.terminal_id
		   AND t.tenant_id = s.tenant_id
		   AND t.branch_id = s.branch_id
		WHERE s.tenant_id = $1
		  AND s.branch_id = $2
		  AND ($3 = '' OR s.status = $3::session_status)
		  AND ($4::timestamptz IS NULL OR s.started_at >= $4)
		  AND ($5::timestamptz IS NULL OR s.started_at < $5)
		ORDER BY s.started_at DESC
		LIMIT $6 OFFSET $7
		`,
		tenantID,
		branchID,
		status,
		startTime,
		endTime,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Session, 0)

	for rows.Next() {
		var result Session

		if err := rows.Scan(
			&result.ID,
			&result.TenantID,
			&result.BranchID,
			&result.TerminalID,
			&result.CustomerID,
			&result.AttendantID,
			&result.SessionType,
			&result.Status,
			&result.StartedAt,
			&result.EndedAt,
			&result.PausedAt,
			&result.TotalPausedSeconds,
			&result.PrepaidAmount,
			&result.AllowedMinutes,
			&result.RatePerMinute,
			&result.MinimumCharge,
			&result.FinalAmount,
			&result.ClientOperationID,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CustomerName,
			&result.CustomerType,
			&result.ParentName,
			&result.TerminalName,
		); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

func (r *Repository) Count(
	ctx context.Context,
	tenantID string,
	branchID string,
	status string,
	startTime *time.Time,
	endTime *time.Time,
) (int64, error) {
	var total int64

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM sessions
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND ($3 = '' OR status = $3::session_status)
		  AND ($4::timestamptz IS NULL OR started_at >= $4)
		  AND ($5::timestamptz IS NULL OR started_at < $5)
		`,
		tenantID,
		branchID,
		status,
		startTime,
		endTime,
	).Scan(&total)

	return total, err
}

func (r *Repository) CreateAuditLog(
	ctx context.Context,
	tenantID string,
	branchID string,
	userID string,
	userRole string,
	action string,
	entityID string,
	reason string,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO audit_logs (
			id,
			tenant_id,
			branch_id,
			user_id,
			user_role,
			action,
			entity_type,
			entity_id,
			reason
		)
		VALUES (
			gen_random_uuid(),
			$1,
			$2,
			$3,
			$4::user_role,
			$5,
			'session',
			$6,
			$7
		)
		`,
		tenantID,
		branchID,
		userID,
		userRole,
		action,
		entityID,
		reason,
	)

	return err
}

func (r *Repository) GetActiveForTerminal(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			s.id,
			s.tenant_id,
			s.branch_id,
			s.terminal_id,
			s.customer_id,
			s.attendant_id,
			s.session_type,
			s.status,
			s.started_at,
			s.ended_at,
			s.paused_at,
			s.total_paused_seconds,
			s.prepaid_amount,
			s.allowed_minutes,
			s.rate_per_minute,
			s.minimum_charge,
			s.final_amount,
			s.client_operation_id,
			s.created_at,
			s.updated_at,
			c.full_name,
			c.customer_type,
			c.parent_name,
			t.machine_name
		FROM sessions s
		INNER JOIN customers c
			ON c.id = s.customer_id
		   AND c.tenant_id = s.tenant_id
		   AND c.branch_id = s.branch_id
		INNER JOIN terminals t
			ON t.id = s.terminal_id
		   AND t.tenant_id = s.tenant_id
		   AND t.branch_id = s.branch_id
		WHERE s.tenant_id = $1
		  AND s.branch_id = $2
		  AND s.terminal_id = $3
		  AND s.status = 'active'
		ORDER BY s.started_at DESC
		LIMIT 1
		`,
		tenantID,
		branchID,
		terminalID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CustomerName,
		&result.CustomerType,
		&result.ParentName,
		&result.TerminalName,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) PauseSession(
	ctx context.Context,
	tenantID string,
	branchID string,
	sessionID string,
	startedAt time.Time,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE sessions
		SET
			paused_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND started_at = $2
		  AND tenant_id = $3
		  AND branch_id = $4
		  AND status = 'active'
		  AND paused_at IS NULL
		RETURNING
			id,
			tenant_id,
			branch_id,
			terminal_id,
			customer_id,
			attendant_id,
			session_type,
			status,
			started_at,
			ended_at,
			paused_at,
			total_paused_seconds,
			prepaid_amount,
			allowed_minutes,
			rate_per_minute,
			minimum_charge,
			final_amount,
			client_operation_id,
			created_at,
			updated_at
		`,
		sessionID,
		startedAt,
		tenantID,
		branchID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidSessionState
	}

	if err != nil {
		return nil, err
	}

	return r.populateDisplayFields(ctx, &result)
}

func (r *Repository) ResumeSession(
	ctx context.Context,
	tenantID string,
	branchID string,
	sessionID string,
	startedAt time.Time,
) (*Session, error) {
	var result Session

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE sessions
		SET
			total_paused_seconds =
				total_paused_seconds +
				EXTRACT(EPOCH FROM (NOW() - paused_at))::BIGINT,
			paused_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		  AND started_at = $2
		  AND tenant_id = $3
		  AND branch_id = $4
		  AND status = 'active'
		  AND paused_at IS NOT NULL
		RETURNING
			id,
			tenant_id,
			branch_id,
			terminal_id,
			customer_id,
			attendant_id,
			session_type,
			status,
			started_at,
			ended_at,
			paused_at,
			total_paused_seconds,
			prepaid_amount,
			allowed_minutes,
			rate_per_minute,
			minimum_charge,
			final_amount,
			client_operation_id,
			created_at,
			updated_at
		`,
		sessionID,
		startedAt,
		tenantID,
		branchID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.TerminalID,
		&result.CustomerID,
		&result.AttendantID,
		&result.SessionType,
		&result.Status,
		&result.StartedAt,
		&result.EndedAt,
		&result.PausedAt,
		&result.TotalPausedSeconds,
		&result.PrepaidAmount,
		&result.AllowedMinutes,
		&result.RatePerMinute,
		&result.MinimumCharge,
		&result.FinalAmount,
		&result.ClientOperationID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidSessionState
	}

	if err != nil {
		return nil, err
	}

	return r.populateDisplayFields(ctx, &result)
}

/*
populateDisplayFields adds only safe, non-sensitive display information.

IMPORTANT:
- No customer ID number is selected.
- No encrypted customer ID is selected.
- No customer ID hash is selected.
- Parent ID number is also intentionally excluded.
*/
func (r *Repository) populateDisplayFields(
	ctx context.Context,
	session *Session,
) (*Session, error) {
	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			c.full_name,
			c.customer_type,
			c.parent_name,
			t.machine_name
		FROM customers c
		INNER JOIN terminals t
			ON t.id = $2
		   AND t.tenant_id = $3
		   AND t.branch_id = $4
		WHERE c.id = $1
		  AND c.tenant_id = $3
		  AND c.branch_id = $4
		`,
		session.CustomerID,
		session.TerminalID,
		session.TenantID,
		session.BranchID,
	).Scan(
		&session.CustomerName,
		&session.CustomerType,
		&session.ParentName,
		&session.TerminalName,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return session, nil
}

func newUUID() string {
	return uuid.NewString()
}
