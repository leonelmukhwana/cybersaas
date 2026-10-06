package expenses

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTenantID    = errors.New("invalid tenant id")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidBranchID    = errors.New("invalid branch id")
	ErrInvalidExpenseID   = errors.New("invalid expense id")
	ErrInvalidCategory    = errors.New("expense category is required")
	ErrInvalidAmount      = errors.New("invalid expense amount")
	ErrInvalidPayment     = errors.New("invalid payment method")
	ErrInvalidExpenseDate = errors.New("invalid expense date")
	ErrBranchAccess       = errors.New("branch access denied")
)

type BranchAccessVerifier interface {
	VerifyAttendantBranchAccess(
		ctx context.Context,
		tenantID string,
		userID string,
		branchID string,
	) error

	VerifyBranchBelongsToTenant(
		ctx context.Context,
		tenantID string,
		branchID string,
	) error
}

type Service struct {
	repository *Repository
	branches   BranchAccessVerifier
}

func NewService(
	repository *Repository,
	branches BranchAccessVerifier,
) *Service {
	return &Service{
		repository: repository,
		branches:   branches,
	}
}

func (s *Service) Create(
	ctx context.Context,
	tenantID string,
	userID string,
	request CreateExpenseRequest,
) (*Expense, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, ErrInvalidTenantID
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	if _, err := uuid.Parse(request.BranchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	category := strings.TrimSpace(request.Category)

	if category == "" {
		return nil, ErrInvalidCategory
	}

	amount := strings.TrimSpace(request.Amount)

	if amount == "" {
		return nil, ErrInvalidAmount
	}

	paymentMethod := strings.ToLower(
		strings.TrimSpace(request.PaymentMethod),
	)

	switch paymentMethod {
	case "cash", "mpesa", "bank", "card", "other":
	default:
		return nil, ErrInvalidPayment
	}

	expenseDateString := strings.TrimSpace(request.ExpenseDate)

	if expenseDateString == "" {
		return nil, ErrInvalidExpenseDate
	}

	expenseDate, err := time.Parse(
		"2006-01-02",
		expenseDateString,
	)
	if err != nil {
		return nil, ErrInvalidExpenseDate
	}

	if s.branches != nil {
		if err := s.branches.VerifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			request.BranchID,
		); err != nil {
			return nil, ErrBranchAccess
		}
	}

	description := request.Description

	if description != nil {
		value := strings.TrimSpace(*description)

		if value == "" {
			description = nil
		} else {
			description = &value
		}
	}

	reference := request.Reference

	if reference != nil {
		value := strings.TrimSpace(*reference)

		if value == "" {
			reference = nil
		} else {
			reference = &value
		}
	}

	return s.repository.Create(
		ctx,
		CreateExpenseParams{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			BranchID:      request.BranchID,
			RecordedBy:    userID,
			Category:      category,
			Description:   description,
			Amount:        amount,
			Currency:      "KES",
			PaymentMethod: paymentMethod,
			Reference:     reference,
			ExpenseDate:   expenseDate,
			Status:        "recorded",
		},
	)
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	id string,
) (*Expense, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, ErrInvalidTenantID
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidExpenseID
	}

	return s.repository.Get(
		ctx,
		tenantID,
		id,
	)
}

func (s *Service) ListForOwner(
	ctx context.Context,
	tenantID string,
	branchID string,
) ([]Expense, int64, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, 0, ErrInvalidTenantID
	}

	var branchFilter *string

	if branchID != "" {
		if _, err := uuid.Parse(branchID); err != nil {
			return nil, 0, ErrInvalidBranchID
		}

		if s.branches != nil {
			if err := s.branches.VerifyBranchBelongsToTenant(
				ctx,
				tenantID,
				branchID,
			); err != nil {
				return nil, 0, ErrBranchAccess
			}
		}

		branchFilter = &branchID
	}

	return s.repository.ListForTenant(
		ctx,
		tenantID,
		branchFilter,
	)
}

func (s *Service) ListForAttendant(
	ctx context.Context,
	tenantID string,
	userID string,
	branchID string,
) ([]Expense, int64, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, 0, ErrInvalidTenantID
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, 0, ErrInvalidUserID
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, 0, ErrInvalidBranchID
	}

	if s.branches != nil {
		if err := s.branches.VerifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			branchID,
		); err != nil {
			return nil, 0, ErrBranchAccess
		}
	}

	return s.repository.ListForBranch(
		ctx,
		tenantID,
		branchID,
	)
}

func (s *Service) ValidateExpenseDate(
	date time.Time,
) error {
	if date.IsZero() {
		return ErrInvalidExpenseDate
	}

	return nil
}
