package customer

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrBranchNotFound        = errors.New("branch not found")
	ErrBranchAccessDenied    = errors.New("branch access denied")
	ErrCustomerAlreadyExists = errors.New("customer already registered in this branch")
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
		return ErrBranchNotFound
	}

	return nil
}

func (r *Repository) VerifyAttendantBranchAccess(
	ctx context.Context,
	tenantID string,
	userID string,
	branchID string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM attendant_branch_assignments aba
			INNER JOIN users u
				ON u.id = aba.user_id
			INNER JOIN branches b
				ON b.id = aba.branch_id
			WHERE aba.user_id = $1
			  AND aba.tenant_id = $2
			  AND aba.branch_id = $3
			  AND aba.unassigned_at IS NULL
			  AND u.role = 'attendant'
			  AND u.status = 'active'
			  AND b.status = 'active'
			  AND b.tenant_id = $2
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

func (r *Repository) BeginTx(
	ctx context.Context,
) (pgx.Tx, error) {
	return r.db.BeginTx(
		ctx,
		pgx.TxOptions{},
	)
}

func (r *Repository) Create(
	ctx context.Context,
	customer Customer,
	idNumberHash *string,
	encryptedID *string,
	parentIDHash *string,
	encryptedParentID *string,
) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := r.CreateTx(
		ctx,
		tx,
		customer,
		idNumberHash,
		encryptedID,
		parentIDHash,
		encryptedParentID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	customer Customer,
	idNumberHash *string,
	encryptedID *string,
	parentIDHash *string,
	encryptedParentID *string,
) error {
	_, err := tx.Exec(
		ctx,
		`
		INSERT INTO customers (
			id,
			tenant_id,
			branch_id,
			customer_type,
			full_name,
			phone,
			id_number_encrypted,
			id_number_hash,
			parent_name,
			parent_phone,
			parent_id_number_encrypted,
			parent_id_number_hash
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12
		)
		`,
		customer.ID,
		customer.TenantID,
		customer.BranchID,
		customer.CustomerType,
		customer.FullName,
		customer.Phone,
		encryptedID,
		idNumberHash,
		customer.ParentName,
		customer.ParentPhone,
		encryptedParentID,
		parentIDHash,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrCustomerAlreadyExists
		}

		return err
	}

	return nil
}

