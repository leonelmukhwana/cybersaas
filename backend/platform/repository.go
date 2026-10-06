package platform

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// ============================================================
// DASHBOARD
// ============================================================

func (r *Repository) GetDashboardStats(
	ctx context.Context,
) (DashboardStats, error) {
	const query = `
		SELECT
			(
				SELECT COUNT(*)
				FROM users
				WHERE role = 'owner'
			),

			(
				SELECT COUNT(*)
				FROM users
				WHERE role = 'owner'
				  AND status = 'active'
			),

			(
				SELECT COUNT(*)
				FROM users
				WHERE role = 'owner'
				  AND status = 'suspended'
			),

			(
				SELECT COUNT(*)
				FROM branches
			),

			(
				SELECT COUNT(*)
				FROM branches
				WHERE status = 'active'
			),

			(
				SELECT COUNT(*)
				FROM subscriptions
				WHERE status = 'active'
				  AND (
						is_lifetime = TRUE
						OR current_period_end IS NULL
						OR current_period_end > NOW()
				  )
			),

			(
				SELECT COUNT(*)
				FROM subscriptions
				WHERE status = 'past_due'
			),

			(
				SELECT COUNT(*)
				FROM subscriptions
				WHERE status = 'expired'
				   OR (
						is_lifetime = FALSE
						AND current_period_end IS NOT NULL
						AND current_period_end <= NOW()
				   )
			),

			(
				SELECT COUNT(*)
				FROM subscriptions
				WHERE is_lifetime = TRUE
			),

			(
				SELECT COALESCE(SUM(amount), 0)
				FROM subscription_payments
				WHERE status = 'completed'
			)::numeric(18,2)::text,

			(
				SELECT COALESCE(SUM(amount), 0)
				FROM subscription_payments
				WHERE status = 'completed'
				  AND created_at >= date_trunc('month', NOW())
			)::numeric(18,2)::text,

			(
				SELECT COALESCE(SUM(amount), 0)
				FROM subscription_payments
				WHERE status = 'completed'
				  AND created_at >= date_trunc('year', NOW())
			)::numeric(18,2)::text,

			(
				SELECT COALESCE(SUM(amount), 0)
				FROM subscription_payments
				WHERE status = 'completed'
				  AND created_at >= NOW() - INTERVAL '3 years'
			)::numeric(18,2)::text
	`

	var stats DashboardStats

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalCyberOwners,
		&stats.ActiveCyberOwners,
		&stats.SuspendedCyberOwners,
		&stats.TotalCyberLocations,
		&stats.ActiveCyberLocations,
		&stats.ActiveSubscriptions,
		&stats.PastDueSubscriptions,
		&stats.ExpiredSubscriptions,
		&stats.LifetimeSubscriptions,
		&stats.SubscriptionRevenueAllTime,
		&stats.SubscriptionRevenueMonth,
		&stats.SubscriptionRevenueYear,
		&stats.SubscriptionRevenueThreeYears,
	)

	if err != nil {
		return DashboardStats{}, fmt.Errorf(
			"get platform dashboard stats: %w",
			err,
		)
	}

	return stats, nil
}

// ============================================================
// CYBER OWNERS
// ============================================================

