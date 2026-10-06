package report

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendantRepository struct {
	db *pgxpool.Pool
}

func NewAttendantRepository(db *pgxpool.Pool) *AttendantRepository {
	return &AttendantRepository{
		db: db,
	}
}

func (r *AttendantRepository) GetSummary(
	ctx context.Context,
	tenantID string,
	attendantID string,
	periodStart string,
	periodEnd string,
) (AttendantReportSummary, error) {
	var result AttendantReportSummary

	var sessionRevenue string
	var salesRevenue string
	var confirmedCollections string

	var sessionCount int64
	var salesCount int64
	var paymentsCount int64

	// ------------------------------------------------------------
	// SESSION REVENUE
	//
	// Only sessions started/handled by this attendant count.
	// Same definition as Owner Report.
	// ------------------------------------------------------------
	sessionQuery := `
		SELECT
			COALESCE(SUM(final_amount), 0)::text,
			COUNT(*)
		FROM sessions
		WHERE tenant_id = $1
		  AND attendant_id = $2
		  AND started_at >= $3
		  AND started_at < $4
		  AND status = 'completed'
		  AND final_amount IS NOT NULL
	`

	if err := r.db.QueryRow(
		ctx,
		sessionQuery,
		tenantID,
		attendantID,
		periodStart,
		periodEnd,
	).Scan(
		&sessionRevenue,
		&sessionCount,
	); err != nil {
		return result, fmt.Errorf(
			"get attendant session revenue: %w",
			err,
		)
	}

	// ------------------------------------------------------------
	// SALES REVENUE
	//
	// Sales are attributed directly through sales.attendant_id.
	// Same definition as Owner Report.
	// ------------------------------------------------------------
	salesQuery := `
		SELECT
			COALESCE(SUM(total_amount), 0)::text,
			COUNT(*)
		FROM sales
		WHERE tenant_id = $1
		  AND attendant_id = $2
		  AND created_at >= $3
		  AND created_at < $4
		  AND status = 'completed'
	`

	if err := r.db.QueryRow(
		ctx,
		salesQuery,
		tenantID,
		attendantID,
		periodStart,
		periodEnd,
	).Scan(
		&salesRevenue,
		&salesCount,
	); err != nil {
		return result, fmt.Errorf(
			"get attendant sales revenue: %w",
			err,
		)
	}

	// ------------------------------------------------------------
	// CONFIRMED COLLECTIONS
	//
	// payments do not contain attendant_id.
	//
	// Therefore:
	//
	// payments -> sales -> attendant_id
	//
	// This prevents one attendant from being credited for
	// another attendant's payments.
	// ------------------------------------------------------------
	paymentQuery := `
		SELECT
			COALESCE(SUM(p.amount), 0)::text,
			COUNT(*)
		FROM payments p
		INNER JOIN sales s
			ON s.id = p.sale_id
		   AND s.tenant_id = p.tenant_id
		   AND s.branch_id = p.branch_id
		WHERE p.tenant_id = $1
		  AND s.attendant_id = $2
		  AND p.created_at >= $3
		  AND p.created_at < $4
		  AND p.status = 'confirmed'
	`

	if err := r.db.QueryRow(
		ctx,
		paymentQuery,
		tenantID,
		attendantID,
		periodStart,
		periodEnd,
	).Scan(
		&confirmedCollections,
		&paymentsCount,
	); err != nil {
		return result, fmt.Errorf(
			"get attendant collections: %w",
			err,
		)
	}

	totalRevenue := addMoney(
		sessionRevenue,
		salesRevenue,
	)

	result = AttendantReportSummary{
		AttendantID: attendantID,

		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,

		SessionRevenue: sessionRevenue,
		SalesRevenue:   salesRevenue,
		TotalRevenue:   totalRevenue,

		ConfirmedCollections: confirmedCollections,

		SessionCount:  sessionCount,
		SalesCount:    salesCount,
		PaymentsCount: paymentsCount,
	}

	return result, nil
}

