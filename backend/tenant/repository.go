package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("owner already has a tenant")
	ErrOwnerNotFound       = errors.New("owner not found")
	ErrNotOwner            = errors.New("user is not a cyber owner")
	ErrTenantAttachFailed  = errors.New("failed to attach tenant to owner")
)

type TrialSubscriptionCreator interface {
	CreateTrialSubscriptionTx(
		ctx context.Context,
		tx pgx.Tx,
		tenantID string,
		startedAt time.Time,
	) error
}

type Repository struct {
	db                       *pgxpool.Pool
	trialSubscriptionCreator TrialSubscriptionCreator
}

func NewRepository(
	db *pgxpool.Pool,
	trialSubscriptionCreator TrialSubscriptionCreator,
) *Repository {
	return &Repository{
		db:                       db,
		trialSubscriptionCreator: trialSubscriptionCreator,
	}
}

// ------------------------------------------------------------
// CREATE TENANT
// ------------------------------------------------------------

func (r *Repository) CreateTenant(
	ctx context.Context,
	userID string,
	businessName string,
) (Tenant, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Tenant{}, err
	}

	defer tx.Rollback(ctx)

	// Capture the trial start time once so the tenant and its
	// trial subscription use the same starting point.
	trialStart := time.Now().UTC()

	// --------------------------------------------------------
	// Get and lock owner
	// --------------------------------------------------------

	var (
		ownerName string
		phone     string
		email     *string
		role      string
		tenantID  *string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			full_name,
			COALESCE(phone, ''),
			email,
			role::text,
			tenant_id
		FROM users
		WHERE id = $1
		FOR UPDATE
		`,
		userID,
	).Scan(
		&ownerName,
		&phone,
		&email,
		&role,
		&tenantID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrOwnerNotFound
	}

	if err != nil {
		return Tenant{}, err
	}

	if role != "owner" {
		return Tenant{}, ErrNotOwner
	}

	// One Cyber Owner can have only one tenant.
	if tenantID != nil && *tenantID != "" {
		return Tenant{}, ErrTenantAlreadyExists
	}

	// --------------------------------------------------------
	// Create tenant
	// --------------------------------------------------------

	newTenantID := uuid.New()

	var tenant Tenant

	err = tx.QueryRow(
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
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			'active'
		)
		RETURNING
			id,
			business_name,
			owner_name,
			phone,
			email,
			status,
			created_at
		`,
		newTenantID,
		businessName,
		ownerName,
		phone,
		email,
	).Scan(
		&tenant.ID,
		&tenant.BusinessName,
		&tenant.OwnerName,
		&tenant.Phone,
		&tenant.Email,
		&tenant.Status,
		&tenant.CreatedAt,
	)

	if err != nil {
		return Tenant{}, err
	}

	// --------------------------------------------------------
	// Attach tenant to owner
	// --------------------------------------------------------

	result, err := tx.Exec(
		ctx,
		`
		UPDATE users
		SET
			tenant_id = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND role = 'owner'
		  AND tenant_id IS NULL
		`,
		newTenantID,
		userID,
	)

	if err != nil {
		return Tenant{}, err
	}

	if result.RowsAffected() != 1 {
		return Tenant{}, ErrTenantAttachFailed
	}

	// --------------------------------------------------------
	// Create 7-day trial subscription
	// --------------------------------------------------------

	if r.trialSubscriptionCreator == nil {
		return Tenant{}, errors.New(
			"trial subscription creator is not configured",
		)
	}

	if err := r.trialSubscriptionCreator.CreateTrialSubscriptionTx(
		ctx,
		tx,
		newTenantID.String(),
		trialStart,
	); err != nil {
		return Tenant{}, fmt.Errorf(
			"failed to create trial subscription: %w",
			err,
		)
	}

	// --------------------------------------------------------
	// Commit
	// --------------------------------------------------------

	if err := tx.Commit(ctx); err != nil {
		return Tenant{}, err
	}

	return tenant, nil
}

// ------------------------------------------------------------
// GET TENANT BY OWNER
// ------------------------------------------------------------

func (r *Repository) GetTenantByOwnerID(
	ctx context.Context,
	userID string,
) (Tenant, error) {

	var tenant Tenant

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			t.id,
			t.business_name,
			t.owner_name,
			t.phone,
			t.email,
			t.status,
			t.created_at
		FROM tenants t
		INNER JOIN users u
			ON u.tenant_id = t.id
		WHERE u.id = $1
		  AND u.role = 'owner'
		LIMIT 1
		`,
		userID,
	).Scan(
		&tenant.ID,
		&tenant.BusinessName,
		&tenant.OwnerName,
		&tenant.Phone,
		&tenant.Email,
		&tenant.Status,
		&tenant.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrTenantNotFound
	}

	if err != nil {
		return Tenant{}, err
	}

	return tenant, nil
}

// ------------------------------------------------------------
// GET TENANT BY ID
// ------------------------------------------------------------

func (r *Repository) GetTenantByID(
	ctx context.Context,
	tenantID string,
) (Tenant, error) {

	var tenant Tenant

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			business_name,
			owner_name,
			phone,
			email,
			status,
			created_at
		FROM tenants
		WHERE id = $1
		`,
		tenantID,
	).Scan(
		&tenant.ID,
		&tenant.BusinessName,
		&tenant.OwnerName,
		&tenant.Phone,
		&tenant.Email,
		&tenant.Status,
		&tenant.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrTenantNotFound
	}

	if err != nil {
		return Tenant{}, err
	}

	return tenant, nil
}

// ------------------------------------------------------------
// UPDATE TENANT
// ------------------------------------------------------------

func (r *Repository) UpdateTenant(
	ctx context.Context,
	userID string,
	businessName string,
) (Tenant, error) {

	var tenant Tenant

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE tenants t
		SET
			business_name = $1
		FROM users u
		WHERE u.id = $2
		  AND u.role = 'owner'
		  AND u.tenant_id = t.id
		RETURNING
			t.id,
			t.business_name,
			t.owner_name,
			t.phone,
			t.email,
			t.status,
			t.created_at
		`,
		businessName,
		userID,
	).Scan(
		&tenant.ID,
		&tenant.BusinessName,
		&tenant.OwnerName,
		&tenant.Phone,
		&tenant.Email,
		&tenant.Status,
		&tenant.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrTenantNotFound
	}

	if err != nil {
		return Tenant{}, err
	}

	return tenant, nil
}