func (r *Repository) Get(
	ctx context.Context,
	customerID string,
	tenantID string,
	branchID string,
) (*Customer, error) {
	var customer Customer

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_type,
			full_name,
			phone,
			created_at,
			updated_at
		FROM customers
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		`,
		customerID,
		tenantID,
		branchID,
	).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.BranchID,
		&customer.CustomerType,
		&customer.FullName,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *Repository) GetTx(
	ctx context.Context,
	tx pgx.Tx,
	customerID string,
	tenantID string,
	branchID string,
) (*Customer, error) {
	var customer Customer

	err := tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_type,
			full_name,
			phone,
			created_at,
			updated_at
		FROM customers
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		`,
		customerID,
		tenantID,
		branchID,
	).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.BranchID,
		&customer.CustomerType,
		&customer.FullName,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	branchID string,
	search string,
	limit int,
	offset int,
) ([]Customer, int64, error) {
	search = strings.TrimSpace(search)

	var total int64

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM customers
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND (
				$3 = ''
				OR full_name ILIKE '%' || $3 || '%'
				OR phone ILIKE '%' || $3 || '%'
		  )
		`,
		tenantID,
		branchID,
		search,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_type,
			full_name,
			phone,
			created_at,
			updated_at
		FROM customers
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND (
				$3 = ''
				OR full_name ILIKE '%' || $3 || '%'
				OR phone ILIKE '%' || $3 || '%'
		  )
		ORDER BY created_at DESC
		LIMIT $4
		OFFSET $5
		`,
		tenantID,
		branchID,
		search,
		limit,
		offset,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	customers := make([]Customer, 0)

	for rows.Next() {
		var customer Customer

		if err := rows.Scan(
			&customer.ID,
			&customer.TenantID,
			&customer.BranchID,
			&customer.CustomerType,
			&customer.FullName,
			&customer.Phone,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *Repository) Update(
	ctx context.Context,
	customer Customer,
	idNumberHash *string,
	encryptedID *string,
	parentIDHash *string,
	encryptedParentID *string,
) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := r.UpdateTx(
		ctx,
		tx,
		customer,
		idNumberHash,
		encryptedID,
		parentIDHash,
		encryptedParentID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) UpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	customer Customer,
	idNumberHash *string,
	encryptedID *string,
	parentIDHash *string,
	encryptedParentID *string,
) error {
	result, err := tx.Exec(
		ctx,
		`
		UPDATE customers
		SET
			customer_type = $1,
			full_name = $2,
			phone = $3,

			id_number_encrypted =
				COALESCE($4, id_number_encrypted),

			id_number_hash =
				COALESCE($5, id_number_hash),

			parent_name = $6,
			parent_phone = $7,

			parent_id_number_encrypted =
				COALESCE($8, parent_id_number_encrypted),

			parent_id_number_hash =
				COALESCE($9, parent_id_number_hash)

		WHERE id = $10
		  AND tenant_id = $11
		  AND branch_id = $12
		`,
		customer.CustomerType,
		customer.FullName,
		customer.Phone,
		encryptedID,
		idNumberHash,
		customer.ParentName,
		customer.ParentPhone,
		encryptedParentID,
		parentIDHash,
		customer.ID,
		customer.TenantID,
		customer.BranchID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(
		err.Error(),
		"duplicate key",
	)
}

func (r *Repository) FindByIDHash(
	ctx context.Context,
	tenantID string,
	branchID string,
	idNumberHash string,
) (*Customer, error) {
	var customer Customer

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_type,
			full_name,
			phone,
			created_at,
			updated_at
		FROM customers
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND id_number_hash = $3
		LIMIT 1
		`,
		tenantID,
		branchID,
		idNumberHash,
	).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.BranchID,
		&customer.CustomerType,
		&customer.FullName,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *Repository) FindByIDHashTx(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	branchID string,
	idNumberHash string,
) (*Customer, error) {
	var customer Customer

	err := tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_type,
			full_name,
			phone,
			created_at,
			updated_at
		FROM customers
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND id_number_hash = $3
		LIMIT 1
		`,
		tenantID,
		branchID,
		idNumberHash,
	).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.BranchID,
		&customer.CustomerType,
		&customer.FullName,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *Repository) FindTerminalCustomerByIDHash(
	ctx context.Context,
	tenantID string,
	branchID string,
	idNumberHash string,
) (*TerminalCustomer, error) {
	var customer TerminalCustomer

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			branch_id,
			customer_type,
			full_name,
			phone,
			id_number_hash,
			updated_at
		FROM customers
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND id_number_hash = $3
		  AND customer_type = 'adult'
		LIMIT 1
		`,
		tenantID,
		branchID,
		idNumberHash,
	).Scan(
		&customer.ID,
		&customer.BranchID,
		&customer.CustomerType,
		&customer.FullName,
		&customer.Phone,
		&customer.IDNumberHash,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *Repository) FindTerminalChildByName(
	ctx context.Context,
	tenantID string,
	branchID string,
	fullName string,
) (*TerminalCustomer, error) {
	fullName = strings.TrimSpace(fullName)

	if tenantID == "" ||
		branchID == "" ||
		fullName == "" {
		return nil, ErrCustomerNotFound
	}

	var customer TerminalCustomer

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			branch_id,
			customer_type,
			full_name,
			phone,
			id_number_hash,
			parent_name,
			updated_at
		FROM customers
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND customer_type = 'child'
		  AND full_name ILIKE $3
		ORDER BY
			CASE
				WHEN LOWER(full_name) = LOWER($4)
				THEN 0
				ELSE 1
			END,
			full_name ASC
		LIMIT 1
		`,
		tenantID,
		branchID,
		"%"+fullName+"%",
		fullName,
	).Scan(
		&customer.ID,
		&customer.BranchID,
		&customer.CustomerType,
		&customer.FullName,
		&customer.Phone,
		&customer.IDNumberHash,
		&customer.ParentName,
		&customer.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &customer, nil
}