func (r *AttendantRepository) GetPaymentMethods(
	ctx context.Context,
	tenantID string,
	attendantID string,
	periodStart string,
	periodEnd string,
) ([]AttendantPaymentMethodTotal, error) {
	query := `
		SELECT
			p.method::text,
			COALESCE(SUM(p.amount), 0)::text,
			COUNT(*)
		FROM payments p
		INNER JOIN sales s
			ON s.id = p.sale_id
		   AND s.tenant_id = p.tenant_id
		   AND s.branch_id = p.branch_id
		WHERE p.tenant_id = $1
		  AND s.attendant_id = $2
		  AND p.created_at >= $3
		  AND p.created_at < $4
		  AND p.status = 'confirmed'
		GROUP BY p.method
		ORDER BY p.method
	`

	rows, err := r.db.Query(
		ctx,
		query,
		tenantID,
		attendantID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get attendant payment methods: %w",
			err,
		)
	}
	defer rows.Close()

	results := make(
		[]AttendantPaymentMethodTotal,
		0,
	)

	for rows.Next() {
		var item AttendantPaymentMethodTotal

		if err := rows.Scan(
			&item.Method,
			&item.Amount,
			&item.Count,
		); err != nil {
			return nil, fmt.Errorf(
				"scan attendant payment method: %w",
				err,
			)
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate attendant payment methods: %w",
			err,
		)
	}

	return results, nil
}

func (r *AttendantRepository) GetDailyTotals(
	ctx context.Context,
	tenantID string,
	attendantID string,
	periodStart string,
	periodEnd string,
) ([]AttendantDailyTotal, error) {
	query := `
		WITH days AS (
			SELECT generate_series(
				$3::timestamptz,
				$4::timestamptz - INTERVAL '1 day',
				INTERVAL '1 day'
			)::date AS report_date
		),

		session_totals AS (
			SELECT
				started_at::date AS report_date,
				COALESCE(SUM(final_amount), 0)::text
					AS session_revenue,
				COUNT(*) AS session_count
			FROM sessions
			WHERE tenant_id = $1
			  AND attendant_id = $2
			  AND started_at >= $3
			  AND started_at < $4
			  AND status = 'completed'
			  AND final_amount IS NOT NULL
			GROUP BY started_at::date
		),

		sale_totals AS (
			SELECT
				created_at::date AS report_date,
				COALESCE(SUM(total_amount), 0)::text
					AS sales_revenue,
				COUNT(*) AS sales_count
			FROM sales
			WHERE tenant_id = $1
			  AND attendant_id = $2
			  AND created_at >= $3
			  AND created_at < $4
			  AND status = 'completed'
			GROUP BY created_at::date
		)

		SELECT
			d.report_date::text,

			COALESCE(st.session_revenue, '0.00'),
			COALESCE(sa.sales_revenue, '0.00'),

			(
				COALESCE(st.session_revenue, '0.00')::numeric
				+
				COALESCE(sa.sales_revenue, '0.00')::numeric
			)::text,

			COALESCE(st.session_count, 0),
			COALESCE(sa.sales_count, 0)

		FROM days d

		LEFT JOIN session_totals st
			ON st.report_date = d.report_date

		LEFT JOIN sale_totals sa
			ON sa.report_date = d.report_date

		ORDER BY d.report_date DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		tenantID,
		attendantID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get attendant daily totals: %w",
			err,
		)
	}
	defer rows.Close()

	results := make(
		[]AttendantDailyTotal,
		0,
	)

	for rows.Next() {
		var item AttendantDailyTotal

		if err := rows.Scan(
			&item.Date,
			&item.SessionRevenue,
			&item.SalesRevenue,
			&item.TotalRevenue,
			&item.SessionCount,
			&item.SalesCount,
		); err != nil {
			return nil, fmt.Errorf(
				"scan attendant daily total: %w",
				err,
			)
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate attendant daily totals: %w",
			err,
		)
	}

	return results, nil
}
