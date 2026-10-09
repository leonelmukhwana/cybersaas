package branch

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("branch not found")
	ErrAlreadyExists = errors.New("branch already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// -----------------------------------------------------------------------------
// TENANT
// -----------------------------------------------------------------------------

func (r *Repository) CreateTenant(
	ctx context.Context,
	tenantID string,
	businessName string,
	ownerName string,
	phone string,
	email string,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO tenants (
			id,
			business_name,
			owner_name,
			phone,
			email,
			status
		)
		VALUES ($1, $2, $3, $4, $5, 'active')
		`,
		tenantID,
		businessName,
		ownerName,
		phone,
		email,
	)

	return err
}

func (r *Repository) UpdateUserTenant(
	ctx context.Context,
	userID string,
	tenantID string,
) error {
	commandTag, err := r.db.Exec(
		ctx,
		`
		UPDATE users
		SET tenant_id = $1
		WHERE id = $2
		  AND role = 'owner'
		`,
		tenantID,
		userID,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) GetTenant(
	ctx context.Context,
	tenantID string,
) (
	businessName string,
	ownerName string,
	phone string,
	email *string,
	err error,
) {
	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			business_name,
			owner_name,
			phone,
			email
		FROM tenants
		WHERE id = $1
		`,
		tenantID,
	).Scan(
		&businessName,
		&ownerName,
		&phone,
		&email,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}

	return
}

func (r *Repository) GetUserOwnerDetails(
	ctx context.Context,
	userID string,
) (
	tenantID *string,
	fullName string,
	phone *string,
	email *string,
	err error,
) {
	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			tenant_id,
			full_name,
			phone,
			email
		FROM users
		WHERE id = $1
		  AND role = 'owner'
		`,
		userID,
	).Scan(
		&tenantID,
		&fullName,
		&phone,
		&email,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}

	return
}

// -----------------------------------------------------------------------------
// BRANCHES
// -----------------------------------------------------------------------------

func (r *Repository) CreateBranch(
	ctx context.Context,
	tenantID string,
	name string,
	address *string,
	phone *string,
) (Branch, error) {
	id := uuid.New().String()

	var b Branch

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO branches (
			id,
			tenant_id,
			name,
			address,
			phone,
			status
		)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING
			id,
			tenant_id,
			name,
			address,
			phone,
			status,
			created_at,
			updated_at
		`,
		id,
		tenantID,
		name,
		address,
		phone,
	).Scan(
		&b.ID,
		&b.TenantID,
		&b.Name,
		&b.Address,
		&b.Phone,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Branch{}, ErrAlreadyExists
		}

		return Branch{}, err
	}

	return b, nil
}

