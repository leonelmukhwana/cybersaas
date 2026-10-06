package mpesa

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConfigurationNotFound = errors.New("mpesa configuration not found")
	ErrBranchNotFound        = errors.New("branch not found")
	ErrSTKRequestNotFound    = errors.New("stk request not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type SaveBranchConfigurationParams struct {
	ID                      string
	TenantID                string
	BranchID                string
	Provider                string
	Environment             string
	BusinessShortCode       *string
	TillNumber              *string
	PaybillNumber           *string
	ConsumerKeyEncrypted    *string
	ConsumerSecretEncrypted *string
	PasskeyEncrypted        *string
	AccountReference        *string
	CallbackURL             *string
	Active                  bool
}

type SavePlatformConfigurationParams struct {
	ID                      string
	Provider                string
	Environment             string
	BusinessShortCode       *string
	TillNumber              *string
	PaybillNumber           *string
	ConsumerKeyEncrypted    *string
	ConsumerSecretEncrypted *string
	PasskeyEncrypted        *string
	AccountReference        *string
	CallbackURL             *string
	Active                  bool
}

type CreateSTKRequestParams struct {
	ID                     string
	TenantID               string
	BranchID               string
	SaleID                 string
	PaymentID              *string
	PhoneNumber            string
	Amount                 string
	AccountReference       string
	TransactionDescription string
}

type CreateSubscriptionSTKRequestParams struct {
	ID                     string
	TenantID               string
	SubscriptionID         string
	SubscriptionPaymentID  string
	PhoneNumber            string
	Amount                 string
	AccountReference       string
	TransactionDescription string
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
		return ErrBranchNotFound
	}

	return nil
}

type BranchCredentials struct {
	ConsumerKeyEncrypted    string
	ConsumerSecretEncrypted string
	PasskeyEncrypted        string
}

func (r *Repository) GetBranchCredentials(
	ctx context.Context,
	tenantID string,
	branchID string,
) (*BranchCredentials, error) {
	var credentials BranchCredentials

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			COALESCE(consumer_key_encrypted, ''),
			COALESCE(consumer_secret_encrypted, ''),
			COALESCE(passkey_encrypted, '')
		FROM mpesa_configurations
		WHERE tenant_id = $1
		  AND branch_id = $2
		LIMIT 1
		`,
		tenantID,
		branchID,
	).Scan(
		&credentials.ConsumerKeyEncrypted,
		&credentials.ConsumerSecretEncrypted,
		&credentials.PasskeyEncrypted,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConfigurationNotFound
		}

		return nil, err
	}

	return &credentials, nil
}

func (r *Repository) GetBranchConfiguration(
	ctx context.Context,
	tenantID string,
	branchID string,
) (*Configuration, error) {
	var configuration Configuration

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			provider,
			environment,
			business_short_code,
			till_number,
			paybill_number,
			account_reference,
			callback_url,
			active,
			consumer_key_encrypted IS NOT NULL,
			consumer_secret_encrypted IS NOT NULL,
			passkey_encrypted IS NOT NULL,
			created_at,
			updated_at
		FROM mpesa_configurations
		WHERE tenant_id = $1
		  AND branch_id = $2
		LIMIT 1
		`,
		tenantID,
		branchID,
	).Scan(
		&configuration.ID,
		&configuration.TenantID,
		&configuration.BranchID,
		&configuration.Provider,
		&configuration.Environment,
		&configuration.BusinessShortCode,
		&configuration.TillNumber,
		&configuration.PaybillNumber,
		&configuration.AccountReference,
		&configuration.CallbackURL,
		&configuration.Active,
		&configuration.ConsumerKeyConfigured,
		&configuration.ConsumerSecretConfigured,
		&configuration.PasskeyConfigured,
		&configuration.CreatedAt,
		&configuration.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConfigurationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &configuration, nil
}

