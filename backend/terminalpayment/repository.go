package terminalpayment

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
	ErrSessionNotCompleted = errors.New("session is not completed")
	ErrPaymentNotFound     = errors.New("payment not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCashPayment(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	req CreateCashPaymentRequest,
) (*CreateCashPaymentResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// ------------------------------------------------------------
	// 1. Get and lock the completed session.
	// ------------------------------------------------------------

	var (
		customerID       *string
		sessionStartedAt time.Time
		sessionStatus    string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			customer_id,
			started_at,
			status
		FROM sessions
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND terminal_id = $4
		FOR UPDATE
		`,
		req.SessionID,
		tenantID,
		branchID,
		terminalID,
	).Scan(
		&customerID,
		&sessionStartedAt,
		&sessionStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	if sessionStatus != "completed" {
		return nil, ErrSessionNotCompleted
	}

	// ------------------------------------------------------------
	// 2. Make retries idempotent.
	// ------------------------------------------------------------

	var (
		existingPaymentID string
		existingSaleID    string
		existingAmount    string
		existingStatus    string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			sale_id,
			amount::text,
			status
		FROM payments
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND client_operation_id = $3
		LIMIT 1
		`,
		tenantID,
		branchID,
		req.ClientOperationID,
	).Scan(
		&existingPaymentID,
		&existingSaleID,
		&existingAmount,
		&existingStatus,
	)

	if err == nil {
		return &CreateCashPaymentResponse{
			PaymentID: existingPaymentID,
			SaleID:    existingSaleID,
			SessionID: req.SessionID,
			Amount:    existingAmount,
			Method:    "cash",
			Status:    existingStatus,
			Currency:  "KES",
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3. Create the sale.
	//
	// IMPORTANT:
	// The sales table does not contain a currency column.
	// Currency is handled by billing/payment configuration.
	// ------------------------------------------------------------

	saleID := uuid.NewString()

	_, err = tx.Exec(
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
			status,
			subtotal,
			discount_value,
			discount_amount,
			total_amount,
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
			'completed',
			$8,
			0,
			0,
			$8,
			$9
		)
		`,
		saleID,
		tenantID,
		branchID,
		customerID,
		req.SessionID,
		sessionStartedAt,
		terminalID,
		req.Amount,
		req.ClientOperationID,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 4. Create sale item.
	// ------------------------------------------------------------

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO sale_items (
			id,
			sale_id,
			description,
			quantity,
			unit_price,
			line_total
		)
		VALUES (
			$1,
			$2,
			'Computer session',
			1,
			$3,
			$3
		)
		`,
		uuid.NewString(),
		saleID,
		req.Amount,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 5. Create confirmed cash payment.
	// ------------------------------------------------------------

	paymentID := uuid.NewString()

	var confirmedAt time.Time

	err = tx.QueryRow(
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
			client_operation_id,
			confirmed_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'cash',
			'confirmed',
			$5,
			$6,
			NOW()
		)
		RETURNING confirmed_at
		`,
		paymentID,
		tenantID,
		branchID,
		saleID,
		req.Amount,
		req.ClientOperationID,
	).Scan(&confirmedAt)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 6. Generate receipt.
	// ------------------------------------------------------------

	receiptID := uuid.NewString()
	receiptNumber := "RCP-" + uuid.NewString()

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO receipts (
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			'sale',
			NOW()
		)
		`,
		receiptID,
		tenantID,
		branchID,
		saleID,
		paymentID,
		receiptNumber,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 7. Commit everything atomically.
	// ------------------------------------------------------------

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateCashPaymentResponse{
		PaymentID:   paymentID,
		SaleID:      saleID,
		SessionID:   req.SessionID,
		Amount:      req.Amount,
		Method:      "cash",
		Status:      "confirmed",
		Currency:    "KES",
		ConfirmedAt: &confirmedAt,
	}, nil
}

// CreateMpesaPayment creates the sale and a pending M-Pesa payment.
//
// Unlike cash payment, the payment is NOT confirmed here.
// It becomes confirmed only after the Safaricom callback is processed.
func (r *Repository) CreateMpesaPayment(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	req CreateMpesaPaymentRequest,
) (*CreateMpesaPaymentResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// ------------------------------------------------------------
	// 1. Get and lock the completed session.
	// ------------------------------------------------------------

	var (
		customerID       *string
		sessionStartedAt time.Time
		sessionStatus    string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			customer_id,
			started_at,
			status
		FROM sessions
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND terminal_id = $4
		FOR UPDATE
		`,
		req.SessionID,
		tenantID,
		branchID,
		terminalID,
	).Scan(
		&customerID,
		&sessionStartedAt,
		&sessionStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	if sessionStatus != "completed" {
		return nil, ErrSessionNotCompleted
	}

	// ------------------------------------------------------------
	// 2. Make retries idempotent.
	// ------------------------------------------------------------

	var (
		existingPaymentID string
		existingSaleID    string
		existingAmount    string
		existingStatus    string
		existingPhone     *string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			sale_id,
			amount::text,
			status,
			phone
		FROM payments
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND client_operation_id = $3
		LIMIT 1
		`,
		tenantID,
		branchID,
		req.ClientOperationID,
	).Scan(
		&existingPaymentID,
		&existingSaleID,
		&existingAmount,
		&existingStatus,
		&existingPhone,
	)

	if err == nil {
		response := &CreateMpesaPaymentResponse{
			PaymentID: existingPaymentID,
			SaleID:    existingSaleID,
			SessionID: req.SessionID,
			Amount:    existingAmount,
			Method:    "mpesa",
			Status:    existingStatus,
			Currency:  "KES",
		}

		if existingPhone != nil {
			response.PhoneNumber = *existingPhone
		}

		return response, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3. Create the sale.
	// ------------------------------------------------------------

	saleID := uuid.NewString()

	_, err = tx.Exec(
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
			status,
			subtotal,
			discount_value,
			discount_amount,
			total_amount,
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
			'completed',
			$8,
			0,
			0,
			$8,
			$9
		)
		`,
		saleID,
		tenantID,
		branchID,
		customerID,
		req.SessionID,
		sessionStartedAt,
		terminalID,
		req.Amount,
		req.ClientOperationID,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 4. Create sale item.
	// ------------------------------------------------------------

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO sale_items (
			id,
			sale_id,
			description,
			quantity,
			unit_price,
			line_total
		)
		VALUES (
			$1,
			$2,
			'Computer session',
			1,
			$3,
			$3
		)
		`,
		uuid.NewString(),
		saleID,
		req.Amount,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 5. Create PENDING M-Pesa payment.
	//
	// Do NOT create a receipt here.
	// The receipt is created only after successful payment.
	// ------------------------------------------------------------

	paymentID := uuid.NewString()

	_, err = tx.Exec(
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
			client_operation_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'mpesa',
			'pending',
			$5,
			$6,
			$7
		)
		`,
		paymentID,
		tenantID,
		branchID,
		saleID,
		req.Amount,
		req.PhoneNumber,
		req.ClientOperationID,
	)

	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 6. Commit.
	// ------------------------------------------------------------

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateMpesaPaymentResponse{
		PaymentID:   paymentID,
		SaleID:      saleID,
		SessionID:   req.SessionID,
		Amount:      req.Amount,
		Method:      "mpesa",
		Status:      "pending",
		Currency:    "KES",
		PhoneNumber: req.PhoneNumber,
	}, nil
}

func (r *Repository) GetMpesaPaymentStatus(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	paymentID string,
) (*MpesaPaymentStatusResponse, error) {
	var result MpesaPaymentStatusResponse

	var confirmedAt *time.Time

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			p.id,
			p.sale_id,
			s.session_id,
			p.amount::text,
			p.method,
			p.status,
			COALESCE(p.confirmed_at, NULL)
		FROM payments p
		INNER JOIN sales s
			ON s.id = p.sale_id
		   AND s.tenant_id = p.tenant_id
		   AND s.branch_id = p.branch_id
		INNER JOIN sessions s2
			ON s2.id = s.session_id
		   AND s2.tenant_id = s.tenant_id
		   AND s2.branch_id = s.branch_id
		   AND s2.terminal_id = $2
		WHERE p.id = $1
		  AND p.tenant_id = $3
		  AND p.branch_id = $4
		  AND p.method = 'mpesa'
		LIMIT 1
		`,
		paymentID,
		terminalID,
		tenantID,
		branchID,
	).Scan(
		&result.PaymentID,
		&result.SaleID,
		&result.SessionID,
		&result.Amount,
		&result.Method,
		&result.Status,
		&confirmedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}

	if err != nil {
		return nil, err
	}

	result.Currency = "KES"
	result.ConfirmedAt = confirmedAt

	return &result, nil
}
