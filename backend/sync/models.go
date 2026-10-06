package sync

import "time"

type SyncEvent struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	BranchID    string     `json:"branch_id"`
	TerminalID  *string    `json:"terminal_id,omitempty"`
	UserID      *string    `json:"user_id,omitempty"`
	EventType   string     `json:"event_type"`
	EntityType  string     `json:"entity_type"`
	EntityID    string     `json:"entity_id"`
	Payload     any        `json:"payload"`
	Status      string     `json:"status"`
	Attempts    int        `json:"attempts"`
	LastError   *string    `json:"last_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

type PushEventRequest struct {
	ID         string `json:"id" binding:"required"`
	EventType  string `json:"event_type" binding:"required"`
	EntityType string `json:"entity_type" binding:"required"`
	EntityID   string `json:"entity_id" binding:"required"`
	Payload    any    `json:"payload" binding:"required"`
	UserID     string `json:"user_id,omitempty"`
}

type PushEventResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type PullEventsResponse struct {
	Events     []SyncEvent `json:"events"`
	HasMore    bool        `json:"has_more"`
	ServerTime time.Time   `json:"server_time"`
}

type SyncState struct {
	ID                   string     `json:"id"`
	TerminalID           string     `json:"terminal_id"`
	LastPullAt           *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt           *time.Time `json:"last_push_at,omitempty"`
	LastSuccessfulSyncAt *time.Time `json:"last_successful_sync_at,omitempty"`
	LastSyncError        *string    `json:"last_sync_error,omitempty"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type SyncResult struct {
	Pushed       int `json:"pushed"`
	AlreadyKnown int `json:"already_known"`
	Failed       int `json:"failed"`
}
