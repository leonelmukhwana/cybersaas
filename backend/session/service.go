package session

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSessionID      = errors.New("invalid session id")
	ErrInvalidBranchID       = errors.New("invalid branch id")
	ErrInvalidTerminalID     = errors.New("invalid terminal id")
	ErrInvalidCustomerID     = errors.New("invalid customer id")
	ErrInvalidOperationID    = errors.New("invalid client operation id")
	ErrInvalidSessionType    = errors.New("session_type must be pay_after or prepaid")
	ErrInvalidAmount         = errors.New("invalid monetary amount")
	ErrInvalidMinutes        = errors.New("allowed_minutes must be greater than zero")
	ErrPrepaidAmountMissing  = errors.New("prepaid_amount is required for prepaid sessions")
	ErrPrepaidMinutesMissing = errors.New("allowed_minutes is required for prepaid sessions")
)

var moneyPattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Start(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	request StartSessionRequest,
) (*Session, error) {
	if _, err := uuid.Parse(request.BranchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(request.TerminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(request.CustomerID); err != nil {
		return nil, ErrInvalidCustomerID
	}

	if _, err := uuid.Parse(request.ClientOperationID); err != nil {
		return nil, ErrInvalidOperationID
	}

	sessionType := strings.ToLower(
		strings.TrimSpace(request.SessionType),
	)

	if sessionType != "pay_after" &&
		sessionType != "prepaid" {
		return nil, ErrInvalidSessionType
	}

	if request.AllowedMinutes != nil &&
		*request.AllowedMinutes <= 0 {
		return nil, ErrInvalidMinutes
	}

	if request.PrepaidAmount != nil {
		value := strings.TrimSpace(*request.PrepaidAmount)

		if !moneyPattern.MatchString(value) {
			return nil, ErrInvalidAmount
		}

		*request.PrepaidAmount = value
	}

	if sessionType == "prepaid" {
		if request.PrepaidAmount == nil ||
			strings.TrimSpace(*request.PrepaidAmount) == "" {
			return nil, ErrPrepaidAmountMissing
		}

		if request.AllowedMinutes == nil ||
			*request.AllowedMinutes <= 0 {
			return nil, ErrPrepaidMinutesMissing
		}
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		request.BranchID,
	); err != nil {
		return nil, err
	}

	if userRole == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			userID,
			tenantID,
			request.BranchID,
		); err != nil {
			return nil, err
		}
	}

	if err := s.repository.VerifyTerminal(
		ctx,
		tenantID,
		request.BranchID,
		request.TerminalID,
	); err != nil {
		return nil, err
	}

	if err := s.repository.VerifyCustomer(
		ctx,
		tenantID,
		request.BranchID,
		request.CustomerID,
	); err != nil {
		return nil, err
	}

	rate, minimum, _, _, _, err :=
		s.repository.GetBillingConfig(
			ctx,
			tenantID,
			request.BranchID,
		)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now().UTC()

	if request.StartedAt != nil {
		startedAt = request.StartedAt.UTC()
	}

	var attendantID *string

	if userRole == "attendant" {
		attendantID = &userID
	}

	result := Session{
		ID:                newUUID(),
		TenantID:          tenantID,
		BranchID:          request.BranchID,
		TerminalID:        request.TerminalID,
		CustomerID:        request.CustomerID,
		AttendantID:       attendantID,
		SessionType:       sessionType,
		Status:            "active",
		StartedAt:         startedAt,
		PrepaidAmount:     request.PrepaidAmount,
		AllowedMinutes:    request.AllowedMinutes,
		RatePerMinute:     &rate,
		MinimumCharge:     &minimum,
		ClientOperationID: request.ClientOperationID,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	created, err := s.repository.Create(
		ctx,
		result,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateAuditLog(
		ctx,
		tenantID,
		request.BranchID,
		userID,
		userRole,
		"session_started",
		created.ID,
		"",
	); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	sessionID string,
) (*Session, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, ErrInvalidSessionID
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		branchID,
	); err != nil {
		return nil, err
	}

	if userRole == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			userID,
			tenantID,
			branchID,
		); err != nil {
			return nil, err
		}
	}

	return s.repository.GetLatestForID(
		ctx,
		tenantID,
		branchID,
		sessionID,
	)
}