func (r *Repository) SaveBranchConfiguration(
	ctx context.Context,
	params SaveBranchConfigurationParams,
) error {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return err
	}

	tenantID, err := uuid.Parse(params.TenantID)
	if err != nil {
		return err
	}

	branchID, err := uuid.Parse(params.BranchID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO mpesa_configurations (
			id,
			tenant_id,
			branch_id,
			provider,
			environment,
			business_short_code,
			till_number,
			paybill_number,
			consumer_key_encrypted,
			consumer_secret_encrypted,
			passkey_encrypted,
			account_reference,
			callback_url,
			active
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14
		)
		ON CONFLICT (branch_id, provider)
		DO UPDATE SET
			environment = EXCLUDED.environment,
			business_short_code = EXCLUDED.business_short_code,
			till_number = EXCLUDED.till_number,
			paybill_number = EXCLUDED.paybill_number,
			consumer_key_encrypted = COALESCE(
				EXCLUDED.consumer_key_encrypted,
				mpesa_configurations.consumer_key_encrypted
			),
			consumer_secret_encrypted = COALESCE(
				EXCLUDED.consumer_secret_encrypted,
				mpesa_configurations.consumer_secret_encrypted
			),
			passkey_encrypted = COALESCE(
				EXCLUDED.passkey_encrypted,
				mpesa_configurations.passkey_encrypted
			),
			account_reference = EXCLUDED.account_reference,
			callback_url = EXCLUDED.callback_url,
			active = EXCLUDED.active,
			updated_at = NOW()
		`,
		id,
		tenantID,
		branchID,
		params.Provider,
		params.Environment,
		params.BusinessShortCode,
		params.TillNumber,
		params.PaybillNumber,
		params.ConsumerKeyEncrypted,
		params.ConsumerSecretEncrypted,
		params.PasskeyEncrypted,
		params.AccountReference,
		params.CallbackURL,
		params.Active,
	)

	return err
}

func (r *Repository) DeleteBranchConfiguration(
	ctx context.Context,
	tenantID string,
	branchID string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		DELETE FROM mpesa_configurations
		WHERE tenant_id = $1
		  AND branch_id = $2
		`,
		tenantID,
		branchID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrConfigurationNotFound
	}

	return nil
}

func (r *Repository) GetPlatformConfiguration(
	ctx context.Context,
) (*Configuration, error) {
	var configuration Configuration

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			provider,
			environment,
			business_short_code,
			till_number,
			paybill_number,
			account_reference,
			callback_url,
			active,
			consumer_key_encrypted IS NOT NULL,
			consumer_secret_encrypted IS NOT NULL,
			passkey_encrypted IS NOT NULL,
			created_at,
			updated_at
		FROM platform_mpesa_configurations
		ORDER BY created_at ASC
		LIMIT 1
		`,
	).Scan(
		&configuration.ID,
		&configuration.Provider,
		&configuration.Environment,
		&configuration.BusinessShortCode,
		&configuration.TillNumber,
		&configuration.PaybillNumber,
		&configuration.AccountReference,
		&configuration.CallbackURL,
		&configuration.Active,
		&configuration.ConsumerKeyConfigured,
		&configuration.ConsumerSecretConfigured,
		&configuration.PasskeyConfigured,
		&configuration.CreatedAt,
		&configuration.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConfigurationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &configuration, nil
}

func (r *Repository) SavePlatformConfiguration(
	ctx context.Context,
	params SavePlatformConfigurationParams,
) error {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO platform_mpesa_configurations (
			id,
			provider,
			environment,
			business_short_code,
			till_number,
			paybill_number,
			consumer_key_encrypted,
			consumer_secret_encrypted,
			passkey_encrypted,
			account_reference,
			callback_url,
			active
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (provider)
		DO UPDATE SET
			environment = EXCLUDED.environment,
			business_short_code = EXCLUDED.business_short_code,
			till_number = EXCLUDED.till_number,
			paybill_number = EXCLUDED.paybill_number,
			consumer_key_encrypted = COALESCE(
				EXCLUDED.consumer_key_encrypted,
				platform_mpesa_configurations.consumer_key_encrypted
			),
			consumer_secret_encrypted = COALESCE(
				EXCLUDED.consumer_secret_encrypted,
				platform_mpesa_configurations.consumer_secret_encrypted
			),
			passkey_encrypted = COALESCE(
				EXCLUDED.passkey_encrypted,
				platform_mpesa_configurations.passkey_encrypted
			),
			account_reference = EXCLUDED.account_reference,
			callback_url = EXCLUDED.callback_url,
			active = EXCLUDED.active,
			updated_at = NOW()
		`,
		id,
		params.Provider,
		params.Environment,
		params.BusinessShortCode,
		params.TillNumber,
		params.PaybillNumber,
		params.ConsumerKeyEncrypted,
		params.ConsumerSecretEncrypted,
		params.PasskeyEncrypted,
		params.AccountReference,
		params.CallbackURL,
		params.Active,
	)

	return err
}

