package sale

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidSaleID       = errors.New("invalid sale id")
	ErrInvalidBranchID     = errors.New("invalid branch id")
	ErrInvalidCustomerID   = errors.New("invalid customer id")
	ErrInvalidSessionID    = errors.New("invalid session id")
	ErrInvalidTerminalID   = errors.New("invalid terminal id")
	ErrInvalidAttendantID  = errors.New("invalid attendant id")
	ErrInvalidSaleItem     = errors.New("invalid sale item")
	ErrInvalidQuantity     = errors.New("quantity must be greater than zero")
	ErrInvalidPrice        = errors.New("price must be a valid non-negative decimal")
	ErrInvalidDiscount     = errors.New("invalid discount")
	ErrInvalidDiscountType = errors.New("discount type must be fixed or percentage")
	ErrNoSaleItems         = errors.New("at least one sale item is required")
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
	request CreateSaleRequest,
) (*Sale, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(request.BranchID); err != nil {
		return nil, ErrInvalidBranchID
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

	if len(request.Items) == 0 {
		return nil, ErrNoSaleItems
	}

	if request.CustomerID != nil {
		if _, err := uuid.Parse(*request.CustomerID); err != nil {
			return nil, ErrInvalidCustomerID
		}
	}

	if request.SessionID != nil {
		if _, err := uuid.Parse(*request.SessionID); err != nil {
			return nil, ErrInvalidSessionID
		}

		if request.SessionStartedAt == nil {
			return nil, errors.New(
				"session_started_at is required when session_id is provided",
			)
		}
	}

	if request.TerminalID != nil {
		if _, err := uuid.Parse(*request.TerminalID); err != nil {
			return nil, ErrInvalidTerminalID
		}
	}

	if request.AttendantID != nil {
		if _, err := uuid.Parse(*request.AttendantID); err != nil {
			return nil, ErrInvalidAttendantID
		}
	}

	subtotal := new(big.Rat)

	for i := range request.Items {
		item := &request.Items[i]

		item.Description = strings.TrimSpace(item.Description)

		if item.Description == "" {
			return nil, ErrInvalidSaleItem
		}

		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		price, err := parseMoney(item.UnitPrice)
		if err != nil {
			return nil, ErrInvalidPrice
		}

		lineTotal := new(big.Rat).Mul(
			big.NewRat(int64(item.Quantity), 1),
			price,
		)

		subtotal.Add(subtotal, lineTotal)

		item.UnitPrice = formatMoney(price)
	}

	discountType, discountValue, err := normalizeDiscount(
		request.DiscountType,
		request.DiscountValue,
	)
	if err != nil {
		return nil, err
	}

	discountAmount := new(big.Rat)

	if discountType != nil {
		value, _ := parseMoney(discountValue)

		switch *discountType {
		case "fixed":
			discountAmount.Set(value)

			if discountAmount.Cmp(subtotal) > 0 {
				discountAmount.Set(subtotal)
			}

		case "percentage":
			discountAmount.Mul(
				subtotal,
				value,
			)
			discountAmount.Quo(
				discountAmount,
				big.NewRat(100, 1),
			)

			if discountAmount.Cmp(subtotal) > 0 {
				discountAmount.Set(subtotal)
			}
		}
	}

	total := new(big.Rat).Sub(
		subtotal,
		discountAmount,
	)

	currency := "KES"

	sale, err := s.repository.Create(
		ctx,
		tenantID,
		request,
		formatMoney(subtotal),
		formatMoney(discountAmount),
		formatMoney(total),
		currency,
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
		"sale_created",
		sale.ID,
		"",
	); err != nil {
		return nil, err
	}

	return sale, nil
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	saleID string,
) (*Sale, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(saleID); err != nil {
		return nil, ErrInvalidSaleID
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
		saleID,
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
) (*SaleListResponse, error) {
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

	status = strings.TrimSpace(strings.ToLower(status))

	switch status {
	case "":
	case "completed":
	case "voided":
	case "refunded":
	default:
		return nil, errors.New("invalid sale status")
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

	sales, err := s.repository.List(
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

	return &SaleListResponse{
		Sales: sales,
		Total: total,
	}, nil
}

func (s *Service) Void(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	saleID string,
	reason string,
) error {
	if _, err := uuid.Parse(branchID); err != nil {
		return ErrInvalidBranchID
	}

	if _, err := uuid.Parse(saleID); err != nil {
		return ErrInvalidSaleID
	}

	reason = strings.TrimSpace(reason)

	if reason == "" {
		return errors.New("void reason is required")
	}

	if len(reason) > 500 {
		return errors.New("void reason must not exceed 500 characters")
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		branchID,
	); err != nil {
		return err
	}

	if userRole == "attendant" {
		return errors.New("attendants cannot void sales")
	}

	if err := s.repository.Void(
		ctx,
		tenantID,
		branchID,
		saleID,
	); err != nil {
		return err
	}

	return s.repository.CreateAuditLog(
		ctx,
		tenantID,
		branchID,
		userID,
		userRole,
		"sale_voided",
		saleID,
		reason,
	)
}

func normalizeDiscount(
	discountType *string,
	discountValue string,
) (*string, string, error) {
	value := strings.TrimSpace(discountValue)

	if discountType == nil || strings.TrimSpace(*discountType) == "" {
		if value != "" && value != "0" && value != "0.00" {
			return nil, "", ErrInvalidDiscount
		}

		return nil, "0.00", nil
	}

	normalizedType := strings.ToLower(
		strings.TrimSpace(*discountType),
	)

	if normalizedType != "fixed" &&
		normalizedType != "percentage" {
		return nil, "", ErrInvalidDiscountType
	}

	if value == "" {
		value = "0"
	}

	parsed, err := parseMoney(value)
	if err != nil {
		return nil, "", ErrInvalidDiscount
	}

	if parsed.Sign() < 0 {
		return nil, "", ErrInvalidDiscount
	}

	if normalizedType == "percentage" &&
		parsed.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, "", ErrInvalidDiscount
	}

	return &normalizedType, formatMoney(parsed), nil
}

var moneyPattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

func parseMoney(value string) (*big.Rat, error) {
	value = strings.TrimSpace(value)

	if !moneyPattern.MatchString(value) {
		return nil, ErrInvalidPrice
	}

	rat := new(big.Rat)

	if _, ok := rat.SetString(value); !ok {
		return nil, ErrInvalidPrice
	}

	if rat.Sign() < 0 {
		return nil, ErrInvalidPrice
	}

	return rat, nil
}

func formatMoney(value *big.Rat) string {
	value = new(big.Rat).Set(value)

	value.Mul(value, big.NewRat(100, 1))

	quotient := new(big.Int).Quo(
		value.Num(),
		value.Denom(),
	)

	cents := quotient.Int64()

	whole := cents / 100
	fraction := cents % 100

	return formatMoneyParts(whole, fraction)
}

func formatMoneyParts(whole, fraction int64) string {
	return strings.TrimSpace(
		formatInt64(whole) + "." + twoDigits(fraction),
	)
}

func formatInt64(value int64) string {
	if value == 0 {
		return "0"
	}

	return new(big.Int).SetInt64(value).String()
}

func twoDigits(value int64) string {
	if value < 10 {
		return "0" + formatInt64(value)
	}

	return formatInt64(value)
}
