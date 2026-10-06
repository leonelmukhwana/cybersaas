package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("subscription resource not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// ------------------------------------------------------------
// PLANS
// ------------------------------------------------------------

func (r *Repository) GetPlans(ctx context.Context) ([]Plan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active,
			created_at
		FROM subscription_plans
		WHERE is_active = TRUE
		ORDER BY monthly_price ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]Plan, 0)

	for rows.Next() {
		var p Plan

		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.IncludedBranches,
			&p.IncludedTerminals,
			&p.ExtraBranchRate,
			&p.ExtraTerminalRate,
			&p.MonthlyPrice,
			&p.IsLifetime,
			&p.IsActive,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}

		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *Repository) GetPlan(
	ctx context.Context,
	planID string,
) (Plan, error) {
	var p Plan

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active,
			created_at
		FROM subscription_plans
		WHERE id = $1
	`, planID).Scan(
		&p.ID,
		&p.Name,
		&p.IncludedBranches,
		&p.IncludedTerminals,
		&p.ExtraBranchRate,
		&p.ExtraTerminalRate,
		&p.MonthlyPrice,
		&p.IsLifetime,
		&p.IsActive,
		&p.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, ErrNotFound
	}

	if err != nil {
		return Plan{}, err
	}

	return p, nil
}

// ------------------------------------------------------------
// PLATFORM ADMIN — ALL SUBSCRIPTION PLANS
// ------------------------------------------------------------

func (r *Repository) GetAllPlans(
	ctx context.Context,
) ([]Plan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active,
			created_at
		FROM subscription_plans
		ORDER BY
			is_active DESC,
			monthly_price ASC,
			name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]Plan, 0)

	for rows.Next() {
		var p Plan

		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.IncludedBranches,
			&p.IncludedTerminals,
			&p.ExtraBranchRate,
			&p.ExtraTerminalRate,
			&p.MonthlyPrice,
			&p.IsLifetime,
			&p.IsActive,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}

		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

// ------------------------------------------------------------
// PLATFORM ADMIN — UPDATE SUBSCRIPTION PLAN
// ------------------------------------------------------------

func (r *Repository) UpdatePlan(
	ctx context.Context,
	planID string,
	updated Plan,
	userID string,
) (Plan, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to begin subscription plan update: %w",
			err,
		)
	}
	defer tx.Rollback(ctx)

	var old Plan

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active,
			created_at
		FROM subscription_plans
		WHERE id = $1
		FOR UPDATE
		`,
		planID,
	).Scan(
		&old.ID,
		&old.Name,
		&old.IncludedBranches,
		&old.IncludedTerminals,
		&old.ExtraBranchRate,
		&old.ExtraTerminalRate,
		&old.MonthlyPrice,
		&old.IsLifetime,
		&old.IsActive,
		&old.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, ErrNotFound
	}

	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to load subscription plan: %w",
			err,
		)
	}

	var result Plan

	err = tx.QueryRow(
		ctx,
		`
		UPDATE subscription_plans
		SET
			name = $2,
			included_branches = $3,
			included_terminals = $4,
			extra_branch_rate = $5,
			extra_terminal_rate = $6,
			monthly_price = $7,
			is_lifetime = $8,
			is_active = $9
		WHERE id = $1
		RETURNING
			id,
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active,
			created_at
		`,
		planID,
		updated.Name,
		updated.IncludedBranches,
		updated.IncludedTerminals,
		updated.ExtraBranchRate,
		updated.ExtraTerminalRate,
		updated.MonthlyPrice,
		updated.IsLifetime,
		updated.IsActive,
	).Scan(
		&result.ID,
		&result.Name,
		&result.IncludedBranches,
		&result.IncludedTerminals,
		&result.ExtraBranchRate,
		&result.ExtraTerminalRate,
		&result.MonthlyPrice,
		&result.IsLifetime,
		&result.IsActive,
		&result.CreatedAt,
	)

	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to update subscription plan: %w",
			err,
		)
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO audit_logs (
			id,
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
			'platform_admin',
			'SUBSCRIPTION_PLAN_UPDATED',
			'subscription_plan',
			$2,
			$3::jsonb,
			$4::jsonb,
			'Subscription pricing updated by platform administrator'
		)
		`,
		userID,
		result.ID,
		planJSON(old),
		planJSON(result),
	)

	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to record subscription plan audit: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return Plan{}, fmt.Errorf(
			"failed to commit subscription plan update: %w",
			err,
		)
	}

	return result, nil
}

// ------------------------------------------------------------
// PLATFORM ADMIN — CREATE SUBSCRIPTION PLAN
// ------------------------------------------------------------

func (r *Repository) CreatePlan(
	ctx context.Context,
	plan Plan,
	userID string,
) (Plan, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to begin subscription plan creation: %w",
			err,
		)
	}
	defer tx.Rollback(ctx)

	// Check duplicate package name.
	var existingID string

	err = tx.QueryRow(
		ctx,
		`
		SELECT id
		FROM subscription_plans
		WHERE LOWER(name) = LOWER($1)
		LIMIT 1
		`,
		plan.Name,
	).Scan(&existingID)

	if err == nil {
		return Plan{}, fmt.Errorf(
			"subscription plan with name %q already exists",
			plan.Name,
		)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, fmt.Errorf(
			"failed to check existing subscription plan: %w",
			err,
		)
	}

	var created Plan

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO subscription_plans (
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
		RETURNING
			id,
			name,
			included_branches,
			included_terminals,
			extra_branch_rate,
			extra_terminal_rate,
			monthly_price,
			is_lifetime,
			is_active,
			created_at
		`,
		plan.Name,
		plan.IncludedBranches,
		plan.IncludedTerminals,
		plan.ExtraBranchRate,
		plan.ExtraTerminalRate,
		plan.MonthlyPrice,
		plan.IsLifetime,
		plan.IsActive,
	).Scan(
		&created.ID,
		&created.Name,
		&created.IncludedBranches,
		&created.IncludedTerminals,
		&created.ExtraBranchRate,
		&created.ExtraTerminalRate,
		&created.MonthlyPrice,
		&created.IsLifetime,
		&created.IsActive,
		&created.CreatedAt,
	)

	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to create subscription plan: %w",
			err,
		)
	}

	// Audit creation.
	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO audit_logs (
			id,
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
			'platform_admin',
			'SUBSCRIPTION_PLAN_CREATED',
			'subscription_plan',
			$2,
			'{}'::jsonb,
			$3::jsonb,
			'Subscription package created by platform administrator'
		)
		`,
		userID,
		created.ID,
		planJSON(created),
	)

	if err != nil {
		return Plan{}, fmt.Errorf(
			"failed to record subscription plan creation audit: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return Plan{}, fmt.Errorf(
			"failed to commit subscription plan creation: %w",
			err,
		)
	}

	return created, nil
}

// ------------------------------------------------------------
// SUBSCRIPTION
// ------------------------------------------------------------

func (r *Repository) GetSubscriptionByTenant(
	ctx context.Context,
	tenantID string,
) (Subscription, error) {
	var s Subscription

	err := r.db.QueryRow(ctx, `
		SELECT
			s.id,
			s.tenant_id,
			s.plan_id,
			COALESCE(p.name, ''),
			s.status,
			s.is_trial,
			s.trial_ends_at,
			s.is_lifetime,
			s.account_balance,
			s.current_period_start,
			s.current_period_end,
			s.peak_branches,
			s.peak_terminals,
			s.created_at,
			s.updated_at
		FROM subscriptions s
		LEFT JOIN subscription_plans p
			ON p.id = s.plan_id
		WHERE s.tenant_id = $1
	`, tenantID).Scan(
		&s.ID,
		&s.TenantID,
		&s.PlanID,
		&s.PlanName,
		&s.Status,
		&s.IsTrial,
		&s.TrialEndsAt,
		&s.IsLifetime,
		&s.AccountBalance,
		&s.CurrentPeriodStart,
		&s.CurrentPeriodEnd,
		&s.PeakBranches,
		&s.PeakTerminals,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}

	if err != nil {
		return Subscription{}, err
	}

	return s, nil
}

func (r *Repository) GetSubscription(
	ctx context.Context,
	subscriptionID string,
) (Subscription, error) {
	var s Subscription

	err := r.db.QueryRow(ctx, `
		SELECT
			s.id,
			s.tenant_id,
			s.plan_id,
			COALESCE(p.name, ''),
			s.status,
			s.is_trial,
			s.trial_ends_at,
			s.is_lifetime,
			s.account_balance,
			s.current_period_start,
			s.current_period_end,
			s.peak_branches,
			s.peak_terminals,
			s.created_at,
			s.updated_at
		FROM subscriptions s
		LEFT JOIN subscription_plans p
			ON p.id = s.plan_id
		WHERE s.id = $1
	`, subscriptionID).Scan(
		&s.ID,
		&s.TenantID,
		&s.PlanID,
		&s.PlanName,
		&s.Status,
		&s.IsTrial,
		&s.TrialEndsAt,
		&s.IsLifetime,
		&s.AccountBalance,
		&s.CurrentPeriodStart,
		&s.CurrentPeriodEnd,
		&s.PeakBranches,
		&s.PeakTerminals,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}

	if err != nil {
		return Subscription{}, err
	}

	return s, nil
}

// ------------------------------------------------------------
// CREATE 7-DAY TRIAL SUBSCRIPTION
// ------------------------------------------------------------

func (r *Repository) CreateTrialSubscriptionTx(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	startedAt time.Time,
) error {
	trialEndsAt := startedAt.AddDate(0, 0, 7)

	_, err := tx.Exec(
		ctx,
		`
		INSERT INTO subscriptions (
			tenant_id,
			plan_id,
			status,
			is_trial,
			trial_ends_at,
			is_lifetime,
			account_balance,
			current_period_start,
			current_period_end,
			peak_branches,
			peak_terminals
		)
		VALUES (
			$1,
			NULL,
			'active',
			TRUE,
			$2,
			FALSE,
			0.00,
			$3,
			NULL,
			0,
			0
		)
		`,
		tenantID,
		trialEndsAt,
		startedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create trial subscription: %w",
			err,
		)
	}

	return nil
}

// ------------------------------------------------------------
// SELECT PLAN
// ------------------------------------------------------------

func (r *Repository) SelectPlan(
	ctx context.Context,
	tenantID string,
	planID string,
) (Subscription, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Subscription{}, fmt.Errorf(
			"failed to begin plan selection: %w",
			err,
		)
	}
	defer tx.Rollback(ctx)

	var active bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT is_active
		FROM subscription_plans
		WHERE id = $1
		`,
		planID,
	).Scan(&active)

	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}

	if err != nil {
		return Subscription{}, fmt.Errorf(
			"failed to verify subscription plan: %w",
			err,
		)
	}

	if !active {
		return Subscription{}, errors.New(
			"selected subscription plan is not active",
		)
	}

	var subscriptionID string

	err = tx.QueryRow(
		ctx,
		`
		SELECT id
		FROM subscriptions
		WHERE tenant_id = $1
		  AND status <> 'cancelled'
		FOR UPDATE
		`,
		tenantID,
	).Scan(&subscriptionID)

	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}

	if err != nil {
		return Subscription{}, fmt.Errorf(
			"failed to find tenant subscription: %w",
			err,
		)
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE subscriptions
		SET
			plan_id = $2,
			updated_at = NOW()
		WHERE id = $1
		`,
		subscriptionID,
		planID,
	)

	if err != nil {
		return Subscription{}, fmt.Errorf(
			"failed to select subscription plan: %w",
			err,
		)
	}

	var s Subscription

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			s.id,
			s.tenant_id,
			s.plan_id,
			COALESCE(p.name, ''),
			s.status,
			s.is_trial,
			s.trial_ends_at,
			s.is_lifetime,
			s.account_balance,
			s.current_period_start,
			s.current_period_end,
			s.peak_branches,
			s.peak_terminals,
			s.created_at,
			s.updated_at
		FROM subscriptions s
		LEFT JOIN subscription_plans p
			ON p.id = s.plan_id
		WHERE s.id = $1
		`,
		subscriptionID,
	).Scan(
		&s.ID,
		&s.TenantID,
		&s.PlanID,
		&s.PlanName,
		&s.Status,
		&s.IsTrial,
		&s.TrialEndsAt,
		&s.IsLifetime,
		&s.AccountBalance,
		&s.CurrentPeriodStart,
		&s.CurrentPeriodEnd,
		&s.PeakBranches,
		&s.PeakTerminals,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		return Subscription{}, fmt.Errorf(
			"failed to load selected subscription: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return Subscription{}, fmt.Errorf(
			"failed to commit plan selection: %w",
			err,
		)
	}

	return s, nil
}

