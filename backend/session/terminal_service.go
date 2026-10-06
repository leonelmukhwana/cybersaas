package session

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) StartFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
	request TerminalStartSessionRequest,
) (*Session, error) {
	terminalID = strings.TrimSpace(terminalID)
	tenantID = strings.TrimSpace(tenantID)
	branchID = strings.TrimSpace(branchID)

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
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

	/*
		Prepaid sessions are calculated by the backend.

		The terminal supplies only the amount paid.

		The terminal must never decide:
		- the branch rate
		- the number of minutes
		- the amount due
	*/
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

		if parseMoney(*request.PrepaidAmount) <= 0 {
			return nil, ErrInvalidAmount
		}
	}

	/*
		Important security check:

		The terminal supplies its own terminal ID through
		authentication, but we still verify that the terminal
		belongs to the authenticated tenant and branch.
	*/
	if err := s.repository.VerifyTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	); err != nil {
		return nil, err
	}

	if err := s.repository.VerifyCustomer(
		ctx,
		tenantID,
		branchID,
		request.CustomerID,
	); err != nil {
		return nil, err
	}

	/*
		Existing branch billing configuration is the source
		of truth.

		The terminal does not provide the rate.
	*/
	rate, minimum, _, _, _, err :=
		s.repository.GetBillingConfig(
			ctx,
			tenantID,
			branchID,
		)

	if err != nil {
		return nil, err
	}

	/*
		For prepaid sessions, calculate the purchased minutes
		from the amount paid and the configured branch rate.

		Example:

		    KSh 10 / KSh 1 per minute = 10 minutes
		    KSh 50 / KSh 1 per minute = 50 minutes

		Fractional minutes are not stored, so we round down.
	*/
	var allowedMinutes *int

	if sessionType == "prepaid" {
		rateValue := parseMoney(rate)
		amountValue := parseMoney(*request.PrepaidAmount)

		if rateValue <= 0 {
			return nil, errors.New(
				"branch billing rate must be greater than zero",
			)
		}

		calculatedMinutes :=
			int(math.Floor(amountValue / rateValue))

		if calculatedMinutes <= 0 {
			return nil, errors.New(
				"prepaid amount is too low for at least one minute",
			)
		}

		allowedMinutes = &calculatedMinutes
	}

	result := Session{
		ID:                newUUID(),
		TenantID:          tenantID,
		BranchID:          branchID,
		TerminalID:        terminalID,
		CustomerID:        request.CustomerID,
		SessionType:       sessionType,
		Status:            "active",
		StartedAt:         nowUTC(),
		PrepaidAmount:     request.PrepaidAmount,
		AllowedMinutes:    allowedMinutes,
		RatePerMinute:     &rate,
		MinimumCharge:     &minimum,
		ClientOperationID: request.ClientOperationID,
		CreatedAt:         nowUTC(),
		UpdatedAt:         nowUTC(),
	}

	if request.StartedAt != nil {
		result.StartedAt = request.StartedAt.UTC()
	}

	created, err := s.repository.Create(
		ctx,
		result,
	)
	if err != nil {
		return nil, err
	}

	/*
		Terminal sessions do not have a web user/attendant.

		We therefore do not create the normal user audit entry
		here. The session itself remains fully tenant/branch/
		terminal scoped.
	*/

	return created, nil
}

type TerminalBillingConfig struct {
	BranchID      string `json:"branch_id"`
	RatePerMinute string `json:"rate_per_minute"`
	MinimumCharge string `json:"minimum_charge"`
	Currency      string `json:"currency"`
}

func (s *Service) GetBillingConfigFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
) (*TerminalBillingConfig, error) {
	terminalID = strings.TrimSpace(terminalID)
	tenantID = strings.TrimSpace(tenantID)
	branchID = strings.TrimSpace(branchID)

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	// Verify that the authenticated terminal really belongs
	// to this tenant and branch.
	if err := s.repository.VerifyTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	); err != nil {
		return nil, err
	}

	rate, minimum, _, _, currency, err :=
		s.repository.GetBillingConfig(
			ctx,
			tenantID,
			branchID,
		)

	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(rate) == "" {
		return nil, errors.New(
			"branch billing rate is not configured",
		)
	}

	if parseMoney(rate) <= 0 {
		return nil, errors.New(
			"branch billing rate must be greater than zero",
		)
	}

	if strings.TrimSpace(currency) == "" {
		currency = "KES"
	}

	return &TerminalBillingConfig{
		BranchID:      branchID,
		RatePerMinute: rate,
		MinimumCharge: minimum,
		Currency:      currency,
	}, nil
}

func (s *Service) GetFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
	sessionID string,
) (*Session, error) {
	terminalID = strings.TrimSpace(terminalID)
	tenantID = strings.TrimSpace(tenantID)
	branchID = strings.TrimSpace(branchID)
	sessionID = strings.TrimSpace(sessionID)

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
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

	session, err := s.repository.GetLatestForID(
		ctx,
		tenantID,
		branchID,
		sessionID,
	)

	if err != nil {
		return nil, err
	}

	if session.TerminalID != terminalID {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func (s *Service) PreviewFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
	sessionID string,
	at time.Time,
) (*BillingPreview, error) {
	session, err := s.GetFromTerminal(
		ctx,
		terminalID,
		tenantID,
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

func (s *Service) EndFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
	sessionID string,
	request TerminalEndSessionRequest,
) (*SessionWithBilling, error) {
	session, err := s.GetFromTerminal(
		ctx,
		terminalID,
		tenantID,
		branchID,
		sessionID,
	)

	if err != nil {
		return nil, err
	}

	if session.Status != "active" {
		return nil, ErrInvalidSessionState
	}

	endedAt := nowUTC()

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

func nowUTC() time.Time {
	return time.Now().UTC()
}

func (s *Service) GetActiveFromTerminal(
	ctx context.Context,
	terminalID string,
	tenantID string,
	branchID string,
) (*Session, error) {
	terminalID = strings.TrimSpace(terminalID)
	tenantID = strings.TrimSpace(tenantID)
	branchID = strings.TrimSpace(branchID)

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if err := s.repository.VerifyTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	); err != nil {
		return nil, err
	}

	return s.repository.GetActiveForTerminal(
		ctx,
		tenantID,
		branchID,
		terminalID,
	)
}
