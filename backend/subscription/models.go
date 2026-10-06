package subscription

import "time"

// Plan represents a SaaS subscription plan.
type Plan struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	IncludedBranches  int       `json:"included_branches"`
	IncludedTerminals int       `json:"included_terminals"`
	ExtraBranchRate   string    `json:"extra_branch_rate"`
	ExtraTerminalRate string    `json:"extra_terminal_rate"`
	MonthlyPrice      string    `json:"monthly_price"`
	IsLifetime        bool      `json:"is_lifetime"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

// Subscription represents one tenant's current SaaS subscription.
type Subscription struct {
	ID       string  `json:"id"`
	TenantID string  `json:"tenant_id"`
	PlanID   *string `json:"plan_id,omitempty"` // Pointer to allow NULL during 7-day trial before plan selection
	PlanName string  `json:"plan_name,omitempty"`

	Status      string     `json:"status"`
	IsTrial     bool       `json:"is_trial"`
	TrialEndsAt *time.Time `json:"trial_ends_at,omitempty"` // Timestamp when the 7-day trial expires
	IsLifetime  bool       `json:"is_lifetime"`

	AccountBalance string `json:"account_balance"`

	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`

	SelectedBranchesLimit  int `json:"selected_branches_limit"`  // Branches limit active or selected for next renewal
	SelectedTerminalsLimit int `json:"selected_terminals_limit"` // Terminals limit active or selected for next renewal

	PeakBranches  int `json:"peak_branches"`
	PeakTerminals int `json:"peak_terminals"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionPayment represents a payment made by a Cyber Owner
// toward their SaaS subscription.
type SubscriptionPayment struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`

	Amount      string `json:"amount"`
	PhoneNumber string `json:"phone_number"`

	PaymentMethod string `json:"payment_method"`

	MpesaCheckoutRequestID *string `json:"mpesa_checkout_request_id,omitempty"`
	MpesaReceiptNumber     *string `json:"mpesa_receipt_number,omitempty"`

	Status        string  `json:"status"`
	FailureReason *string `json:"failure_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LedgerEntry represents an immutable subscription financial entry.
type LedgerEntry struct {
	ID             string  `json:"id"`
	SubscriptionID string  `json:"subscription_id"`
	TenantID       string  `json:"tenant_id"`
	PaymentID      *string `json:"payment_id,omitempty"`

	EntryType    string    `json:"entry_type"`
	Amount       string    `json:"amount"`
	BalanceAfter string    `json:"balance_after"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

// SubscriptionOverview is used by the Cyber Owner dashboard.
type SubscriptionOverview struct {
	Subscription Subscription `json:"subscription"`
	Plan         *Plan        `json:"plan,omitempty"` // Pointer to accommodate trial state prior to plan selection

	BranchesUsed  int `json:"branches_used"`
	BranchesLimit int `json:"branches_limit"`

	TerminalsUsed  int `json:"terminals_used"`
	TerminalsLimit int `json:"terminals_limit"`

	IsActive               bool   `json:"is_active"`
	IsExpired              bool   `json:"is_expired"`
	IsTrial                bool   `json:"is_trial"`
	IsTrialExpired         bool   `json:"is_trial_expired"`
	MustSelectPackage      bool   `json:"must_select_package"` // Triggers paywall/package selection UI in dashboard
	SuggestedMonthlyAmount string `json:"suggested_monthly_amount,omitempty"`
	IsLifetime             bool   `json:"is_lifetime"`
}
