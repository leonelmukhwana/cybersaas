package payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrSaleNotFound         = errors.New("sale not found")
	ErrBranchAccessDenied   = errors.New("branch access denied")
	ErrPaymentAlreadyFinal  = errors.New("payment is already in a final state")
	ErrDuplicatePayment     = errors.New("duplicate payment operation")
	ErrPaymentReferenceUsed = errors.New("payment reference already exists")
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

func (r *Repository) VerifySaleBelongsToBranch(
	ctx context.Context,
	tenantID string,
	branchID string,
	saleID string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM sales
			WHERE id = $1
			  AND tenant_id = $2
			  AND branch_id = $3
		)
		`,
		saleID,
		tenantID,
		branchID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return ErrSaleNotFound
	}

	return nil
}

func (r *Repository) GetSaleTotal(
	ctx context.Context,
	tenantID string,
	branchID string,
	saleID string,
) (string, error) {
	var total string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT total_amount::text
		FROM sales
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND status <> 'voided'
		`,
		saleID,
		tenantID,
		branchID,
	).Scan(&total)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSaleNotFound
	}

	return total, err
}

func (r *Repository) GetConfirmedPaymentTotal(
	ctx context.Context,
	tenantID string,
	branchID string,
	saleID string,
) (string, error) {
	var total string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(
			SUM(amount),
			0
		)::text
		FROM payments
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND sale_id = $3
		  AND status = 'confirmed'
		`,
		tenantID,
		branchID,
		saleID,
	).Scan(&total)

	return total, err
}

func (r *Repository) Create(
	ctx context.Context,
	tenantID string,
	request CreatePaymentRequest,
) (*Payment, error) {
	id := newUUID()

	var payment Payment

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO payments (
			id,
			tenant_id,
			branch_id,
			sale_id,
			method,
			status,
			amount,
			phone,
			external_reference,
			client_operation_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5::payment_method,
			'pending',
			$6,
			$7,
			$8,
			$9
		)
		RETURNING
			id,
			tenant_id,
			branch_id,
			sale_id,
			method,
			status,
			amount::text,
			phone,
			external_reference,
			mpesa_receipt_number,
			provider_request_id,
			provider_transaction_id,
			failure_reason,
			confirmed_at,
			client_operation_id,
			created_at,
			updated_at
		`,
		id,
		tenantID,
		request.BranchID,
		request.SaleID,
		request.Method,
		request.Amount,
		request.Phone,
		request.ExternalReference,
		request.ClientOperationID,
	).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.BranchID,
		&payment.SaleID,
		&payment.Method,
		&payment.Status,
		&payment.Amount,
		&payment.Phone,
		&payment.ExternalReference,
		&payment.MPesaReceiptNumber,
		&payment.ProviderRequestID,
		&payment.ProviderTransactionID,
		&payment.FailureReason,
		&payment.ConfirmedAt,
		&payment.ClientOperationID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicatePayment
		}

		return nil, err
	}

	return &payment, nil
}

