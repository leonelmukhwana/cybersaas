package platform

import "time"

// ============================================================
// DASHBOARD
// ============================================================

type DashboardStats struct {
	TotalCyberOwners      int64 `json:"total_cyber_owners"`
	ActiveCyberOwners     int64 `json:"active_cyber_owners"`
	SuspendedCyberOwners  int64 `json:"suspended_cyber_owners"`
	TotalCyberLocations   int64 `json:"total_cyber_locations"`
	ActiveCyberLocations  int64 `json:"active_cyber_locations"`
	ActiveSubscriptions   int64 `json:"active_subscriptions"`
	PastDueSubscriptions  int64 `json:"past_due_subscriptions"`
	ExpiredSubscriptions  int64 `json:"expired_subscriptions"`
	LifetimeSubscriptions int64 `json:"lifetime_subscriptions"`

	SubscriptionRevenueAllTime    string `json:"subscription_revenue_all_time"`
	SubscriptionRevenueMonth      string `json:"subscription_revenue_month"`
	SubscriptionRevenueYear       string `json:"subscription_revenue_year"`
	SubscriptionRevenueThreeYears string `json:"subscription_revenue_three_years"`
}

// ============================================================
// AUDIT LOGS
// ============================================================

type AuditLog struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id,omitempty"`
	UserName   *string   `json:"user_name,omitempty"`
	UserRole   *string   `json:"user_role,omitempty"`
	Action     string    `json:"action"`
	EntityType *string   `json:"entity_type,omitempty"`
	EntityID   *string   `json:"entity_id,omitempty"`
	OldValue   []byte    `json:"old_value,omitempty"`
	NewValue   []byte    `json:"new_value,omitempty"`
	Reason     *string   `json:"reason,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	DeviceID   *string   `json:"device_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLogResponse struct {
	Items  []AuditLog `json:"items"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// ============================================================
// CYBER OWNERS
// ============================================================

type CyberOwnerListItem struct {
	ID                    string     `json:"id"`
	FullName              string     `json:"full_name"`
	Email                 *string    `json:"email,omitempty"`
	Phone                 *string    `json:"phone,omitempty"`
	Status                string     `json:"status"`
	SubscriptionID        *string    `json:"subscription_id,omitempty"`
	PackageName           *string    `json:"package_name,omitempty"`
	Amount                string     `json:"amount"`
	SubscriptionStatus    *string    `json:"subscription_status,omitempty"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	IsLifetime            bool       `json:"is_lifetime"`
}

type CyberOwnerListResponse struct {
	Items  []CyberOwnerListItem `json:"items"`
	Total  int64                `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

// ============================================================
// SUBSCRIPTIONS
// ============================================================

type SubscriptionListItem struct {
	ID                 string     `json:"id"`
	OwnerID            string     `json:"owner_id"`
	OwnerName          string     `json:"owner_name"`
	OwnerEmail         *string    `json:"owner_email,omitempty"`
	OwnerPhone         *string    `json:"owner_phone,omitempty"`
	PackageName        string     `json:"package_name"`
	Amount             string     `json:"amount"`
	Status             string     `json:"status"`
	IsLifetime         bool       `json:"is_lifetime"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type SubscriptionListResponse struct {
	Items  []SubscriptionListItem `json:"items"`
	Total  int64                  `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

// ============================================================
// REVENUE
// ============================================================

type RevenueListItem struct {
	ID                string    `json:"id"`
	OwnerID           string    `json:"owner_id"`
	OwnerName         string    `json:"owner_name"`
	OwnerEmail        *string   `json:"owner_email,omitempty"`
	Amount            string    `json:"amount"`
	PaymentMethod     string    `json:"payment_method"`
	Status            string    `json:"status"`
	MPesaReceipt      *string   `json:"mpesa_receipt_number,omitempty"`
	CheckoutRequestID *string   `json:"mpesa_checkout_request_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type RevenueListResponse struct {
	Items  []RevenueListItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