func (r *Repository) GetBranch(
	ctx context.Context,
	tenantID string,
	branchID string,
) (Branch, error) {
	var b Branch

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			name,
			address,
			phone,
			status,
			created_at,
			updated_at
		FROM branches
		WHERE id = $1
		  AND tenant_id = $2
		`,
		branchID,
		tenantID,
	).Scan(
		&b.ID,
		&b.TenantID,
		&b.Name,
		&b.Address,
		&b.Phone,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Branch{}, ErrNotFound
	}

	if err != nil {
		return Branch{}, err
	}

	return b, nil
}

func (r *Repository) ListBranches(
	ctx context.Context,
	tenantID string,
) ([]Branch, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			name,
			address,
			phone,
			status,
			created_at,
			updated_at
		FROM branches
		WHERE tenant_id = $1
		ORDER BY created_at ASC
		`,
		tenantID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	branches := make([]Branch, 0)

	for rows.Next() {
		var b Branch

		if err := rows.Scan(
			&b.ID,
			&b.TenantID,
			&b.Name,
			&b.Address,
			&b.Phone,
			&b.Status,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}

		branches = append(branches, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return branches, nil
}

func (r *Repository) CountBranches(
	ctx context.Context,
	tenantID string,
) (int64, error) {
	var count int64

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM branches
		WHERE tenant_id = $1
		`,
		tenantID,
	).Scan(&count)

	return count, err
}

func (r *Repository) UpdateBranch(
	ctx context.Context,
	tenantID string,
	branchID string,
	name string,
	address *string,
	phone *string,
) (Branch, error) {
	var b Branch

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE branches
		SET
			name = $1,
			address = $2,
			phone = $3
		WHERE id = $4
		  AND tenant_id = $5
		RETURNING
			id,
			tenant_id,
			name,
			address,
			phone,
			status,
			created_at,
			updated_at
		`,
		name,
		address,
		phone,
		branchID,
		tenantID,
	).Scan(
		&b.ID,
		&b.TenantID,
		&b.Name,
		&b.Address,
		&b.Phone,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Branch{}, ErrNotFound
	}

	if err != nil {
		if isUniqueViolation(err) {
			return Branch{}, ErrAlreadyExists
		}

		return Branch{}, err
	}

	return b, nil
}

func (r *Repository) UpdateBranchStatus(
	ctx context.Context,
	tenantID string,
	branchID string,
	status string,
) (Branch, error) {
	var b Branch

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE branches
		SET status = $1
		WHERE id = $2
		  AND tenant_id = $3
		RETURNING
			id,
			tenant_id,
			name,
			address,
			phone,
			status,
			created_at,
			updated_at
		`,
		status,
		branchID,
		tenantID,
	).Scan(
		&b.ID,
		&b.TenantID,
		&b.Name,
		&b.Address,
		&b.Phone,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Branch{}, ErrNotFound
	}

	if err != nil {
		return Branch{}, err
	}

	return b, nil
}

// -----------------------------------------------------------------------------
// AUDIT
// -----------------------------------------------------------------------------

func (r *Repository) CreateAuditLog(
	ctx context.Context,
	tenantID *string,
	branchID *string,
	userID string,
	userRole string,
	action string,
	entityType string,
	entityID string,
	oldValue interface{},
	newValue interface{},
	reason string,
	ipAddress *string,
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
			reason,
			ip_address
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
			$10,
			$11
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
		ipAddress,
	)

	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && containsCode(err.Error(), "23505")
}

func containsCode(message string, code string) bool {
	return len(message) >= len(code) &&
		containsSubstring(message, code)
}

func containsSubstring(s string, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}

func (r *Repository) GetBillingConfig(
	ctx context.Context,
	tenantID string,
	branchID string,
) (BillingConfig, error) {
	// Create the default configuration if the branch exists and
	// doesn't have one yet. This does not create anything for
	// a branch belonging to another tenant.
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO billing_configs (id, tenant_id, branch_id)
		SELECT gen_random_uuid(), $1, b.id
		FROM branches b
		WHERE b.id = $2
		  AND b.tenant_id = $1
		ON CONFLICT (branch_id) DO NOTHING
		`,
		tenantID,
		branchID,
	)
	if err != nil {
		return BillingConfig{}, err
	}

	var config BillingConfig

	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			bc.id,
			bc.tenant_id,
			bc.branch_id,
			bc.rate_per_minute,
			bc.minimum_charge,
			bc.billing_interval_minutes,
			bc.rounding_mode,
			bc.currency,
			bc.created_at,
			bc.updated_at
		FROM billing_configs bc
		INNER JOIN branches b
			ON b.id = bc.branch_id
		   AND b.tenant_id = bc.tenant_id
		WHERE bc.tenant_id = $1
		  AND bc.branch_id = $2
		`,
		tenantID,
		branchID,
	).Scan(
		&config.ID,
		&config.TenantID,
		&config.BranchID,
		&config.RatePerMinute,
		&config.MinimumCharge,
		&config.BillingIntervalMinutes,
		&config.RoundingMode,
		&config.Currency,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return BillingConfig{}, ErrNotFound
	}
	if err != nil {
		return BillingConfig{}, err
	}

	return config, nil
}

func (r *Repository) UpdateBillingConfig(
	ctx context.Context,
	tenantID string,
	branchID string,
	req UpdateBillingConfigRequest,
) (BillingConfig, error) {
	var config BillingConfig

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO billing_configs (
			id,
			tenant_id,
			branch_id,
			rate_per_minute,
			minimum_charge,
			billing_interval_minutes,
			rounding_mode,
			currency
		)
		SELECT
			gen_random_uuid(),
			$1,
			b.id,
			$3,
			$4,
			$5,
			$6,
			$7
		FROM branches b
		WHERE b.id = $2
		  AND b.tenant_id = $1
		ON CONFLICT (branch_id) DO UPDATE SET
			tenant_id = EXCLUDED.tenant_id,
			rate_per_minute = EXCLUDED.rate_per_minute,
			minimum_charge = EXCLUDED.minimum_charge,
			billing_interval_minutes = EXCLUDED.billing_interval_minutes,
			rounding_mode = EXCLUDED.rounding_mode,
			currency = EXCLUDED.currency,
			updated_at = NOW()
		RETURNING
			id,
			tenant_id,
			branch_id,
			rate_per_minute,
			minimum_charge,
			billing_interval_minutes,
			rounding_mode,
			currency,
			created_at,
			updated_at
		`,
		tenantID,
		branchID,
		req.RatePerMinute,
		req.MinimumCharge,
		req.BillingIntervalMinutes,
		req.RoundingMode,
		req.Currency,
	).Scan(
		&config.ID,
		&config.TenantID,
		&config.BranchID,
		&config.RatePerMinute,
		&config.MinimumCharge,
		&config.BillingIntervalMinutes,
		&config.RoundingMode,
		&config.Currency,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return BillingConfig{}, ErrNotFound
	}
	if err != nil {
		return BillingConfig{}, err
	}

	return config, nil
}

// Keep fmt imported for future PostgreSQL error handling and clearer
// repository diagnostics.
var _ = fmt.Sprintf
