package receipt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrReceiptNotFound      = errors.New("receipt not found")
	ErrSaleNotFound         = errors.New("sale not found")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentNotConfirmed  = errors.New("payment is not confirmed")
	ErrReceiptAlreadyExists = errors.New("receipt already exists")
	ErrBranchAccessDenied   = errors.New("branch access denied")
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
	branchID string,
	saleID string,
	paymentID *string,
) (*Receipt, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	/*
			Serialize receipt-number generation per branch.

		PostgreSQL advisory locks are transaction-scoped, so the
		lock is automatically released when this transaction commits
		or rolls back.
	*/
	_, err = tx.Exec(
		ctx,
		`
		SELECT pg_advisory_xact_lock(
			hashtextextended($1, 0)
		)
		`,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	var paymentStatus string
	var paymentAmount string

	if paymentID != nil {
		err = tx.QueryRow(
			ctx,
			`
			SELECT
				status,
				amount::text
			FROM payments
			WHERE id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
			  AND sale_id = $4
			FOR UPDATE
			`,
			*paymentID,
			tenantID,
			branchID,
			saleID,
		).Scan(
			&paymentStatus,
			&paymentAmount,
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}

		if err != nil {
			return nil, err
		}

		if paymentStatus != "confirmed" {
			return nil, ErrPaymentNotConfirmed
		}
	}

	/*
		One primary receipt per sale.
	*/
	var existing Receipt

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		FROM receipts
		WHERE sale_id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		`,
		saleID,
		tenantID,
		branchID,
	).Scan(
		&existing.ID,
		&existing.TenantID,
		&existing.BranchID,
		&existing.SaleID,
		&existing.PaymentID,
		&existing.ReceiptNumber,
		&existing.ReceiptType,
		&existing.IssuedAt,
		&existing.PrintedAt,
		&existing.ReprintCount,
		&existing.CreatedAt,
	)

	if err == nil {
		return &existing, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var saleExists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM sales
			WHERE id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
			  AND status <> 'voided'
		)
		`,
		saleID,
		tenantID,
		branchID,
	).Scan(&saleExists)

	if err != nil {
		return nil, err
	}

	if !saleExists {
		return nil, ErrSaleNotFound
	}

	var sequence int64

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			COALESCE(
				MAX(
					CASE
						WHEN receipt_number ~ '^RCP-[0-9]{8}-[0-9]+$'
						THEN split_part(receipt_number, '-', 3)::BIGINT
						ELSE 0
					END
				),
				0
			) + 1
		FROM receipts
		WHERE branch_id = $1
		`,
		branchID,
	).Scan(&sequence)

	if err != nil {
		return nil, err
	}

	receiptNumber := fmt.Sprintf(
		"RCP-%s-%06d",
		currentDate(),
		sequence,
	)

	receiptID := uuid.NewString()

	var receipt Receipt

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO receipts (
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			'sale'
		)
		RETURNING
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		`,
		receiptID,
		tenantID,
		branchID,
		saleID,
		paymentID,
		receiptNumber,
	).Scan(
		&receipt.ID,
		&receipt.TenantID,
		&receipt.BranchID,
		&receipt.SaleID,
		&receipt.PaymentID,
		&receipt.ReceiptNumber,
		&receipt.ReceiptType,
		&receipt.IssuedAt,
		&receipt.PrintedAt,
		&receipt.ReprintCount,
		&receipt.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrReceiptAlreadyExists
		}

		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	_ = paymentAmount

	return &receipt, nil
}

