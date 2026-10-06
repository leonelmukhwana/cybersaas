package payment

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidPaymentID = errors.New("invalid payment id")
	ErrInvalidSaleID    = errors.New("invalid sale id")
	ErrInvalidBranchID  = errors.New("invalid branch id")
	ErrInvalidAmount    = errors.New("amount must be a valid non-negative decimal")
	ErrInvalidMethod    = errors.New("payment method must be cash, mpesa, or other")
	ErrInvalidStatus    = errors.New("invalid payment status")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	request CreatePaymentRequest,
) (*Payment, error) {
	if _, err := uuid.Parse(request.BranchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(request.SaleID); err != nil {
		return nil, ErrInvalidSaleID
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

	if err := s.repository.VerifySaleBelongsToBranch(
		ctx,
		tenantID,
		request.BranchID,
		request.SaleID,
	); err != nil {
		return nil, err
	}

	request.Method = strings.ToLower(strings.TrimSpace(request.Method))

	switch request.Method {
	case "cash", "mpesa", "other":
	default:
		return nil, ErrInvalidMethod
	}

	amount, err := parseMoney(request.Amount)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	request.Amount = formatMoney(amount)

	if request.Method == "mpesa" {
		if request.Phone == nil ||
			strings.TrimSpace(*request.Phone) == "" {
			return nil, errors.New(
				"phone is required for mpesa payments",
			)
		}
	}

	if request.ClientOperationID != nil {
		if _, err := uuid.Parse(*request.ClientOperationID); err != nil {
			return nil, errors.New(
				"invalid client_operation_id",
			)
		}
	}

	payment, err := s.repository.Create(
		ctx,
		tenantID,
		request,
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
		"payment_created",
		payment.ID,
		"",
	); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	paymentID string,
) (*Payment, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(paymentID); err != nil {
		return nil, ErrInvalidPaymentID
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

	return s.repository.Get(
		ctx,
		tenantID,
		branchID,
		paymentID,
	)
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
) (*PaymentListResponse, error) {
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

	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "":
	case "pending":
	case "confirmed":
	case "failed":
	case "cancelled":
	case "refunded":
	default:
		return nil, ErrInvalidStatus
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

	payments, err := s.repository.List(
		ctx,
		tenantID,
		branchID,
		status,
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
	)
	if err != nil {
		return nil, err
	}

	return &PaymentListResponse{
		Payments: payments,
		Total:    total,
	}, nil
}

func (s *Service) Confirm(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	paymentID string,
	request ConfirmPaymentRequest,
) (*Payment, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(paymentID); err != nil {
		return nil, ErrInvalidPaymentID
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

	payment, err := s.repository.Confirm(
		ctx,
		tenantID,
		branchID,
		paymentID,
		request,
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
		"payment_confirmed",
		payment.ID,
		"",
	); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) Fail(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	paymentID string,
	reason string,
) error {
	if _, err := uuid.Parse(branchID); err != nil {
		return ErrInvalidBranchID
	}

	if _, err := uuid.Parse(paymentID); err != nil {
		return ErrInvalidPaymentID
	}

	reason = strings.TrimSpace(reason)

	if reason == "" {
		return errors.New("failure reason is required")
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		branchID,
	); err != nil {
		return err
	}

	if userRole == "attendant" {
		if err := s.repository.VerifyAttendantBranchAccess(
			ctx,
			userID,
			tenantID,
			branchID,
		); err != nil {
			return err
		}
	}

	if err := s.repository.Fail(
		ctx,
		tenantID,
		branchID,
		paymentID,
		reason,
	); err != nil {
		return err
	}

	return s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"payment_failed",
		paymentID,
		reason,
	)
}

func (s *Service) Refund(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	paymentID string,
	reason string,
) error {
	if _, err := uuid.Parse(branchID); err != nil {
		return ErrInvalidBranchID
	}

	if _, err := uuid.Parse(paymentID); err != nil {
		return ErrInvalidPaymentID
	}

	if userRole != "owner" {
		return errors.New("only owners can refund payments")
	}

	reason = strings.TrimSpace(reason)

	if reason == "" {
		return errors.New("refund reason is required")
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		branchID,
	); err != nil {
		return err
	}

	if err := s.repository.Refund(
		ctx,
		tenantID,
		branchID,
		paymentID,
		reason,
	); err != nil {
		return err
	}

	return s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"payment_refunded",
		paymentID,
		reason,
	)
}

var moneyPattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

func parseMoney(value string) (*big.Rat, error) {
	value = strings.TrimSpace(value)

	if !moneyPattern.MatchString(value) {
		return nil, ErrInvalidAmount
	}

	rat := new(big.Rat)

	if _, ok := rat.SetString(value); !ok {
		return nil, ErrInvalidAmount
	}

	if rat.Sign() < 0 {
		return nil, ErrInvalidAmount
	}

	return rat, nil
}

func formatMoney(value *big.Rat) string {
	cents := new(big.Rat).Mul(
		value,
		big.NewRat(100, 1),
	)

	numerator := new(big.Int).Quo(
		cents.Num(),
		cents.Denom(),
	)

	raw := numerator.Int64()

	return formatCents(raw)
}

func formatCents(value int64) string {
	whole := value / 100
	cents := value % 100

	if cents < 0 {
		cents = -cents
	}

	return strings.TrimSpace(
		new(big.Int).SetInt64(whole).String() +
			"." +
			twoDigits(cents),
	)
}

func twoDigits(value int64) string {
	if value < 10 {
		return "0" + new(big.Int).SetInt64(value).String()
	}

	return new(big.Int).SetInt64(value).String()
}
