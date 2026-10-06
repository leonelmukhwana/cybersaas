package session

import "time"

type TerminalStartSessionRequest struct {
	CustomerID        string     `json:"customer_id" binding:"required"`
	SessionType       string     `json:"session_type" binding:"required"`
	PrepaidAmount     *string    `json:"prepaid_amount"`
	AllowedMinutes    *int       `json:"allowed_minutes"`
	StartedAt         *time.Time `json:"started_at"`
	ClientOperationID string     `json:"client_operation_id" binding:"required"`
}

type TerminalEndSessionRequest struct {
	EndedAt *time.Time `json:"ended_at"`
}