// ------------------------------------------------------------
// PAYMENTS
// ------------------------------------------------------------

func (r *Repository) GetSubscriptionPayments(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]SubscriptionPayment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			subscription_id,
			tenant_id,
			amount,
			phone_number,
			payment_method,
			mpesa_checkout_request_id,
			mpesa_receipt_number,
			status,
			failure_reason,
			created_at,
			updated_at
		FROM subscription_payments
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]SubscriptionPayment, 0)

	for rows.Next() {
		var p SubscriptionPayment

		if err := rows.Scan(
			&p.ID,
			&p.SubscriptionID,
			&p.TenantID,
			&p.Amount,
			&p.PhoneNumber,
			&p.PaymentMethod,
			&p.MpesaCheckoutRequestID,
			&p.MpesaReceiptNumber,
			&p.Status,
			&p.FailureReason,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

// ------------------------------------------------------------
// LEDGER
// ------------------------------------------------------------

func (r *Repository) GetLedger(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]LedgerEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			subscription_id,
			tenant_id,
			payment_id,
			entry_type,
			amount,
			balance_after,
			description,
			created_at
		FROM subscription_ledger
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]LedgerEntry, 0)

	for rows.Next() {
		var e LedgerEntry

		if err := rows.Scan(
			&e.ID,
			&e.SubscriptionID,
			&e.TenantID,
			&e.PaymentID,
			&e.EntryType,
			&e.Amount,
			&e.BalanceAfter,
			&e.Description,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}

		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// ------------------------------------------------------------
// USAGE
// ------------------------------------------------------------

func (r *Repository) GetUsage(
	ctx context.Context,
	tenantID string,
) (branches int, terminals int, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT
			(
				SELECT COUNT(*)
				FROM branches
				WHERE tenant_id = $1
				  AND status = 'active'
			),
			(
				SELECT COUNT(*)
				FROM terminals t
				INNER JOIN branches b
					ON b.id = t.branch_id
				WHERE b.tenant_id = $1
				  AND b.status = 'active'
				  AND t.status = 'active'
			)
	`,
		tenantID,
	).Scan(
		&branches,
		&terminals,
	)

	return
}

// ------------------------------------------------------------
// TENANT
// ------------------------------------------------------------

func (r *Repository) TenantExists(
	ctx context.Context,
	tenantID string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM tenants
			WHERE id = $1
		)
	`, tenantID).Scan(&exists)

	return exists, err
}

