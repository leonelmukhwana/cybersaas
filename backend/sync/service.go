package sync

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEventType  = errors.New("event type is required")
	ErrInvalidEntityType = errors.New("entity type is required")
	ErrInvalidLimit      = errors.New("invalid sync limit")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) PushEvent(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
	request PushEventRequest,
) (*PushEventResponse, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, errors.New("invalid branch id")
	}

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, errors.New("invalid terminal id")
	}

	request.EventType = strings.TrimSpace(request.EventType)
	request.EntityType = strings.TrimSpace(request.EntityType)

	if request.EventType == "" {
		return nil, ErrInvalidEventType
	}

	if request.EntityType == "" {
		return nil, ErrInvalidEntityType
	}

	if len(request.EventType) > 100 {
		return nil, errors.New("event type is too long")
	}

	if len(request.EntityType) > 100 {
		return nil, errors.New("entity type is too long")
	}

	event, alreadyKnown, err := s.repository.CreateEvent(
		ctx,
		tenantID,
		branchID,
		terminalID,
		request,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	if err := s.repository.UpdateSyncState(
		ctx,
		terminalID,
		nil,
		&now,
		nil,
		nil,
	); err != nil {
		return nil, err
	}

	status := event.Status

	if alreadyKnown {
		return &PushEventResponse{
			ID:     event.ID,
			Status: "already_known",
		}, nil
	}

	return &PushEventResponse{
		ID:     event.ID,
		Status: status,
	}, nil
}

func (s *Service) PullEvents(
	ctx context.Context,
	tenantID string,
	branchID string,
	terminalID string,
	since *time.Time,
	limit int,
) (*PullEventsResponse, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	events, hasMore, err := s.repository.PullEvents(
		ctx,
		tenantID,
		branchID,
		terminalID,
		since,
		limit,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	if err := s.repository.UpdateSyncState(
		ctx,
		terminalID,
		&now,
		nil,
		nil,
		nil,
	); err != nil {
		return nil, err
	}

	return &PullEventsResponse{
		Events:     events,
		HasMore:    hasMore,
		ServerTime: now,
	}, nil
}

func (s *Service) MarkSyncSuccessful(
	ctx context.Context,
	terminalID string,
) error {
	now := time.Now().UTC()

	return s.repository.UpdateSyncState(
		ctx,
		terminalID,
		nil,
		nil,
		&now,
		nil,
	)
}

func (s *Service) RecordSyncError(
	ctx context.Context,
	terminalID string,
	err error,
) error {
	if err == nil {
		return nil
	}

	message := err.Error()

	return s.repository.UpdateSyncState(
		ctx,
		terminalID,
		nil,
		nil,
		nil,
		&message,
	)
}

func (s *Service) GetState(
	ctx context.Context,
	terminalID string,
) (*SyncState, error) {
	return s.repository.GetSyncState(ctx, terminalID)
}