func (r *Repository) Get(
	ctx context.Context,
	tenantID string,
	branchID string,
	receiptID string,
) (*ReceiptWithSale, error) {
	var receipt ReceiptWithSale

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			r.id,
			r.tenant_id,
			r.branch_id,
			r.sale_id,
			r.payment_id,
			r.receipt_number,
			r.receipt_type,
			r.issued_at,
			r.printed_at,
			r.reprint_count,
			r.created_at,
			s.total_amount::text,
			COALESCE(p.amount::text, '0'),
			COALESCE(p.method::text, ''),
			COALESCE(p.status::text, '')
		FROM receipts r
		JOIN sales s
			ON s.id = r.sale_id
		LEFT JOIN payments p
			ON p.id = r.payment_id
		WHERE r.id = $1
		  AND r.tenant_id = $2
		  AND r.branch_id = $3
		`,
		receiptID,
		tenantID,
		branchID,
	).Scan(
		&receipt.ID,
		&receipt.TenantID,
		&receipt.BranchID,
		&receipt.SaleID,
		&receipt.PaymentID,
		&receipt.ReceiptNumber,
		&receipt.ReceiptType,
		&receipt.IssuedAt,
		&receipt.PrintedAt,
		&receipt.ReprintCount,
		&receipt.CreatedAt,
		&receipt.SaleTotal,
		&receipt.PaymentAmount,
		&receipt.PaymentMethod,
		&receipt.PaymentStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReceiptNotFound
	}

	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	branchID string,
	limit int,
	offset int,
) ([]Receipt, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		FROM receipts
		WHERE tenant_id = $1
		  AND branch_id = $2
		ORDER BY issued_at DESC
		LIMIT $3 OFFSET $4
		`,
		tenantID,
		branchID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	receipts := make([]Receipt, 0)

	for rows.Next() {
		var receipt Receipt

		if err := rows.Scan(
			&receipt.ID,
			&receipt.TenantID,
			&receipt.BranchID,
			&receipt.SaleID,
			&receipt.PaymentID,
			&receipt.ReceiptNumber,
			&receipt.ReceiptType,
			&receipt.IssuedAt,
			&receipt.PrintedAt,
			&receipt.ReprintCount,
			&receipt.CreatedAt,
		); err != nil {
			return nil, err
		}

		receipts = append(receipts, receipt)
	}

	return receipts, rows.Err()
}

func (r *Repository) Count(
	ctx context.Context,
	tenantID string,
	branchID string,
) (int64, error) {
	var total int64

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM receipts
		WHERE tenant_id = $1
		  AND branch_id = $2
		`,
		tenantID,
		branchID,
	).Scan(&total)

	return total, err
}

func (r *Repository) MarkPrinted(
	ctx context.Context,
	tenantID string,
	branchID string,
	receiptID string,
) (*Receipt, error) {
	var receipt Receipt

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE receipts
		SET printed_at = COALESCE(printed_at, NOW())
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		RETURNING
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		`,
		receiptID,
		tenantID,
		branchID,
	).Scan(
		&receipt.ID,
		&receipt.TenantID,
		&receipt.BranchID,
		&receipt.SaleID,
		&receipt.PaymentID,
		&receipt.ReceiptNumber,
		&receipt.ReceiptType,
		&receipt.IssuedAt,
		&receipt.PrintedAt,
		&receipt.ReprintCount,
		&receipt.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReceiptNotFound
	}

	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r *Repository) Reprint(
	ctx context.Context,
	tenantID string,
	branchID string,
	receiptID string,
	userID string,
	terminalID *string,
	reason *string,
) (*Receipt, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var receipt Receipt

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		FROM receipts
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		FOR UPDATE
		`,
		receiptID,
		tenantID,
		branchID,
	).Scan(
		&receipt.ID,
		&receipt.TenantID,
		&receipt.BranchID,
		&receipt.SaleID,
		&receipt.PaymentID,
		&receipt.ReceiptNumber,
		&receipt.ReceiptType,
		&receipt.IssuedAt,
		&receipt.PrintedAt,
		&receipt.ReprintCount,
		&receipt.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReceiptNotFound
	}

	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO receipt_reprints (
			id,
			receipt_id,
			user_id,
			terminal_id,
			reason
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
		`,
		uuid.NewString(),
		receiptID,
		userID,
		terminalID,
		reason,
	)
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(
		ctx,
		`
		UPDATE receipts
		SET
			reprint_count = reprint_count + 1,
			printed_at = COALESCE(printed_at, NOW())
		WHERE id = $1
		RETURNING
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		`,
		receiptID,
	).Scan(
		&receipt.ID,
		&receipt.TenantID,
		&receipt.BranchID,
		&receipt.SaleID,
		&receipt.PaymentID,
		&receipt.ReceiptNumber,
		&receipt.ReceiptType,
		&receipt.IssuedAt,
		&receipt.PrintedAt,
		&receipt.ReprintCount,
		&receipt.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &receipt, nil
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
			'receipt',
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

func currentDate() string {
	return time.Now().Format("20060102")
}
