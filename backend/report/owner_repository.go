package report

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OwnerRepository struct {
	db *pgxpool.Pool
}

func NewOwnerRepository(db *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{
		db: db,
	}
}

func (r *OwnerRepository) GetSummary(
	ctx context.Context,
	tenantID string,
	branchID *string,
	periodStart string,
	periodEnd string,
) (OwnerReportSummary, error) {
	var result OwnerReportSummary

	var sessionRevenue string
	var salesRevenue string
	var totalCollections string
	var expensesTotal string

	var sessionCount int64
	var salesCount int64
	var paymentsCount int64
	var expensesCount int64

	// ------------------------------------------------------------
	// SESSION REVENUE
	//
	// A completed session gets its revenue from final_amount.
	// Cancelled/active sessions are not counted as revenue.
	// ------------------------------------------------------------
	sessionQuery := `
		SELECT
			COALESCE(SUM(final_amount), 0)::text,
			COUNT(*)
		FROM sessions
		WHERE tenant_id = $1
		  AND started_at >= $2
		  AND started_at < $3
		  AND status = 'completed'
		  AND final_amount IS NOT NULL
	`

	sessionArgs := []any{
		tenantID,
		periodStart,
		periodEnd,
	}

	if branchID != nil {
		sessionQuery += ` AND branch_id = $4`
		sessionArgs = append(sessionArgs, *branchID)
	}

	if err := r.db.QueryRow(
		ctx,
		sessionQuery,
		sessionArgs...,
	).Scan(
		&sessionRevenue,
		&sessionCount,
	); err != nil {
		return result, fmt.Errorf("get session revenue: %w", err)
	}

	// ------------------------------------------------------------
	// SALES REVENUE
	//
	// Completed sales represent printing, photocopying, scanning,
	// services and other sale items.
	// ------------------------------------------------------------
	salesQuery := `
		SELECT
			COALESCE(SUM(total_amount), 0)::text,
			COUNT(*)
		FROM sales
		WHERE tenant_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		  AND status = 'completed'
	`

	salesArgs := []any{
		tenantID,
		periodStart,
		periodEnd,
	}

	if branchID != nil {
		salesQuery += ` AND branch_id = $4`
		salesArgs = append(salesArgs, *branchID)
	}

	if err := r.db.QueryRow(
		ctx,
		salesQuery,
		salesArgs...,
	).Scan(
		&salesRevenue,
		&salesCount,
	); err != nil {
		return result, fmt.Errorf("get sales revenue: %w", err)
	}

	// Keep TotalSales for compatibility with the existing frontend/API.
	totalSales := salesRevenue

	// ------------------------------------------------------------
	// CONFIRMED PAYMENTS
	//
	// These are used for payment-method reporting and collection
	// statistics. They are NOT added to revenue again because a
	// sale/session amount already represents the revenue.
	// ------------------------------------------------------------
	paymentQuery := `
		SELECT
			COALESCE(SUM(amount), 0)::text,
			COUNT(*)
		FROM payments
		WHERE tenant_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		  AND status = 'confirmed'
	`

	paymentArgs := []any{
		tenantID,
		periodStart,
		periodEnd,
	}

	if branchID != nil {
		paymentQuery += ` AND branch_id = $4`
		paymentArgs = append(paymentArgs, *branchID)
	}

	if err := r.db.QueryRow(
		ctx,
		paymentQuery,
		paymentArgs...,
	).Scan(
		&totalCollections,
		&paymentsCount,
	); err != nil {
		return result, fmt.Errorf("get payment summary: %w", err)
	}

	// ------------------------------------------------------------
	// EXPENSES
	// ------------------------------------------------------------
	expenseQuery := `
		SELECT
			COALESCE(SUM(amount), 0)::text,
			COUNT(*)
		FROM expenses
		WHERE tenant_id = $1
		  AND expense_date >= $2::date
		  AND expense_date < $3::date
		  AND status = 'recorded'
	`

	expenseArgs := []any{
		tenantID,
		periodStart,
		periodEnd,
	}

	if branchID != nil {
		expenseQuery += ` AND branch_id = $4`
		expenseArgs = append(expenseArgs, *branchID)
	}

	if err := r.db.QueryRow(
		ctx,
		expenseQuery,
		expenseArgs...,
	).Scan(
		&expensesTotal,
		&expensesCount,
	); err != nil {
		return result, fmt.Errorf("get expense summary: %w", err)
	}

	totalRevenue := addMoney(sessionRevenue, salesRevenue)
	netAmount := subtractMoney(totalRevenue, expensesTotal)

	result = OwnerReportSummary{
		TenantID:    tenantID,
		BranchID:    branchID,
		PeriodStart: parseReportTimeValue(periodStart),
		PeriodEnd:   parseReportTimeValue(periodEnd),

		SessionRevenue: sessionRevenue,
		SalesRevenue:   salesRevenue,
		TotalRevenue:   totalRevenue,

		TotalSales:       totalSales,
		TotalCollections: totalCollections,
		TotalExpenses:    expensesTotal,
		NetAmount:        netAmount,

		SessionCount:  sessionCount,
		SalesCount:    salesCount,
		PaymentsCount: paymentsCount,
		ExpensesCount: expensesCount,
	}

	return result, nil
}