func (r *Repository) Get(
	ctx context.Context,
	tenantID string,
	branchID string,
	paymentID string,
) (*Payment, error) {
	var payment Payment

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			method,
			status,
			amount::text,
			phone,
			external_reference,
			mpesa_receipt_number,
			provider_request_id,
			provider_transaction_id,
			failure_reason,
			confirmed_at,
			client_operation_id,
			created_at,
			updated_at
		FROM payments
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		`,
		paymentID,
		tenantID,
		branchID,
	).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.BranchID,
		&payment.SaleID,
		&payment.Method,
		&payment.Status,
		&payment.Amount,
		&payment.Phone,
		&payment.ExternalReference,
		&payment.MPesaReceiptNumber,
		&payment.ProviderRequestID,
		&payment.ProviderTransactionID,
		&payment.FailureReason,
		&payment.ConfirmedAt,
		&payment.ClientOperationID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	branchID string,
	status string,
	limit int,
	offset int,
) ([]Payment, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			method,
			status,
			amount::text,
			phone,
			external_reference,
			mpesa_receipt_number,
			provider_request_id,
			provider_transaction_id,
			failure_reason,
			confirmed_at,
			client_operation_id,
			created_at,
			updated_at
		FROM payments
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

	payments := make([]Payment, 0)

	for rows.Next() {
		var payment Payment

		if err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.BranchID,
			&payment.SaleID,
			&payment.Method,
			&payment.Status,
			&payment.Amount,
			&payment.Phone,
			&payment.ExternalReference,
			&payment.MPesaReceiptNumber,
			&payment.ProviderRequestID,
			&payment.ProviderTransactionID,
			&payment.FailureReason,
			&payment.ConfirmedAt,
			&payment.ClientOperationID,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, rows.Err()
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
		FROM payments
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

func (r *Repository) Confirm(
	ctx context.Context,
	tenantID string,
	branchID string,
	paymentID string,
	request ConfirmPaymentRequest,
) (*Payment, error) {
	var payment Payment

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE payments
		SET
			status = 'confirmed',
			external_reference = COALESCE($4, external_reference),
			mpesa_receipt_number = COALESCE($5, mpesa_receipt_number),
			provider_request_id = COALESCE($6, provider_request_id),
			provider_transaction_id = COALESCE($7, provider_transaction_id),
			confirmed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND status = 'pending'
		RETURNING
			id,
			tenant_id,
			branch_id,
			sale_id,
			method,
			status,
			amount::text,
			phone,
			external_reference,
			mpesa_receipt_number,
			provider_request_id,
			provider_transaction_id,
			failure_reason,
			confirmed_at,
			client_operation_id,
			created_at,
			updated_at
		`,
		paymentID,
		tenantID,
		branchID,
		request.ExternalReference,
		request.MPesaReceiptNumber,
		request.ProviderRequestID,
		request.ProviderTransactionID,
	).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.BranchID,
		&payment.SaleID,
		&payment.Method,
		&payment.Status,
		&payment.Amount,
		&payment.Phone,
		&payment.ExternalReference,
		&payment.MPesaReceiptNumber,
		&payment.ProviderRequestID,
		&payment.ProviderTransactionID,
		&payment.FailureReason,
		&payment.ConfirmedAt,
		&payment.ClientOperationID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentAlreadyFinal
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrPaymentReferenceUsed
		}

		return nil, err
	}

	return &payment, nil
}

func (r *Repository) Fail(
	ctx context.Context,
	tenantID string,
	branchID string,
	paymentID string,
	reason string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE payments
		SET
			status = 'failed',
			failure_reason = $4,
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND status = 'pending'
		`,
		paymentID,
		tenantID,
		branchID,
		reason,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		var exists bool

		err := r.db.QueryRow(
			ctx,
			`
			SELECT EXISTS (
				SELECT 1
				FROM payments
				WHERE id = $1
				  AND tenant_id = $2
				  AND branch_id = $3
			)
			`,
			paymentID,
			tenantID,
			branchID,
		).Scan(&exists)

		if err != nil {
			return err
		}

		if !exists {
			return ErrPaymentNotFound
		}

		return ErrPaymentAlreadyFinal
	}

	return nil
}

func (r *Repository) Refund(
	ctx context.Context,
	tenantID string,
	branchID string,
	paymentID string,
	reason string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE payments
		SET
			status = 'refunded',
			failure_reason = $4,
			updated_at = NOW()
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND status = 'confirmed'
		`,
		paymentID,
		tenantID,
		branchID,
		reason,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		var exists bool

		err := r.db.QueryRow(
			ctx,
			`
			SELECT EXISTS (
				SELECT 1
				FROM payments
				WHERE id = $1
				  AND tenant_id = $2
				  AND branch_id = $3
			)
			`,
			paymentID,
			tenantID,
			branchID,
		).Scan(&exists)

		if err != nil {
			return err
		}

		if !exists {
			return ErrPaymentNotFound
		}

		return ErrPaymentAlreadyFinal
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
			'payment',
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

func newUUID() string {
	return uuid.NewString()
}
