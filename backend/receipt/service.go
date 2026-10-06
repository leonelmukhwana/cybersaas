package receipt

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidReceiptID = errors.New("invalid receipt id")
	ErrInvalidBranchID  = errors.New("invalid branch id")
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
	branchID string,
	saleID string,
	paymentID *string,
) (*Receipt, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(saleID); err != nil {
		return nil, errors.New("invalid sale id")
	}

	if paymentID != nil {
		if _, err := uuid.Parse(*paymentID); err != nil {
			return nil, errors.New("invalid payment id")
		}
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

	receipt, err := s.repository.Create(
		ctx,
		tenantID,
		branchID,
		saleID,
		paymentID,
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
		"receipt_created",
		receipt.ID,
		"",
	); err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s *Service) Get(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	receiptID string,
) (*ReceiptWithSale, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(receiptID); err != nil {
		return nil, ErrInvalidReceiptID
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
		receiptID,
	)
}

func (s *Service) List(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	limit int,
	offset int,
) (*ReceiptListResponse, error) {
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

	receipts, err := s.repository.List(
		ctx,
		tenantID,
		branchID,
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
	)
	if err != nil {
		return nil, err
	}

	return &ReceiptListResponse{
		Receipts: receipts,
		Total:    total,
	}, nil
}

func (s *Service) MarkPrinted(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	receiptID string,
) (*Receipt, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(receiptID); err != nil {
		return nil, ErrInvalidReceiptID
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

	receipt, err := s.repository.MarkPrinted(
		ctx,
		tenantID,
		branchID,
		receiptID,
	)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s *Service) Reprint(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
	receiptID string,
	request ReprintReceiptRequest,
) (*Receipt, error) {
	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(receiptID); err != nil {
		return nil, ErrInvalidReceiptID
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

	reason := strings.TrimSpace(request.Reason)

	if reason == "" {
		return nil, errors.New("reprint reason is required")
	}

	if len(reason) > 500 {
		return nil, errors.New(
			"reprint reason must not exceed 500 characters",
		)
	}

	if request.TerminalID != nil {
		if _, err := uuid.Parse(*request.TerminalID); err != nil {
			return nil, errors.New("invalid terminal id")
		}
	}

	receipt, err := s.repository.Reprint(
		ctx,
		tenantID,
		branchID,
		receiptID,
		userID,
		request.TerminalID,
		&reason,
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
		"receipt_reprinted",
		receipt.ID,
		reason,
	); err != nil {
		return nil, err
	}

	return receipt, nil
}