func (r *Repository) GetCyberOwners(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (CyberOwnerListResponse, error) {

	const countQuery = `
		SELECT COUNT(*)
		FROM users u
		WHERE u.role = 'owner'
		  AND (
				$1 = ''
				OR u.full_name ILIKE '%' || $1 || '%'
				OR COALESCE(u.email, '') ILIKE '%' || $1 || '%'
				OR COALESCE(u.phone, '') ILIKE '%' || $1 || '%'
				OR EXISTS (
					SELECT 1
					FROM subscriptions s
					LEFT JOIN subscription_plans sp
						ON sp.id = s.plan_id
					WHERE s.tenant_id = u.tenant_id
					  AND COALESCE(sp.name, '') ILIKE '%' || $1 || '%'
				)
		  )
	`

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		search,
	).Scan(&total); err != nil {
		return CyberOwnerListResponse{}, fmt.Errorf(
			"count cyber owners: %w",
			err,
		)
	}

	const query = `
		SELECT
			u.id,
			u.full_name,
			u.email,
			u.phone,
			u.status,

			s.id,
			sp.name,

			COALESCE(
				(
					SELECT p.amount::numeric(12,2)::text
					FROM subscription_payments p
					WHERE p.subscription_id = s.id
					  AND p.status = 'completed'
					ORDER BY p.created_at DESC
					LIMIT 1
				),
				COALESCE(sp.monthly_price, 0)::numeric(12,2)::text
			),

			s.status,
			s.current_period_end,
			COALESCE(s.is_lifetime, FALSE)

		FROM users u

		LEFT JOIN subscriptions s
			ON s.tenant_id = u.tenant_id

		LEFT JOIN subscription_plans sp
			ON sp.id = s.plan_id

		WHERE u.role = 'owner'
		  AND (
				$1 = ''
				OR u.full_name ILIKE '%' || $1 || '%'
				OR COALESCE(u.email, '') ILIKE '%' || $1 || '%'
				OR COALESCE(u.phone, '') ILIKE '%' || $1 || '%'
				OR COALESCE(sp.name, '') ILIKE '%' || $1 || '%'
		  )

		ORDER BY u.created_at DESC

		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		search,
		limit,
		offset,
	)

	if err != nil {
		return CyberOwnerListResponse{}, fmt.Errorf(
			"get cyber owners: %w",
			err,
		)
	}

	defer rows.Close()

	items := make([]CyberOwnerListItem, 0)

	for rows.Next() {
		var (
			item               CyberOwnerListItem
			ownerID            uuid.UUID
			subscriptionID     *uuid.UUID
			packageName        *string
			subscriptionStatus *string
		)

		err := rows.Scan(
			&ownerID,
			&item.FullName,
			&item.Email,
			&item.Phone,
			&item.Status,
			&subscriptionID,
			&packageName,
			&item.Amount,
			&subscriptionStatus,
			&item.SubscriptionExpiresAt,
			&item.IsLifetime,
		)

		if err != nil {
			return CyberOwnerListResponse{}, fmt.Errorf(
				"scan cyber owner: %w",
				err,
			)
		}

		item.ID = ownerID.String()

		if subscriptionID != nil {
			value := subscriptionID.String()
			item.SubscriptionID = &value
		}

		item.PackageName = packageName
		item.SubscriptionStatus = subscriptionStatus

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return CyberOwnerListResponse{}, fmt.Errorf(
			"iterate cyber owners: %w",
			err,
		)
	}

	return CyberOwnerListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ============================================================
// CURRENT SUBSCRIPTIONS
// ============================================================

func (r *Repository) GetSubscriptions(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (SubscriptionListResponse, error) {

	const countQuery = `
		SELECT COUNT(*)
		FROM subscriptions s
		INNER JOIN subscription_plans sp
			ON sp.id = s.plan_id
		INNER JOIN users u
			ON u.tenant_id = s.tenant_id
		   AND u.role = 'owner'
		WHERE
			s.status = 'active'
			AND (
				s.is_lifetime = TRUE
				OR s.current_period_end IS NULL
				OR s.current_period_end > NOW()
			)
			AND (
				$1 = ''
				OR u.full_name ILIKE '%' || $1 || '%'
				OR COALESCE(u.email, '') ILIKE '%' || $1 || '%'
				OR COALESCE(u.phone, '') ILIKE '%' || $1 || '%'
				OR sp.name ILIKE '%' || $1 || '%'
			)
	`

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		search,
	).Scan(&total); err != nil {
		return SubscriptionListResponse{}, fmt.Errorf(
			"count subscriptions: %w",
			err,
		)
	}

	const query = `
		SELECT
			s.id,
			u.id,
			u.full_name,
			u.email,
			u.phone,
			sp.name,

			COALESCE(
				(
					SELECT p.amount::numeric(12,2)::text
					FROM subscription_payments p
					WHERE p.subscription_id = s.id
					  AND p.status = 'completed'
					ORDER BY p.created_at DESC
					LIMIT 1
				),
				COALESCE(sp.monthly_price, 0)::numeric(12,2)::text
			),

			s.status,
			s.is_lifetime,
			s.current_period_start,
			s.current_period_end,
			s.created_at

		FROM subscriptions s

		INNER JOIN subscription_plans sp
			ON sp.id = s.plan_id

		INNER JOIN users u
			ON u.tenant_id = s.tenant_id
		   AND u.role = 'owner'

		WHERE
			s.status = 'active'
			AND (
				s.is_lifetime = TRUE
				OR s.current_period_end IS NULL
				OR s.current_period_end > NOW()
			)
			AND (
				$1 = ''
				OR u.full_name ILIKE '%' || $1 || '%'
				OR COALESCE(u.email, '') ILIKE '%' || $1 || '%'
				OR COALESCE(u.phone, '') ILIKE '%' || $1 || '%'
				OR sp.name ILIKE '%' || $1 || '%'
			)

		ORDER BY s.created_at DESC

		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		search,
		limit,
		offset,
	)

	if err != nil {
		return SubscriptionListResponse{}, fmt.Errorf(
			"get subscriptions: %w",
			err,
		)
	}

	defer rows.Close()

	items := make([]SubscriptionListItem, 0)

	for rows.Next() {
		var (
			item    SubscriptionListItem
			id      uuid.UUID
			ownerID uuid.UUID
		)

		err := rows.Scan(
			&id,
			&ownerID,
			&item.OwnerName,
			&item.OwnerEmail,
			&item.OwnerPhone,
			&item.PackageName,
			&item.Amount,
			&item.Status,
			&item.IsLifetime,
			&item.CurrentPeriodStart,
			&item.CurrentPeriodEnd,
			&item.CreatedAt,
		)

		if err != nil {
			return SubscriptionListResponse{}, fmt.Errorf(
				"scan subscription: %w",
				err,
			)
		}

		item.ID = id.String()
		item.OwnerID = ownerID.String()

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return SubscriptionListResponse{}, fmt.Errorf(
			"iterate subscriptions: %w",
			err,
		)
	}

	return SubscriptionListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ============================================================
