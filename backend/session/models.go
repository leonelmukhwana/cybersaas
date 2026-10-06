package session

import "time"

type Session struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenant_id"`
	BranchID    string  `json:"branch_id"`
	TerminalID  string  `json:"terminal_id"`
	CustomerID  string  `json:"customer_id"`
	AttendantID *string `json:"attendant_id,omitempty"`

	// Safe display fields for the dashboard.
	// Sensitive customer ID numbers are intentionally excluded.
	CustomerName string  `json:"customer_name"`
	CustomerType string  `json:"customer_type"`
	ParentName   *string `json:"parent_name,omitempty"`
	TerminalName *string `json:"terminal_name,omitempty"`

	SessionType        string     `json:"session_type"`
	Status             string     `json:"status"`
	StartedAt          time.Time  `json:"started_at"`
	EndedAt            *time.Time `json:"ended_at,omitempty"`
	PausedAt           *time.Time `json:"paused_at,omitempty"`
	TotalPausedSeconds int64      `json:"total_paused_seconds"`
	PrepaidAmount      *string    `json:"prepaid_amount,omitempty"`
	AllowedMinutes     *int       `json:"allowed_minutes,omitempty"`
	RatePerMinute      *string    `json:"rate_per_minute,omitempty"`
	MinimumCharge      *string    `json:"minimum_charge,omitempty"`
	FinalAmount        *string    `json:"final_amount,omitempty"`
	ClientOperationID  string     `json:"client_operation_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type StartSessionRequest struct {
	BranchID          string     `json:"branch_id" binding:"required"`
	TerminalID        string     `json:"terminal_id" binding:"required"`
	CustomerID        string     `json:"customer_id" binding:"required"`
	SessionType       string     `json:"session_type" binding:"required"`
	PrepaidAmount     *string    `json:"prepaid_amount"`
	AllowedMinutes    *int       `json:"allowed_minutes"`
	StartedAt         *time.Time `json:"started_at"`
	ClientOperationID string     `json:"client_operation_id" binding:"required"`
}

type EndSessionRequest struct {
	EndedAt *time.Time `json:"ended_at"`
}

type CancelSessionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type SessionListResponse struct {
	Sessions []Session `json:"sessions"`
	Total    int64     `json:"total"`
}

type BillingPreview struct {
	SessionID        string `json:"session_id"`
	ElapsedMinutes   int64  `json:"elapsed_minutes"`
	BillableMinutes  int64  `json:"billable_minutes"`
	RatePerMinute    string `json:"rate_per_minute"`
	MinimumCharge    string `json:"minimum_charge"`
	CalculatedAmount string `json:"calculated_amount"`
	Currency         string `json:"currency"`
}

type SessionWithBilling struct {
	Session
	ElapsedMinutes   int64  `json:"elapsed_minutes"`
	BillableMinutes  int64  `json:"billable_minutes"`
	CalculatedAmount string `json:"calculated_amount"`
	Currency         string `json:"currency"`
}