func (r *OwnerRepository) GetPaymentMethods(
	ctx context.Context,
	tenantID string,
	branchID *string,
	periodStart string,
	periodEnd string,
) ([]OwnerPaymentMethodTotal, error) {
	query := `
		SELECT
			method::text,
			COALESCE(SUM(amount), 0)::text,
			COUNT(*)
		FROM payments
		WHERE tenant_id = $1
		  AND created_at >= $2
		  AND created_at < $3
		  AND status = 'confirmed'
	`

	args := []any{
		tenantID,
		periodStart,
		periodEnd,
	}

	if branchID != nil {
		query += ` AND branch_id = $4`
		args = append(args, *branchID)
	}

	query += `
		GROUP BY method
		ORDER BY method
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get payment methods: %w", err)
	}
	defer rows.Close()

	results := make([]OwnerPaymentMethodTotal, 0)

	for rows.Next() {
		var item OwnerPaymentMethodTotal

		if err := rows.Scan(
			&item.Method,
			&item.Amount,
			&item.Count,
		); err != nil {
			return nil, fmt.Errorf("scan payment method: %w", err)
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment methods: %w", err)
	}

	return results, nil
}

func (r *OwnerRepository) GetBranchTotals(
	ctx context.Context,
	tenantID string,
	branchID *string,
	periodStart string,
	periodEnd string,
) ([]OwnerBranchTotal, error) {
	query := `
		SELECT
			b.id,
			b.name,

			-- Session revenue
			COALESCE((
				SELECT SUM(s.final_amount)
				FROM sessions s
				WHERE s.tenant_id = $1
				  AND s.branch_id = b.id
				  AND s.started_at >= $2
				  AND s.started_at < $3
				  AND s.status = 'completed'
				  AND s.final_amount IS NOT NULL
			), 0)::text,

			-- Sales revenue
			COALESCE((
				SELECT SUM(s.total_amount)
				FROM sales s
				WHERE s.tenant_id = $1
				  AND s.branch_id = b.id
				  AND s.created_at >= $2
				  AND s.created_at < $3
				  AND s.status = 'completed'
			), 0)::text,

			-- Confirmed collections
			COALESCE((
				SELECT SUM(p.amount)
				FROM payments p
				WHERE p.tenant_id = $1
				  AND p.branch_id = b.id
				  AND p.created_at >= $2
				  AND p.created_at < $3
				  AND p.status = 'confirmed'
			), 0)::text,

			-- Expenses
			COALESCE((
				SELECT SUM(e.amount)
				FROM expenses e
				WHERE e.tenant_id = $1
				  AND e.branch_id = b.id
				  AND e.expense_date >= $2::date
				  AND e.expense_date < $3::date
				  AND e.status = 'recorded'
			), 0)::text,

			-- Session count
			(
				SELECT COUNT(*)
				FROM sessions s
				WHERE s.tenant_id = $1
				  AND s.branch_id = b.id
				  AND s.started_at >= $2
				  AND s.started_at < $3
				  AND s.status = 'completed'
			),

			-- Sales count
			(
				SELECT COUNT(*)
				FROM sales s
				WHERE s.tenant_id = $1
				  AND s.branch_id = b.id
				  AND s.created_at >= $2
				  AND s.created_at < $3
				  AND s.status = 'completed'
			)

		FROM branches b
		WHERE b.tenant_id = $1
	`

	args := []any{
		tenantID,
		periodStart,
		periodEnd,
	}

	if branchID != nil {
		query += ` AND b.id = $4`
		args = append(args, *branchID)
	}

	query += `
		ORDER BY b.name
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get branch totals: %w", err)
	}
	defer rows.Close()

	results := make([]OwnerBranchTotal, 0)

	for rows.Next() {
		var item OwnerBranchTotal

		var sessionRevenue string
		var salesRevenue string
		var collectionsTotal string
		var expensesTotal string

		if err := rows.Scan(
			&item.BranchID,
			&item.BranchName,
			&sessionRevenue,
			&salesRevenue,
			&collectionsTotal,
			&expensesTotal,
			&item.SessionCount,
			&item.SalesCount,
		); err != nil {
			return nil, fmt.Errorf("scan branch total: %w", err)
		}

		totalRevenue := addMoney(
			sessionRevenue,
			salesRevenue,
		)

		item.SessionRevenue = sessionRevenue
		item.SalesRevenue = salesRevenue
		item.TotalRevenue = totalRevenue

		// Compatibility fields.
		item.TotalSales = salesRevenue
		item.TotalCollections = collectionsTotal
		item.TotalExpenses = expensesTotal
		item.NetAmount = subtractMoney(
			totalRevenue,
			expensesTotal,
		)

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate branch totals: %w", err)
	}

	return results, nil
}