// REVENUE
// ============================================================

func (r *Repository) GetRevenue(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (RevenueListResponse, error) {

	const countQuery = `
		SELECT COUNT(*)
		FROM subscription_payments p
		INNER JOIN subscriptions s
			ON s.id = p.subscription_id
		INNER JOIN users u
			ON u.tenant_id = s.tenant_id
		   AND u.role = 'owner'
		WHERE
			(
				$1 = ''
				OR u.full_name ILIKE '%' || $1 || '%'
				OR COALESCE(u.email, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.phone_number, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.payment_method::text, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.status::text, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.mpesa_receipt_number, '') ILIKE '%' || $1 || '%'
			)
	`

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		search,
	).Scan(&total); err != nil {
		return RevenueListResponse{}, fmt.Errorf(
			"count platform revenue: %w",
			err,
		)
	}

	const query = `
		SELECT
			p.id,
			u.id,
			u.full_name,
			u.email,
			p.amount::numeric(12,2)::text,
			p.payment_method::text,
			p.status::text,
			p.mpesa_receipt_number,
			p.mpesa_checkout_request_id,
			p.created_at

		FROM subscription_payments p

		INNER JOIN subscriptions s
			ON s.id = p.subscription_id

		INNER JOIN users u
			ON u.tenant_id = s.tenant_id
		   AND u.role = 'owner'

		WHERE
			(
				$1 = ''
				OR u.full_name ILIKE '%' || $1 || '%'
				OR COALESCE(u.email, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.phone_number, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.payment_method::text, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.status::text, '') ILIKE '%' || $1 || '%'
				OR COALESCE(p.mpesa_receipt_number, '') ILIKE '%' || $1 || '%'
			)

		ORDER BY p.created_at DESC

		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		search,
		limit,
		offset,
	)

	if err != nil {
		return RevenueListResponse{}, fmt.Errorf(
			"get platform revenue: %w",
			err,
		)
	}

	defer rows.Close()

	items := make([]RevenueListItem, 0)

	for rows.Next() {
		var (
			item    RevenueListItem
			id      uuid.UUID
			ownerID uuid.UUID
		)

		err := rows.Scan(
			&id,
			&ownerID,
			&item.OwnerName,
			&item.OwnerEmail,
			&item.Amount,
			&item.PaymentMethod,
			&item.Status,
			&item.MPesaReceipt,
			&item.CheckoutRequestID,
			&item.CreatedAt,
		)

		if err != nil {
			return RevenueListResponse{}, fmt.Errorf(
				"scan platform revenue: %w",
				err,
			)
		}

		item.ID = id.String()
		item.OwnerID = ownerID.String()

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return RevenueListResponse{}, fmt.Errorf(
			"iterate platform revenue: %w",
			err,
		)
	}

	return RevenueListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ============================================================
// AUDIT LOGS
// ============================================================

func (r *Repository) GetAuditLogs(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (AuditLogResponse, error) {

	const countQuery = `
		SELECT COUNT(*)
		FROM audit_logs a
		LEFT JOIN users u
			ON u.id = a.user_id
		WHERE
			$1 = ''
			OR COALESCE(u.full_name, '') ILIKE '%' || $1 || '%'
			OR COALESCE(a.user_role::text, '') ILIKE '%' || $1 || '%'
			OR a.action ILIKE '%' || $1 || '%'
			OR COALESCE(a.entity_type, '') ILIKE '%' || $1 || '%'
			OR COALESCE(a.reason, '') ILIKE '%' || $1 || '%'
			OR COALESCE(a.device_id, '') ILIKE '%' || $1 || '%'
	`

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		search,
	).Scan(&total); err != nil {
		return AuditLogResponse{}, fmt.Errorf(
			"count audit logs: %w",
			err,
		)
	}

	const query = `
		SELECT
			a.id,
			a.user_id,
			u.full_name,
			a.user_role::text,
			a.action,
			a.entity_type,
			a.entity_id,
			a.old_value,
			a.new_value,
			a.reason,
			a.ip_address::text,
			a.device_id,
			a.created_at

		FROM audit_logs a

		LEFT JOIN users u
			ON u.id = a.user_id

		WHERE
			$1 = ''
			OR COALESCE(u.full_name, '') ILIKE '%' || $1 || '%'
			OR COALESCE(a.user_role::text, '') ILIKE '%' || $1 || '%'
			OR a.action ILIKE '%' || $1 || '%'
			OR COALESCE(a.entity_type, '') ILIKE '%' || $1 || '%'
			OR COALESCE(a.reason, '') ILIKE '%' || $1 || '%'
			OR COALESCE(a.device_id, '') ILIKE '%' || $1 || '%'

		ORDER BY a.created_at DESC

		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		search,
		limit,
		offset,
	)

	if err != nil {
		return AuditLogResponse{}, fmt.Errorf(
			"get audit logs: %w",
			err,
		)
	}

	defer rows.Close()

	logs := make([]AuditLog, 0)

	for rows.Next() {
		var (
			logEntry AuditLog
			id       uuid.UUID
			userID   *uuid.UUID
			entityID *uuid.UUID
			userName *string
			userRole *string
			oldValue []byte
			newValue []byte
			reason   *string
			ip       *string
			deviceID *string
		)

		err := rows.Scan(
			&id,
			&userID,
			&userName,
			&userRole,
			&logEntry.Action,
			&logEntry.EntityType,
			&entityID,
			&oldValue,
			&newValue,
			&reason,
			&ip,
			&deviceID,
			&logEntry.CreatedAt,
		)

		if err != nil {
			return AuditLogResponse{}, fmt.Errorf(
				"scan audit log: %w",
				err,
			)
		}

		logEntry.ID = id.String()

		if userID != nil {
			value := userID.String()
			logEntry.UserID = &value
		}

		if entityID != nil {
			value := entityID.String()
			logEntry.EntityID = &value
		}

		logEntry.UserName = userName
		logEntry.UserRole = userRole
		logEntry.OldValue = oldValue
		logEntry.NewValue = newValue
		logEntry.Reason = reason
		logEntry.IPAddress = ip
		logEntry.DeviceID = deviceID

		logs = append(logs, logEntry)
	}

	if err := rows.Err(); err != nil {
		return AuditLogResponse{}, fmt.Errorf(
			"iterate audit logs: %w",
			err,
		)
	}

	return AuditLogResponse{
		Items:  logs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
