package attendant

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("attendant not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAttendant(
	ctx context.Context,
	tenantID string,
	req CreateAttendantRequest,
	passwordHash string,
) (Attendant, error) {
	id := uuid.New()

	var attendant Attendant

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			id,
			tenant_id,
			full_name,
			email,
			phone,
			password_hash,
			role,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			'attendant',
			'active'
		)
		RETURNING
			id,
			tenant_id,
			full_name,
			email,
			phone,
			status,
			created_at,
			updated_at
		`,
		id,
		tenantID,
		req.FullName,
		req.Email,
		req.Phone,
		passwordHash,
	).Scan(
		&attendant.ID,
		&attendant.TenantID,
		&attendant.FullName,
		&attendant.Email,
		&attendant.Phone,
		&attendant.Status,
		&attendant.CreatedAt,
		&attendant.UpdatedAt,
	)

	if err != nil {
		return Attendant{}, err
	}

	return attendant, nil
}

func (r *Repository) GetAttendant(
	ctx context.Context,
	tenantID string,
	attendantID string,
) (Attendant, error) {
	var a Attendant

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			u.id,
			u.tenant_id,
			u.full_name,
			u.email,
			u.phone,
			u.status,
			aba.branch_id,
			NULL::text,
			u.created_at,
			u.updated_at
		FROM users u
		LEFT JOIN attendant_branch_assignments aba
			ON aba.user_id = u.id
			AND aba.tenant_id = u.tenant_id
			AND aba.unassigned_at IS NULL
		WHERE u.id = $1
		  AND u.tenant_id = $2
		  AND u.role = 'attendant'
		`,
		attendantID,
		tenantID,
	).Scan(
		&a.ID,
		&a.TenantID,
		&a.FullName,
		&a.Email,
		&a.Phone,
		&a.Status,
		&a.BranchID,
		&a.BranchName,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Attendant{}, ErrNotFound
	}

	if err != nil {
		return Attendant{}, err
	}

	return a, nil
}

func (r *Repository) ListAttendants(
	ctx context.Context,
	tenantID string,
) ([]Attendant, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			u.id,
			u.tenant_id,
			u.full_name,
			u.email,
			u.phone,
			u.status,
			aba.branch_id,
			b.name,
			u.created_at,
			u.updated_at
		FROM users u
		LEFT JOIN attendant_branch_assignments aba
			ON aba.user_id = u.id
			AND aba.tenant_id = u.tenant_id
			AND aba.unassigned_at IS NULL
		LEFT JOIN branches b
			ON b.id = aba.branch_id
		WHERE u.tenant_id = $1
		  AND u.role = 'attendant'
		ORDER BY u.created_at DESC
		`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendants []Attendant

	for rows.Next() {
		var a Attendant

		if err := rows.Scan(
			&a.ID,
			&a.TenantID,
			&a.FullName,
			&a.Email,
			&a.Phone,
			&a.Status,
			&a.BranchID,
			&a.BranchName,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		attendants = append(attendants, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attendants, nil
}

func (r *Repository) CountAttendants(
	ctx context.Context,
	tenantID string,
) (int64, error) {
	var count int64

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM users
		WHERE tenant_id = $1
		  AND role = 'attendant'
		`,
		tenantID,
	).Scan(&count)

	return count, err
}

func (r *Repository) UpdateAttendant(
	ctx context.Context,
	tenantID string,
	attendantID string,
	req UpdateAttendantRequest,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE users
		SET
			full_name = $1,
			email = $2,
			phone = $3
		WHERE id = $4
		  AND tenant_id = $5
		  AND role = 'attendant'
		`,
		req.FullName,
		req.Email,
		req.Phone,
		attendantID,
		tenantID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) UpdateStatus(
	ctx context.Context,
	tenantID string,
	attendantID string,
	status string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE users
		SET status = $1
		WHERE id = $2
		  AND tenant_id = $3
		  AND role = 'attendant'
		`,
		status,
		attendantID,
		tenantID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
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
		)
		`,
		branchID,
		tenantID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("branch does not belong to this tenant")
	}

	return nil
}

func (r *Repository) AssignBranch(
	ctx context.Context,
	tenantID string,
	attendantID string,
	branchID string,
	assignedBy string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
			  AND tenant_id = $2
			  AND role = 'attendant'
		)
		`,
		attendantID,
		tenantID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return ErrNotFound
	}

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE id = $1
			  AND tenant_id = $2
		)
		`,
		branchID,
		tenantID,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("branch does not belong to this tenant")
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE attendant_branch_assignments
		SET unassigned_at = NOW()
		WHERE user_id = $1
		  AND tenant_id = $2
		  AND unassigned_at IS NULL
		`,
		attendantID,
		tenantID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO attendant_branch_assignments (
			id,
			tenant_id,
			user_id,
			branch_id,
			assigned_by
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		uuid.New(),
		tenantID,
		attendantID,
		branchID,
		assignedBy,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return errors.New(
					"attendant already has an active branch assignment",
				)
			}
		}

		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) UnassignBranch(
	ctx context.Context,
	tenantID string,
	attendantID string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE attendant_branch_assignments
		SET unassigned_at = NOW()
		WHERE user_id = $1
		  AND tenant_id = $2
		  AND unassigned_at IS NULL
		`,
		attendantID,
		tenantID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) CreateAuditLog(
	ctx context.Context,
	tenantID string,
	userID string,
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
			user_id,
			user_role,
			action,
			entity_type,
			entity_id,
			reason
		)
		VALUES (
			$1,
			$2,
			$3,
			'owner',
			$4,
			'attendant',
			$5,
			$6
		)
		`,
		uuid.New(),
		tenantID,
		userID,
		action,
		entityID,
		reason,
	)

	return err
}
