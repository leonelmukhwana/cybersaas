package sale

import (
	"context"
	"errors"

	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSaleNotFound       = errors.New("sale not found")
	ErrBranchNotFound     = errors.New("branch not found")
	ErrBranchAccessDenied = errors.New("branch access denied")
	ErrSaleAlreadyVoided  = errors.New("sale already voided")
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
	tenantID string,
	request CreateSaleRequest,
	subtotal string,
	discountAmount string,
	totalAmount string,
	currency string,
) (*Sale, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// The client generates the sale ID.
	// This allows the same local sale to keep the same ID
	// when synchronization is retried.
	saleID := uuid.NewString()

	var sale Sale

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO sales (
			id,
			tenant_id,
			branch_id,
			customer_id,
			session_id,
			session_started_at,
			terminal_id,
			attendant_id,
			status,
			subtotal,
			discount_type,
			discount_value,
			discount_amount,
			total_amount,
			currency,
			client_operation_id
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
			'completed',
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15
		)
		ON CONFLICT (client_operation_id)
		DO UPDATE SET id = sales.id
		RETURNING
			id,
			tenant_id,
			branch_id,
			customer_id,
			session_id,
			session_started_at,
			terminal_id,
			attendant_id,
			status,
			subtotal::text,
			discount_type,
			discount_value::text,
			discount_amount::text,
			total_amount::text,
			currency,
			created_at,
			updated_at
		`,
		saleID,
		tenantID,
		request.BranchID,
		request.CustomerID,
		request.SessionID,
		request.SessionStartedAt,
		request.TerminalID,
		request.AttendantID,
		subtotal,
		request.DiscountType,
		request.DiscountValue,
		discountAmount,
		totalAmount,
		currency,
		request.ClientOperationID,
	).Scan(
		&sale.ID,
		&sale.TenantID,
		&sale.BranchID,
		&sale.CustomerID,
		&sale.SessionID,
		&sale.SessionStartedAt,
		&sale.TerminalID,
		&sale.AttendantID,
		&sale.Status,
		&sale.Subtotal,
		&sale.DiscountType,
		&sale.DiscountValue,
		&sale.DiscountAmount,
		&sale.TotalAmount,
		&sale.Currency,
		&sale.CreatedAt,
		&sale.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Only create sale items when this sale is actually new.
	// If client_operation_id already existed, the sale already
	// exists on the server and its items must not be duplicated.
	var itemCount int

	err = tx.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM sale_items
		WHERE sale_id = $1
		`,
		sale.ID,
	).Scan(&itemCount)

	if err != nil {
		return nil, err
	}

	if itemCount == 0 {
		for _, item := range request.Items {
			itemID := uuid.NewString()

			_, err = tx.Exec(
				ctx,
				`
				INSERT INTO sale_items (
					id,
					sale_id,
					service_id,
					description,
					quantity,
					unit_price,
					line_total
				)
				VALUES (
					$1,
					$2,
					$3,
					$4,
					$5,
					$6,
					$7
				)
				`,
				itemID,
				sale.ID,
				item.ServiceID,
				item.Description,
				item.Quantity,
				item.UnitPrice,
				calculateLineTotal(item.Quantity, item.UnitPrice),
			)

			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	sale.Items, err = r.GetItems(ctx, sale.ID)
	if err != nil {
		return nil, err
	}

	return &sale, nil
}
func (r *Repository) Get(
	ctx context.Context,
	tenantID string,
	branchID string,
	saleID string,
) (*Sale, error) {
	var sale Sale

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_id,
			session_id,
			session_started_at,
			terminal_id,
			attendant_id,
			status,
			subtotal::text,
			discount_type,
			discount_value::text,
			discount_amount::text,
			total_amount::text,
			currency,
			created_at,
			updated_at
		FROM sales
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		`,
		saleID,
		tenantID,
		branchID,
	).Scan(
		&sale.ID,
		&sale.TenantID,
		&sale.BranchID,
		&sale.CustomerID,
		&sale.SessionID,
		&sale.SessionStartedAt,
		&sale.TerminalID,
		&sale.AttendantID,
		&sale.Status,
		&sale.Subtotal,
		&sale.DiscountType,
		&sale.DiscountValue,
		&sale.DiscountAmount,
		&sale.TotalAmount,
		&sale.Currency,
		&sale.CreatedAt,
		&sale.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSaleNotFound
	}

	if err != nil {
		return nil, err
	}

	sale.Items, err = r.GetItems(ctx, sale.ID)
	if err != nil {
		return nil, err
	}

	return &sale, nil
}

func (r *Repository) GetItems(
	ctx context.Context,
	saleID string,
) ([]SaleItem, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			sale_id,
			service_id,
			description,
			quantity,
			unit_price::text,
			line_total::text
		FROM sale_items
		WHERE sale_id = $1
		ORDER BY id
		`,
		saleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SaleItem, 0)

	for rows.Next() {
		var item SaleItem

		if err := rows.Scan(
			&item.ID,
			&item.SaleID,
			&item.ServiceID,
			&item.Description,
			&item.Quantity,
			&item.UnitPrice,
			&item.LineTotal,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	branchID string,
	status string,
	limit int,
	offset int,
) ([]Sale, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			customer_id,
			session_id,
			session_started_at,
			terminal_id,
			attendant_id,
			status,
			subtotal::text,
			discount_type,
			discount_value::text,
			discount_amount::text,
			total_amount::text,
			currency,
			created_at,
			updated_at
		FROM sales
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND ($3 = '' OR status = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
		`,
		tenantID,
		branchID,
		status,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := make([]Sale, 0)

	for rows.Next() {
		var sale Sale

		if err := rows.Scan(
			&sale.ID,
			&sale.TenantID,
			&sale.BranchID,
			&sale.CustomerID,
			&sale.SessionID,
			&sale.SessionStartedAt,
			&sale.TerminalID,
			&sale.AttendantID,
			&sale.Status,
			&sale.Subtotal,
			&sale.DiscountType,
			&sale.DiscountValue,
			&sale.DiscountAmount,
			&sale.TotalAmount,
			&sale.Currency,
			&sale.CreatedAt,
			&sale.UpdatedAt,
		); err != nil {
			return nil, err
		}

		sales = append(sales, sale)
	}

	return sales, rows.Err()
}

func (r *Repository) Count(
	ctx context.Context,
	tenantID string,
	branchID string,
	status string,
) (int64, error) {
	var total int64

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM sales
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND ($3 = '' OR status = $3)
		`,
		tenantID,
		branchID,
		status,
	).Scan(&total)

	return total, err
}

func (r *Repository) Void(
	ctx context.Context,
	tenantID string,
	branchID string,
	saleID string,
) error {
	commandTag, err := r.db.Exec(
		ctx,
		`
		UPDATE sales
		SET
			status = 'voided',
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND status <> 'voided'
		`,
		saleID,
		tenantID,
		branchID,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		var status string

		err := r.db.QueryRow(
			ctx,
			`
			SELECT status
			FROM sales
			WHERE id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
			`,
			saleID,
			tenantID,
			branchID,
		).Scan(&status)

		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSaleNotFound
		}

		if err != nil {
			return err
		}

		if status == "voided" {
			return ErrSaleAlreadyVoided
		}
	}

	return nil
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
			$4,
			$5,
			'sale',
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

func calculateLineTotal(quantity int, unitPrice string) string {
	price := new(big.Float)

	_, ok := price.SetString(unitPrice)
	if !ok {
		return "0.00"
	}

	qty := new(big.Float).SetInt64(int64(quantity))

	total := new(big.Float).Mul(price, qty)

	return total.Text('f', 2)
}