/*
	STK REQUESTS
*/

func (r *Repository) CreateSTKRequest(
	ctx context.Context,
	params CreateSTKRequestParams,
) error {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return err
	}

	tenantID, err := uuid.Parse(params.TenantID)
	if err != nil {
		return err
	}

	branchID, err := uuid.Parse(params.BranchID)
	if err != nil {
		return err
	}

	saleID, err := uuid.Parse(params.SaleID)
	if err != nil {
		return err
	}

	var paymentID *uuid.UUID

	if params.PaymentID != nil && *params.PaymentID != "" {
		parsed, err := uuid.Parse(*params.PaymentID)
		if err != nil {
			return err
		}

		paymentID = &parsed
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO mpesa_stk_requests (
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			phone_number,
			amount,
			account_reference,
			transaction_description,
			status
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
			'pending'
		)
		`,
		id,
		tenantID,
		branchID,
		saleID,
		paymentID,
		params.PhoneNumber,
		params.Amount,
		params.AccountReference,
		params.TransactionDescription,
	)

	return err
}

func (r *Repository) GetSTKRequest(
	ctx context.Context,
	id string,
) (*STKRequest, error) {
	requestID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrSTKRequestNotFound
	}

	var request STKRequest

	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			phone_number,
			amount,
			account_reference,
			transaction_description,
			merchant_request_id,
			checkout_request_id,
			response_code,
			response_description,
			customer_message,
			status,
			result_code,
			result_description,
			mpesa_receipt_number,
			transaction_date,
			callback_received_at,
			created_at,
			updated_at
		FROM mpesa_stk_requests
		WHERE id = $1
		`,
		requestID,
	).Scan(
		&request.ID,
		&request.TenantID,
		&request.BranchID,
		&request.SaleID,
		&request.PaymentID,
		&request.PhoneNumber,
		&request.Amount,
		&request.AccountReference,
		&request.TransactionDescription,
		&request.MerchantRequestID,
		&request.CheckoutRequestID,
		&request.ResponseCode,
		&request.ResponseDescription,
		&request.CustomerMessage,
		&request.Status,
		&request.ResultCode,
		&request.ResultDescription,
		&request.MPesaReceiptNumber,
		&request.TransactionDate,
		&request.CallbackReceivedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSTKRequestNotFound
	}

	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *Repository) GetSTKRequestByCheckoutRequestID(
	ctx context.Context,
	checkoutRequestID string,
) (*STKRequest, error) {
	var request STKRequest

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			sale_id,
			payment_id,
			phone_number,
			amount,
			account_reference,
			transaction_description,
			merchant_request_id,
			checkout_request_id,
			response_code,
			response_description,
			customer_message,
			status,
			result_code,
			result_description,
			mpesa_receipt_number,
			transaction_date,
			callback_received_at,
			created_at,
			updated_at
		FROM mpesa_stk_requests
		WHERE checkout_request_id = $1
		`,
		checkoutRequestID,
	).Scan(
		&request.ID,
		&request.TenantID,
		&request.BranchID,
		&request.SaleID,
		&request.PaymentID,
		&request.PhoneNumber,
		&request.Amount,
		&request.AccountReference,
		&request.TransactionDescription,
		&request.MerchantRequestID,
		&request.CheckoutRequestID,
		&request.ResponseCode,
		&request.ResponseDescription,
		&request.CustomerMessage,
		&request.Status,
		&request.ResultCode,
		&request.ResultDescription,
		&request.MPesaReceiptNumber,
		&request.TransactionDate,
		&request.CallbackReceivedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSTKRequestNotFound
	}

	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *Repository) UpdateSTKAccepted(
	ctx context.Context,
	id string,
	merchantRequestID string,
	checkoutRequestID string,
	responseCode string,
	responseDescription string,
	customerMessage string,
) error {
	requestID, err := uuid.Parse(id)
	if err != nil {
		return ErrSTKRequestNotFound
	}

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE mpesa_stk_requests
		SET
			merchant_request_id = $2,
			checkout_request_id = $3,
			response_code = $4,
			response_description = $5,
			customer_message = $6,
			status = 'accepted',
			updated_at = NOW()
		WHERE id = $1
		`,
		requestID,
		merchantRequestID,
		checkoutRequestID,
		responseCode,
		responseDescription,
		customerMessage,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

func (r *Repository) CompleteSTKRequest(
	ctx context.Context,
	checkoutRequestID string,
	resultCode string,
	resultDescription string,
	mpesaReceiptNumber *string,
	transactionDate *string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE mpesa_stk_requests
		SET
			result_code = $2,
			result_description = $3,
			mpesa_receipt_number = $4,
			transaction_date = CASE
				WHEN $5 IS NULL OR $5 = '' THEN NULL
				ELSE to_timestamp($5, 'YYYYMMDDHH24MISS')
			END,
			status = 'completed',
			callback_received_at = NOW(),
			updated_at = NOW()
		WHERE checkout_request_id = $1
		  AND status IN ('pending', 'accepted')
		`,
		checkoutRequestID,
		resultCode,
		resultDescription,
		mpesaReceiptNumber,
		transactionDate,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

func (r *Repository) FailSTKRequest(
	ctx context.Context,
	checkoutRequestID string,
	resultCode string,
	resultDescription string,
	status string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
		UPDATE mpesa_stk_requests
		SET
			result_code = $2,
			result_description = $3,
			status = $4,
			callback_received_at = NOW(),
			updated_at = NOW()
		WHERE checkout_request_id = $1
		  AND status IN ('pending', 'accepted')
		`,
		checkoutRequestID,
		resultCode,
		resultDescription,
		status,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

func (r *Repository) UpdateSTKCompleted(
	ctx context.Context,
	stkID string,
	resultCode string,
	resultDescription string,
	mpesaReceiptNumber *string,
	transactionDate *time.Time,
) error {
	const query = `
		UPDATE mpesa_stk_requests
		SET
			status = 'completed',
			result_code = $2,
			result_description = $3,
			mpesa_receipt_number = $4,
			transaction_date = $5,
			callback_received_at = NOW()
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		stkID,
		resultCode,
		resultDescription,
		mpesaReceiptNumber,
		transactionDate,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

func (r *Repository) UpdateSTKFailed(
	ctx context.Context,
	stkID string,
	resultCode string,
	resultDescription string,
) error {
	const query = `
		UPDATE mpesa_stk_requests
		SET
			status = 'failed',
			result_code = $2,
			result_description = $3,
			callback_received_at = NOW()
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		stkID,
		resultCode,
		resultDescription,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

// subscription payment brequest
func (r *Repository) CreateSubscriptionSTKRequest(
	ctx context.Context,
	params CreateSubscriptionSTKRequestParams,
) error {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return err
	}

	tenantID, err := uuid.Parse(params.TenantID)
	if err != nil {
		return err
	}

	subscriptionID, err := uuid.Parse(params.SubscriptionID)
	if err != nil {
		return err
	}

	paymentID, err := uuid.Parse(params.SubscriptionPaymentID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO subscription_mpesa_stk_requests (
			id,
			tenant_id,
			subscription_id,
			subscription_payment_id,
			phone_number,
			amount,
			account_reference,
			transaction_description,
			status
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
			'pending'
		)
		`,
		id,
		tenantID,
		subscriptionID,
		paymentID,
		params.PhoneNumber,
		params.Amount,
		params.AccountReference,
		params.TransactionDescription,
	)

	return err
}

// get the request
func (r *Repository) GetSubscriptionSTKRequest(
	ctx context.Context,
	id string,
) (*SubscriptionSTKRequest, error) {
	requestID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrSTKRequestNotFound
	}

	var request SubscriptionSTKRequest

	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			subscription_id,
			subscription_payment_id,
			phone_number,
			amount,
			account_reference,
			transaction_description,
			merchant_request_id,
			checkout_request_id,
			response_code,
			response_description,
			customer_message,
			status,
			result_code,
			result_description,
			mpesa_receipt_number,
			transaction_date,
			callback_received_at,
			created_at,
			updated_at
		FROM subscription_mpesa_stk_requests
		WHERE id = $1
		`,
		requestID,
	).Scan(
		&request.ID,
		&request.TenantID,
		&request.SubscriptionID,
		&request.SubscriptionPaymentID,
		&request.PhoneNumber,
		&request.Amount,
		&request.AccountReference,
		&request.TransactionDescription,
		&request.MerchantRequestID,
		&request.CheckoutRequestID,
		&request.ResponseCode,
		&request.ResponseDescription,
		&request.CustomerMessage,
		&request.Status,
		&request.ResultCode,
		&request.ResultDescription,
		&request.MPesaReceiptNumber,
		&request.TransactionDate,
		&request.CallbackReceivedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSTKRequestNotFound
	}

	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *Repository) GetSubscriptionSTKRequestByCheckoutRequestID(
	ctx context.Context,
	checkoutRequestID string,
) (*SubscriptionSTKRequest, error) {
	var request SubscriptionSTKRequest

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			subscription_id,
			subscription_payment_id,
			phone_number,
			amount,
			account_reference,
			transaction_description,
			merchant_request_id,
			checkout_request_id,
			response_code,
			response_description,
			customer_message,
			status,
			result_code,
			result_description,
			mpesa_receipt_number,
			transaction_date,
			callback_received_at,
			created_at,
			updated_at
		FROM subscription_mpesa_stk_requests
		WHERE checkout_request_id = $1
		`,
		checkoutRequestID,
	).Scan(
		&request.ID,
		&request.TenantID,
		&request.SubscriptionID,
		&request.SubscriptionPaymentID,
		&request.PhoneNumber,
		&request.Amount,
		&request.AccountReference,
		&request.TransactionDescription,
		&request.MerchantRequestID,
		&request.CheckoutRequestID,
		&request.ResponseCode,
		&request.ResponseDescription,
		&request.CustomerMessage,
		&request.Status,
		&request.ResultCode,
		&request.ResultDescription,
		&request.MPesaReceiptNumber,
		&request.TransactionDate,
		&request.CallbackReceivedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSTKRequestNotFound
	}

	if err != nil {
		return nil, err
	}

	return &request, nil
}
func (r *Repository) UpdateSubscriptionSTKAccepted(
	ctx context.Context,
	id string,
	merchantRequestID string,
	checkoutRequestID string,
	responseCode string,
	responseDescription string,
	customerMessage string,
) error {
	requestID, err := uuid.Parse(id)
	if err != nil {
		return ErrSTKRequestNotFound
	}

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE subscription_mpesa_stk_requests
		SET
			merchant_request_id = $2,
			checkout_request_id = $3,
			response_code = $4,
			response_description = $5,
			customer_message = $6,
			status = 'accepted',
			updated_at = NOW()
		WHERE id = $1
		`,
		requestID,
		merchantRequestID,
		checkoutRequestID,
		responseCode,
		responseDescription,
		customerMessage,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

func (r *Repository) UpdateSubscriptionSTKCompleted(
	ctx context.Context,
	stkID string,
	resultCode string,
	resultDescription string,
	mpesaReceiptNumber *string,
	transactionDate *time.Time,
) error {
	requestID, err := uuid.Parse(stkID)
	if err != nil {
		return ErrSTKRequestNotFound
	}

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE subscription_mpesa_stk_requests
		SET
			status = 'completed',
			result_code = $2,
			result_description = $3,
			mpesa_receipt_number = $4,
			transaction_date = $5,
			callback_received_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND status IN ('pending', 'accepted')
		`,
		requestID,
		resultCode,
		resultDescription,
		mpesaReceiptNumber,
		transactionDate,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}

func (r *Repository) UpdateSubscriptionSTKFailed(
	ctx context.Context,
	stkID string,
	resultCode string,
	resultDescription string,
) error {
	requestID, err := uuid.Parse(stkID)
	if err != nil {
		return ErrSTKRequestNotFound
	}

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE subscription_mpesa_stk_requests
		SET
			status = 'failed',
			result_code = $2,
			result_description = $3,
			callback_received_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND status IN ('pending', 'accepted')
		`,
		requestID,
		resultCode,
		resultDescription,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSTKRequestNotFound
	}

	return nil
}