// ------------------------------------------------------------
// PAYMENT CREATION
// ------------------------------------------------------------

func (r *Repository) CreateSubscriptionPayment(
	ctx context.Context,
	subscriptionID string,
	tenantID string,
	amount string,
	phoneNumber string,
	paymentMethod string,
) (SubscriptionPayment, error) {
	var p SubscriptionPayment

	err := r.db.QueryRow(ctx, `
		INSERT INTO subscription_payments (
			subscription_id,
			tenant_id,
			amount,
			phone_number,
			payment_method,
			status
		)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING
			id,
			subscription_id,
			tenant_id,
			amount,
			phone_number,
			payment_method,
			mpesa_checkout_request_id,
			mpesa_receipt_number,
			status,
			failure_reason,
			created_at,
			updated_at
	`,
		subscriptionID,
		tenantID,
		amount,
		phoneNumber,
		paymentMethod,
	).Scan(
		&p.ID,
		&p.SubscriptionID,
		&p.TenantID,
		&p.Amount,
		&p.PhoneNumber,
		&p.PaymentMethod,
		&p.MpesaCheckoutRequestID,
		&p.MpesaReceiptNumber,
		&p.Status,
		&p.FailureReason,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return SubscriptionPayment{}, err
	}

	return p, nil
}

// ------------------------------------------------------------
// PAYMENT CHECKOUT REQUEST ID
// ------------------------------------------------------------