func (s *Service) PreviewBilling(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	sessionID string,
	at time.Time,
) (*BillingPreview, error) {
	session, err := s.Get(
		ctx,
		tenantID,
		userID,
		userRole,
		branchID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}

	calculated, elapsed, billable :=
		calculateBilling(session, at)

	currency := "KES"

	_, _, _, _, configuredCurrency, configErr :=
		s.repository.GetBillingConfig(
			ctx,
			tenantID,
			branchID,
		)

	if configErr == nil &&
		configuredCurrency != "" {
		currency = configuredCurrency
	}

	return &BillingPreview{
		SessionID:        session.ID,
		ElapsedMinutes:   elapsed,
		BillableMinutes:  billable,
		RatePerMinute:    valueOrZero(session.RatePerMinute),
		MinimumCharge:    valueOrZero(session.MinimumCharge),
		CalculatedAmount: calculated,
		Currency:         currency,
	}, nil
}

func (s *Service) End(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	sessionID string,
	request EndSessionRequest,
) (*SessionWithBilling, error) {
	session, err := s.Get(
		ctx,
		tenantID,
		userID,
		userRole,
		branchID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" {
		return nil, ErrInvalidSessionState
	}

	endedAt := time.Now().UTC()

	if request.EndedAt != nil {
		endedAt = request.EndedAt.UTC()
	}

	if endedAt.Before(session.StartedAt) {
		return nil, errors.New(
			"ended_at cannot be before started_at",
		)
	}

	calculated, elapsed, billable :=
		calculateBilling(session, endedAt)

	result, err := s.repository.End(
		ctx,
		tenantID,
		branchID,
		session.ID,
		session.StartedAt,
		endedAt,
		calculated,
		"completed",
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"session_completed",
		result.ID,
		"",
	); err != nil {
		return nil, err
	}

	currency := "KES"

	_, _, _, _, configuredCurrency, configErr :=
		s.repository.GetBillingConfig(
			ctx,
			tenantID,
			branchID,
		)

	if configErr == nil &&
		configuredCurrency != "" {
		currency = configuredCurrency
	}

	return &SessionWithBilling{
		Session:          *result,
		ElapsedMinutes:   elapsed,
		BillableMinutes:  billable,
		CalculatedAmount: calculated,
		Currency:         currency,
	}, nil
}

func (s *Service) Cancel(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	sessionID string,
	reason string,
) (*Session, error) {
	reason = strings.TrimSpace(reason)

	if reason == "" {
		return nil, errors.New("cancellation reason is required")
	}

	if len(reason) > 500 {
		return nil, errors.New(
			"cancellation reason must not exceed 500 characters",
		)
	}

	session, err := s.Get(
		ctx,
		tenantID,
		userID,
		userRole,
		branchID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" {
		return nil, ErrInvalidSessionState
	}

	result, err := s.repository.Cancel(
		ctx,
		tenantID,
		branchID,
		session.ID,
		session.StartedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"session_cancelled",
		result.ID,
		reason,
	); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) List(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	status string,
	limit int,
	offset int,
) (*SessionListResponse, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		branchID,
	); err != nil {
		return nil, err
	}

	if userRole == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			userID,
			tenantID,
			branchID,
		); err != nil {
			return nil, err
		}
	}

	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	status = strings.TrimSpace(
		strings.ToLower(status),
	)

	if status != "" {
		switch status {
		case "active",
			"paused",
			"completed",
			"cancelled":
		default:
			return nil, errors.New("invalid session status")
		}
	}

	/*
		Retention rule:

		- Sessions remain permanently in the database.
		- Attendants only see sessions for the current
		  calendar day in Kenya.
		- Owners can see historical sessions.
		- Date filtering is applied at the repository level,
		  so both rows and total use the same restriction.
	*/
	var startTime *time.Time
	var endTime *time.Time

	if userRole == "attendant" {
		start, end := todayKenyaRange(time.Now())

		startTime = &start
		endTime = &end
	}

	results, err := s.repository.List(
		ctx,
		tenantID,
		branchID,
		status,
		startTime,
		endTime,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	total, err := s.repository.Count(
		ctx,
		tenantID,
		branchID,
		status,
		startTime,
		endTime,
	)
	if err != nil {
		return nil, err
	}

	return &SessionListResponse{
		Sessions: results,
		Total:    total,
	}, nil
}

func (s *Service) PauseFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
	sessionID string,
) (*Session, error) {
	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, ErrInvalidSessionID
	}

	if err := s.repository.VerifyTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	); err != nil {
		return nil, err
	}

	session, err := s.repository.GetActiveForTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	)
	if err != nil {
		return nil, err
	}

	if session.ID != sessionID {
		return nil, ErrSessionNotFound
	}

	if session.Status != "active" ||
		session.PausedAt != nil {
		return nil, ErrInvalidSessionState
	}

	return s.repository.PauseSession(
		ctx,
		tenantID,
		branchID,
		sessionID,
		session.StartedAt,
	)
}

