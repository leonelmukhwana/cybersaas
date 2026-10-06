package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEventNotFound    = errors.New("sync event not found")
	ErrInvalidEventID   = errors.New("invalid event id")
	ErrInvalidEntityID  = errors.New("invalid entity id")
	ErrSyncAccessDenied = errors.New("sync access denied")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateEvent(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
	request PushEventRequest,
) (*SyncEvent, bool, error) {
	eventID, err := uuid.Parse(request.ID)
	if err != nil {
		return nil, false, ErrInvalidEventID
	}

	entityID, err := uuid.Parse(request.EntityID)
	if err != nil {
		return nil, false, ErrInvalidEntityID
	}

	payload, err := json.Marshal(request.Payload)
	if err != nil {
		return nil, false, fmt.Errorf("marshal sync payload: %w", err)
	}

	var userID *string
	if request.UserID != "" {
		if _, err := uuid.Parse(request.UserID); err != nil {
			return nil, false, errors.New("invalid user id")
		}

		userID = &request.UserID
	}

	var event SyncEvent

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO sync_events (
			id,
			tenant_id,
			branch_id,
			terminal_id,
			user_id,
			event_type,
			entity_type,
			entity_id,
			payload,
			status,
			attempts
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending',0)
		ON CONFLICT (id) DO NOTHING
		RETURNING
			id,
			tenant_id,
			branch_id,
			terminal_id,
			user_id,
			event_type,
			entity_type,
			entity_id,
			payload,
			status,
			attempts,
			last_error,
			created_at,
			processed_at
		`,
		eventID,
		tenantID,
		branchID,
		terminalID,
		userID,
		request.EventType,
		request.EntityType,
		entityID,
		payload,
	).Scan(
		&event.ID,
		&event.TenantID,
		&event.BranchID,
		&event.TerminalID,
		&event.UserID,
		&event.EventType,
		&event.EntityType,
		&event.EntityID,
		&payload,
		&event.Status,
		&event.Attempts,
		&event.LastError,
		&event.CreatedAt,
		&event.ProcessedAt,
	)

	if err == nil {
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &event.Payload); err != nil {
				return nil, false, fmt.Errorf("decode stored payload: %w", err)
			}
		}

		return &event, false, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, fmt.Errorf("create sync event: %w", err)
	}

	existing, err := r.GetEventForTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
		eventID.String(),
	)
	if err != nil {
		return nil, false, err
	}

	return existing, true, nil
}

func (r *Repository) GetEventForTerminal(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
	eventID string,
) (*SyncEvent, error) {
	var event SyncEvent
	var payload []byte

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			terminal_id,
			user_id,
			event_type,
			entity_type,
			entity_id,
			payload,
			status,
			attempts,
			last_error,
			created_at,
			processed_at
		FROM sync_events
		WHERE id = $1
		  AND tenant_id = $2
		  AND branch_id = $3
		  AND terminal_id = $4
		`,
		eventID,
		tenantID,
		branchID,
		terminalID,
	).Scan(
		&event.ID,
		&event.TenantID,
		&event.BranchID,
		&event.TerminalID,
		&event.UserID,
		&event.EventType,
		&event.EntityType,
		&event.EntityID,
		&payload,
		&event.Status,
		&event.Attempts,
		&event.LastError,
		&event.CreatedAt,
		&event.ProcessedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEventNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get sync event: %w", err)
	}

	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &event.Payload); err != nil {
			return nil, fmt.Errorf("decode sync payload: %w", err)
		}
	}

	return &event, nil
}

func (r *Repository) PullEvents(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
	since *time.Time,
	limit int,
) ([]SyncEvent, bool, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			terminal_id,
			user_id,
			event_type,
			entity_type,
			entity_id,
			payload,
			status,
			attempts,
			last_error,
			created_at,
			processed_at
		FROM sync_events
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND (
				terminal_id IS NULL
				OR terminal_id <> $3
		  )
		  AND ($4::timestamptz IS NULL OR created_at > $4)
		ORDER BY created_at ASC, id ASC
		LIMIT $5
		`,
		tenantID,
		branchID,
		terminalID,
		since,
		limit+1,
	)
	if err != nil {
		return nil, false, fmt.Errorf("pull sync events: %w", err)
	}
	defer rows.Close()

	events := make([]SyncEvent, 0, limit)
	hasMore := false

	for rows.Next() {
		var event SyncEvent
		var payload []byte

		if err := rows.Scan(
			&event.ID,
			&event.TenantID,
			&event.BranchID,
			&event.TerminalID,
			&event.UserID,
			&event.EventType,
			&event.EntityType,
			&event.EntityID,
			&payload,
			&event.Status,
			&event.Attempts,
			&event.LastError,
			&event.CreatedAt,
			&event.ProcessedAt,
		); err != nil {
			return nil, false, fmt.Errorf("scan sync event: %w", err)
		}

		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &event.Payload); err != nil {
				return nil, false, fmt.Errorf("decode sync event payload: %w", err)
			}
		}

		if len(events) == limit {
			hasMore = true
			break
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate sync events: %w", err)
	}

	return events, hasMore, nil
}

func (r *Repository) UpdateSyncState(
	ctx context.Context,
	terminalID string,
	pullAt *time.Time,
	pushAt *time.Time,
	successAt *time.Time,
	syncError *string,
) error {
	id := uuid.New()

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO terminal_sync_state (
			id,
			terminal_id,
			last_pull_at,
			last_push_at,
			last_successful_sync_at,
			last_sync_error
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (terminal_id)
		DO UPDATE SET
			last_pull_at = COALESCE(EXCLUDED.last_pull_at, terminal_sync_state.last_pull_at),
			last_push_at = COALESCE(EXCLUDED.last_push_at, terminal_sync_state.last_push_at),
			last_successful_sync_at = COALESCE(
				EXCLUDED.last_successful_sync_at,
				terminal_sync_state.last_successful_sync_at
			),
			last_sync_error = EXCLUDED.last_sync_error,
			updated_at = NOW()
		`,
		id,
		terminalID,
		pullAt,
		pushAt,
		successAt,
		syncError,
	)
	if err != nil {
		return fmt.Errorf("update terminal sync state: %w", err)
	}

	return nil
}

func (r *Repository) GetSyncState(
	ctx context.Context,
	terminalID string,
) (*SyncState, error) {
	var state SyncState

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			terminal_id,
			last_pull_at,
			last_push_at,
			last_successful_sync_at,
			last_sync_error,
			updated_at
		FROM terminal_sync_state
		WHERE terminal_id = $1
		`,
		terminalID,
	).Scan(
		&state.ID,
		&state.TerminalID,
		&state.LastPullAt,
		&state.LastPushAt,
		&state.LastSuccessfulSyncAt,
		&state.LastSyncError,
		&state.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return &SyncState{
			TerminalID: terminalID,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get terminal sync state: %w", err)
	}

	return &state, nil
}