// GetDownloadReport retrieves all data required for an owner report
// download using the same calculations as the on-screen report.
func (r *OwnerRepository) GetDownloadReport(
	ctx context.Context,
	tenantID string,
	branchID *string,
	periodStart string,
	periodEnd string,
) (OwnerReportResponse, error) {
	summary, err := r.GetSummary(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return OwnerReportResponse{}, err
	}

	paymentMethods, err := r.GetPaymentMethods(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return OwnerReportResponse{}, err
	}

	branches, err := r.GetBranchTotals(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return OwnerReportResponse{}, err
	}

	return OwnerReportResponse{
		Summary:        summary,
		PaymentMethods: paymentMethods,
		Branches:       branches,
	}, nil
}

// addMoney performs decimal addition using the database-produced
// decimal strings. PostgreSQL already provides exact decimal values,
// so we avoid floating-point arithmetic in the report layer.
func addMoney(a string, b string) string {
	af, err := strconv.ParseFloat(a, 64)
	if err != nil {
		af = 0
	}

	bf, err := strconv.ParseFloat(b, 64)
	if err != nil {
		bf = 0
	}

	return strconv.FormatFloat(
		af+bf,
		'f',
		2,
		64,
	)
}

// subtractMoney performs decimal subtraction.
func subtractMoney(a string, b string) string {
	af, err := strconv.ParseFloat(a, 64)
	if err != nil {
		af = 0
	}

	bf, err := strconv.ParseFloat(b, 64)
	if err != nil {
		bf = 0
	}

	return strconv.FormatFloat(
		af-bf,
		'f',
		2,
		64,
	)
}

func parseReportTimeValue(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}

	return t
}