func (s *Service) Pause(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	sessionID string,
) (*Session, error) {
	session, err := s.Get(
		ctx,
		tenantID,
		userID,
		userRole,
		branchID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" ||
		session.PausedAt != nil {
		return nil, ErrInvalidSessionState
	}

	result, err := s.repository.PauseSession(
		ctx,
		tenantID,
		branchID,
		sessionID,
		session.StartedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"session_paused",
		result.ID,
		"",
	); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) Resume(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	sessionID string,
) (*Session, error) {
	session, err := s.Get(
		ctx,
		tenantID,
		userID,
		userRole,
		branchID,
		sessionID,
	)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" ||
		session.PausedAt == nil {
		return nil, ErrInvalidSessionState
	}

	result, err := s.repository.ResumeSession(
		ctx,
		tenantID,
		branchID,
		sessionID,
		session.StartedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"session_resumed",
		result.ID,
		"",
	); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) ResumeFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
	sessionID string,
) (*Session, error) {
	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, ErrInvalidSessionID
	}

	if err := s.repository.VerifyTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	); err != nil {
		return nil, err
	}

	session, err := s.repository.GetActiveForTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	)
	if err != nil {
		return nil, err
	}

	if session.ID != sessionID {
		return nil, ErrSessionNotFound
	}

	if session.Status != "active" ||
		session.PausedAt == nil {
		return nil, ErrInvalidSessionState
	}

	return s.repository.ResumeSession(
		ctx,
		tenantID,
		branchID,
		sessionID,
		session.StartedAt,
	)
}

/*
todayKenyaRange returns the UTC boundaries corresponding to
midnight today and midnight tomorrow in Kenya.

Example:

Kenya:
2026-09-30 00:00:00

	|
	| converted to UTC
	v

2026-09-29 21:00:00 UTC

This keeps the database timestamps in UTC while ensuring
the attendant dashboard follows the Kenyan calendar day.
*/
func todayKenyaRange(now time.Time) (time.Time, time.Time) {
	location, err := time.LoadLocation("Africa/Nairobi")

	if err != nil {
		location = time.FixedZone(
			"EAT",
			3*60*60,
		)
	}

	kenyaNow := now.In(location)

	start := time.Date(
		kenyaNow.Year(),
		kenyaNow.Month(),
		kenyaNow.Day(),
		0,
		0,
		0,
		0,
		location,
	)

	end := start.AddDate(
		0,
		0,
		1,
	)

	return start.UTC(), end.UTC()
}

func calculateBilling(
	session *Session,
	at time.Time,
) (string, int64, int64) {
	if at.Before(session.StartedAt) {
		at = session.StartedAt
	}

	activeSeconds := int64(
		at.Sub(session.StartedAt).Seconds(),
	)

	activeSeconds -= session.TotalPausedSeconds

	if session.PausedAt != nil &&
		at.After(*session.PausedAt) {
		activeSeconds -= int64(
			at.Sub(*session.PausedAt).Seconds(),
		)
	}

	if activeSeconds < 0 {
		activeSeconds = 0
	}

	elapsedMinutes := int64(
		math.Ceil(float64(activeSeconds) / 60.0),
	)

	if elapsedMinutes < 0 {
		elapsedMinutes = 0
	}

	billableMinutes := elapsedMinutes

	if session.AllowedMinutes != nil &&
		*session.AllowedMinutes > 0 &&
		billableMinutes > int64(*session.AllowedMinutes) {
		billableMinutes = int64(*session.AllowedMinutes)
	}

	rate := parseMoney(
		valueOrZero(session.RatePerMinute),
	)

	minimum := parseMoney(
		valueOrZero(session.MinimumCharge),
	)

	amount := float64(billableMinutes) * rate

	if amount < minimum {
		amount = minimum
	}

	if session.SessionType == "prepaid" &&
		session.PrepaidAmount != nil {
		prepaid := parseMoney(*session.PrepaidAmount)

		if amount > prepaid {
			amount = prepaid
		}
	}

	return formatMoney(amount), elapsedMinutes, billableMinutes
}

func parseMoney(value string) float64 {
	value = strings.TrimSpace(value)

	if value == "" {
		return 0
	}

	result, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return 0
	}

	return result
}

func formatMoney(value float64) string {
	return strconv.FormatFloat(
		math.Round(value*100)/100,
		'f',
		2,
		64,
	)
}

func valueOrZero(value *string) string {
	if value == nil {
		return "0.00"
	}

	return *value
}