func (r *Repository) SetSubscriptionPaymentCheckoutRequestID(
	ctx context.Context,
	paymentID string,
	checkoutRequestID string,
) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE subscription_payments
		SET
			mpesa_checkout_request_id = $2,
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'pending'
	`,
		paymentID,
		checkoutRequestID,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to save subscription checkout request ID: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ------------------------------------------------------------
// SUBSCRIPTION M-PESA STK
// ------------------------------------------------------------

type SubscriptionSTKRequest struct {
	ID                     string
	TenantID               string
	SubscriptionID         string
	SubscriptionPaymentID  string
	PhoneNumber            string
	Amount                 string
	AccountReference       string
	TransactionDescription string
	MerchantRequestID      *string
	CheckoutRequestID      *string
	ResponseCode           *string
	ResponseDescription    *string
	CustomerMessage        *string
	Status                 string
	ResultCode             *string
	ResultDescription      *string
	MpesaReceiptNumber     *string
	TransactionDate        *time.Time
	CallbackReceivedAt     *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
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

func (r *Repository) CreateSubscriptionSTKRequest(
	ctx context.Context,
	params CreateSubscriptionSTKRequestParams,
) (SubscriptionSTKRequest, error) {
	var request SubscriptionSTKRequest

	err := r.db.QueryRow(ctx, `
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
		RETURNING
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
	`,
		params.ID,
		params.TenantID,
		params.SubscriptionID,
		params.SubscriptionPaymentID,
		params.PhoneNumber,
		params.Amount,
		params.AccountReference,
		params.TransactionDescription,
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
		&request.MpesaReceiptNumber,
		&request.TransactionDate,
		&request.CallbackReceivedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		return SubscriptionSTKRequest{}, err
	}

	return request, nil
}

// ------------------------------------------------------------
// SUBSCRIPTION STK ACCEPTED
// ------------------------------------------------------------

func (r *Repository) UpdateSubscriptionSTKAccepted(
	ctx context.Context,
	requestID string,
	merchantRequestID string,
	checkoutRequestID string,
	responseCode string,
	responseDescription string,
	customerMessage string,
) error {
	commandTag, err := r.db.Exec(ctx, `
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
		  AND status IN ('pending', 'accepted')
	`,
		requestID,
		merchantRequestID,
		checkoutRequestID,
		responseCode,
		responseDescription,
		customerMessage,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update subscription STK accepted state: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ------------------------------------------------------------
// FIND SUBSCRIPTION STK BY CHECKOUT REQUEST ID
// ------------------------------------------------------------

func (r *Repository) GetSubscriptionSTKByCheckoutRequestID(
	ctx context.Context,
	checkoutRequestID string,
) (SubscriptionSTKRequest, error) {
	var request SubscriptionSTKRequest

	err := r.db.QueryRow(ctx, `
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
	`, checkoutRequestID).Scan(
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
		&request.MpesaReceiptNumber,
		&request.TransactionDate,
		&request.CallbackReceivedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return SubscriptionSTKRequest{}, ErrNotFound
	}

	if err != nil {
		return SubscriptionSTKRequest{}, err
	}

	return request, nil
}

// ------------------------------------------------------------
// SUBSCRIPTION STK COMPLETED
// ------------------------------------------------------------

func (r *Repository) UpdateSubscriptionSTKCompleted(
	ctx context.Context,
	checkoutRequestID string,
	resultCode string,
	resultDescription string,
	mpesaReceiptNumber string,
	transactionDate *time.Time,
) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE subscription_mpesa_stk_requests
		SET
			status = 'completed',
			result_code = $2,
			result_description = $3,
			mpesa_receipt_number = $4,
			transaction_date = $5,
			callback_received_at = NOW(),
			updated_at = NOW()
		WHERE checkout_request_id = $1
		  AND status <> 'completed'
	`,
		checkoutRequestID,
		resultCode,
		resultDescription,
		mpesaReceiptNumber,
		transactionDate,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to mark subscription STK completed: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		var exists bool

		err := r.db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM subscription_mpesa_stk_requests
				WHERE checkout_request_id = $1
				  AND status = 'completed'
			)
		`, checkoutRequestID).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		return ErrNotFound
	}

	return nil
}

// ------------------------------------------------------------
// SUBSCRIPTION STK FAILED
// ------------------------------------------------------------

func (r *Repository) UpdateSubscriptionSTKFailed(
	ctx context.Context,
	checkoutRequestID string,
	resultCode string,
	resultDescription string,
) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE subscription_mpesa_stk_requests
		SET
			status = 'failed',
			result_code = $2,
			result_description = $3,
			callback_received_at = NOW(),
			updated_at = NOW()
		WHERE checkout_request_id = $1
		  AND status NOT IN ('completed', 'failed')
	`,
		checkoutRequestID,
		resultCode,
		resultDescription,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to mark subscription STK failed: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		var exists bool

		err := r.db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM subscription_mpesa_stk_requests
				WHERE checkout_request_id = $1
			)
		`, checkoutRequestID).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		return ErrNotFound
	}

	return nil
}

// ------------------------------------------------------------
// PAYMENT COMPLETION
// ------------------------------------------------------------

func (r *Repository) MarkPaymentCompleted(
	ctx context.Context,
	paymentID string,
	receiptNumber string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to begin payment completion transaction: %w",
			err,
		)
	}
	defer tx.Rollback(ctx)

	var (
		subscriptionID string
		tenantID       string
		amount         string
		paymentStatus  string
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			subscription_id,
			tenant_id,
			amount,
			status
		FROM subscription_payments
		WHERE id = $1
		FOR UPDATE
		`,
		paymentID,
	).Scan(
		&subscriptionID,
		&tenantID,
		&amount,
		&paymentStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf(
			"failed to query payment: %w",
			err,
		)
	}

	if paymentStatus == "completed" {
		return nil
	}

	if paymentStatus != "pending" {
		return fmt.Errorf(
			"payment cannot be completed from status %s",
			paymentStatus,
		)
	}

	now := time.Now().UTC()

	var (
		subStatus            string
		subscriptionTenantID string
		isTrial              bool
		planID               *string
		currentPeriodStart   time.Time
		currentPeriodEnd     *time.Time
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			tenant_id,
			status,
			is_trial,
			plan_id,
			current_period_start,
			current_period_end
		FROM subscriptions
		WHERE id = $1
		FOR UPDATE
		`,
		subscriptionID,
	).Scan(
		&subscriptionTenantID,
		&subStatus,
		&isTrial,
		&planID,
		&currentPeriodStart,
		&currentPeriodEnd,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf(
			"failed to query subscription: %w",
			err,
		)
	}

	if subscriptionTenantID != tenantID {
		return errors.New(
			"payment tenant does not match subscription tenant",
		)
	}

	if planID == nil || *planID == "" {
		return errors.New(
			"cannot complete subscription payment without a selected plan",
		)
	}

	var (
		planIsLifetime bool
		planIsActive   bool
	)

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			is_lifetime,
			is_active
		FROM subscription_plans
		WHERE id = $1
		`,
		*planID,
	).Scan(
		&planIsLifetime,
		&planIsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New(
			"selected subscription plan was not found",
		)
	}

	if err != nil {
		return fmt.Errorf(
			"failed to query selected subscription plan: %w",
			err,
		)
	}

	if !planIsActive {
		return errors.New(
			"selected subscription plan is no longer active",
		)
	}

	var (
		newPeriodStart time.Time
		newPeriodEnd   *time.Time
	)

	if planIsLifetime {
		newPeriodStart = now
		newPeriodEnd = nil
	} else if !isTrial &&
		subStatus == "active" &&
		currentPeriodEnd != nil &&
		currentPeriodEnd.After(now) {

		newPeriodStart = currentPeriodStart

		end := currentPeriodEnd.AddDate(0, 1, 0)
		newPeriodEnd = &end
	} else {
		newPeriodStart = now

		end := now.AddDate(0, 1, 0)
		newPeriodEnd = &end
	}

	result, err := tx.Exec(
		ctx,
		`
		UPDATE subscription_payments
		SET
			status = 'completed',
			mpesa_receipt_number = $2,
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'pending'
		`,
		paymentID,
		receiptNumber,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to mark payment completed: %w",
			err,
		)
	}

	if result.RowsAffected() != 1 {
		return errors.New(
			"subscription payment was not updated",
		)
	}

	var newAccountBalance string

	err = tx.QueryRow(
		ctx,
		`
		UPDATE subscriptions
		SET
			status = 'active',
			is_trial = FALSE,
			trial_ends_at = NULL,
			is_lifetime = $2,
			account_balance = account_balance + $3::numeric,
			current_period_start = $4,
			current_period_end = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING account_balance
		`,
		subscriptionID,
		planIsLifetime,
		amount,
		newPeriodStart,
		newPeriodEnd,
	).Scan(&newAccountBalance)

	if err != nil {
		return fmt.Errorf(
			"failed to activate paid subscription: %w",
			err,
		)
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO subscription_ledger (
			subscription_id,
			tenant_id,
			payment_id,
			entry_type,
			amount,
			balance_after,
			description
		)
		VALUES (
			$1,
			$2,
			$3,
			'CREDIT',
			$4,
			$5,
			'SaaS subscription payment'
		)
		`,
		subscriptionID,
		tenantID,
		paymentID,
		amount,
		newAccountBalance,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to record subscription ledger entry: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"failed to commit payment completion: %w",
			err,
		)
	}

	return nil
}

// ------------------------------------------------------------
// PAYMENT FAILED
// ------------------------------------------------------------

func (r *Repository) MarkPaymentFailed(
	ctx context.Context,
	paymentID string,
	failureReason string,
) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE subscription_payments
		SET
			status = 'failed',
			failure_reason = $2,
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'pending'
	`,
		paymentID,
		failureReason,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to mark subscription payment failed: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		var status string

		err := r.db.QueryRow(ctx, `
			SELECT status
			FROM subscription_payments
			WHERE id = $1
		`, paymentID).Scan(&status)

		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}

		if err != nil {
			return fmt.Errorf(
				"failed to check subscription payment status: %w",
				err,
			)
		}

		if status == "failed" {
			return nil
		}

		return fmt.Errorf(
			"subscription payment cannot be marked failed from status %s",
			status,
		)
	}

	return nil
}

// ------------------------------------------------------------
// HELPERS
// ------------------------------------------------------------

func planJSON(plan Plan) string {
	data, err := json.Marshal(plan)
	if err != nil {
		return "{}"
	}

	return string(data)
}
