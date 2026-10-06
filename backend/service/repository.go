package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrBranchNotFound       = errors.New("branch not found")
	ErrBranchAccessDenied   = errors.New("branch access denied")
	ErrServiceAlreadyExists = errors.New("service already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) VerifyBranchBelongsToTenant(
	ctx context.Context,
	branchID string,
	tenantID string,
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

func (r *Repository) Create(
	ctx context.Context,
	id string,
	tenantID string,
	branchID string,
	name string,
	description *string,
	price string,
) (*Service, error) {
	var service Service

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO services (
			id,
			tenant_id,
			branch_id,
			name,
			description,
			price,
			active
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			TRUE
		)
		RETURNING
			id,
			tenant_id,
			branch_id,
			name,
			description,
			price::text,
			active,
			created_at,
			updated_at
		`,
		id,
		tenantID,
		branchID,
		name,
		description,
		price,
	).Scan(
		&service.ID,
		&service.TenantID,
		&service.BranchID,
		&service.Name,
		&service.Description,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrServiceAlreadyExists
		}

		return nil, err
	}

	return &service, nil
}

func (r *Repository) Get(
	ctx context.Context,
	id string,
	tenantID string,
	branchID string,
) (*Service, error) {
	var service Service

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			name,
			description,
			price::text,
			active,
			created_at,
			updated_at
		FROM services
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		`,
		id,
		tenantID,
		branchID,
	).Scan(
		&service.ID,
		&service.TenantID,
		&service.BranchID,
		&service.Name,
		&service.Description,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrServiceNotFound
	}

	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	branchID string,
	activeOnly bool,
	limit int,
	offset int,
) ([]Service, error) {
	query := `
		SELECT
			id,
			tenant_id,
			branch_id,
			name,
			description,
			price::text,
			active,
			created_at,
			updated_at
		FROM services
		WHERE tenant_id = $1
		  AND branch_id = $2
	`

	args := []interface{}{
		tenantID,
		branchID,
	}

	if activeOnly {
		query += ` AND active = TRUE`
	}

	query += `
		ORDER BY name ASC, created_at ASC
		LIMIT $3
		OFFSET $4
	`

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]Service, 0)

	for rows.Next() {
		var service Service

		err := rows.Scan(
			&service.ID,
			&service.TenantID,
			&service.BranchID,
			&service.Name,
			&service.Description,
			&service.Price,
			&service.Active,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *Repository) Count(
	ctx context.Context,
	tenantID string,
	branchID string,
	activeOnly bool,
) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM services
		WHERE tenant_id = $1
		  AND branch_id = $2
	`

	args := []interface{}{
		tenantID,
		branchID,
	}

	if activeOnly {
		query += ` AND active = TRUE`
	}

	var total int64

	err := r.db.QueryRow(
		ctx,
		query,
		args...,
	).Scan(&total)

	return total, err
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	tenantID string,
	branchID string,
	name string,
	description *string,
	price string,
) (*Service, error) {
	var service Service

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE services
		SET
			name = $4,
			description = $5,
			price = $6
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		RETURNING
			id,
			tenant_id,
			branch_id,
			name,
			description,
			price::text,
			active,
			created_at,
			updated_at
		`,
		id,
		tenantID,
		branchID,
		name,
		description,
		price,
	).Scan(
		&service.ID,
		&service.TenantID,
		&service.BranchID,
		&service.Name,
		&service.Description,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrServiceNotFound
	}

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrServiceAlreadyExists
		}

		return nil, err
	}

	return &service, nil
}

func (r *Repository) UpdateStatus(
	ctx context.Context,
	id string,
	tenantID string,
	branchID string,
	active bool,
) (*Service, error) {
	var service Service

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE services
		SET active = $4
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		RETURNING
			id,
			tenant_id,
			branch_id,
			name,
			description,
			price::text,
			active,
			created_at,
			updated_at
		`,
		id,
		tenantID,
		branchID,
		active,
	).Scan(
		&service.ID,
		&service.TenantID,
		&service.BranchID,
		&service.Name,
		&service.Description,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrServiceNotFound
	}

	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *Repository) CreateAuditLog(
	ctx context.Context,
	tenantID string,
	branchID string,
	userID string,
	userRole string,
	action string,
	entityType string,
	entityID string,
	oldValue []byte,
	newValue []byte,
	reason *string,
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
			old_value,
			new_value,
			reason
		)
		VALUES (
			gen_random_uuid(),
			$1,
			$2,
			$3,
			$4::user_role,
			$5,
			$6,
			$7,
			$8::jsonb,
			$9::jsonb,
			$10
		)
		`,
		tenantID,
		branchID,
		userID,
		userRole,
		action,
		entityType,
		entityID,
		oldValue,
		newValue,
		reason,
	)

	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return strings.Contains(
		strings.ToLower(err.Error()),
		"duplicate key",
	)
}
