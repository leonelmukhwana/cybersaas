package terminalpayment

import (
	"context"
	"errors"
	"strings"

	"cybersaas/backend/mpesa"

	"github.com/google/uuid"
)

type Service struct {
	repo         *Repository
	mpesaService *mpesa.Service
}

func NewService(
	repo *Repository,
	mpesaService *mpesa.Service,
) *Service {
	return &Service{
		repo:         repo,
		mpesaService: mpesaService,
	}
}

// CreateCashPayment creates and confirms a cash payment.
//
// IMPORTANT:
// This existing cash flow is intentionally preserved.
func (s *Service) CreateCashPayment(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	req CreateCashPaymentRequest,
) (*CreateCashPaymentResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	terminalID = strings.TrimSpace(terminalID)
	branchID = strings.TrimSpace(branchID)

	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal id is required")
	}

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Amount = strings.TrimSpace(req.Amount)
	req.ClientOperationID = strings.TrimSpace(req.ClientOperationID)

	if req.SessionID == "" {
		return nil, errors.New("session id is required")
	}

	if req.Amount == "" {
		return nil, errors.New("amount is required")
	}

	if req.ClientOperationID == "" {
		return nil, errors.New("client operation id is required")
	}

	if _, err := uuid.Parse(req.SessionID); err != nil {
		return nil, errors.New("invalid session id")
	}

	if _, err := uuid.Parse(req.ClientOperationID); err != nil {
		return nil, errors.New("invalid client operation id")
	}

	return s.repo.CreateCashPayment(
		ctx,
		tenantID,
		terminalID,
		branchID,
		req,
	)
}

// CreateMpesaPayment creates the terminal sale/payment and starts
// the Safaricom M-Pesa STK Push.
//
// The payment remains pending until the Safaricom callback confirms it.
// STK acceptance is NOT treated as payment confirmation.
func (s *Service) CreateMpesaPayment(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	req CreateMpesaPaymentRequest,
) (*CreateMpesaPaymentResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	terminalID = strings.TrimSpace(terminalID)
	branchID = strings.TrimSpace(branchID)

	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal id is required")
	}

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Amount = strings.TrimSpace(req.Amount)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.ClientOperationID = strings.TrimSpace(req.ClientOperationID)

	if req.SessionID == "" {
		return nil, errors.New("session id is required")
	}

	if req.Amount == "" {
		return nil, errors.New("amount is required")
	}

	if req.PhoneNumber == "" {
		return nil, errors.New("phone number is required")
	}

	if req.ClientOperationID == "" {
		return nil, errors.New("client operation id is required")
	}

	if _, err := uuid.Parse(req.SessionID); err != nil {
		return nil, errors.New("invalid session id")
	}

	if _, err := uuid.Parse(req.ClientOperationID); err != nil {
		return nil, errors.New("invalid client operation id")
	}

	if s.mpesaService == nil {
		return nil, errors.New("mpesa service is not configured")
	}

	// ------------------------------------------------------------
	// 1. Create the sale and pending payment.
	// ------------------------------------------------------------

	result, err := s.repo.CreateMpesaPayment(
		ctx,
		tenantID,
		terminalID,
		branchID,
		req,
	)
	if err != nil {
		return nil, err
	}

	// If the same client operation was submitted again and the
	// payment already has a terminal M-Pesa STK request, the
	// repository result can safely be returned as pending.
	//
	// The actual STK request is linked to the existing payment
	// through PaymentID below.

	paymentID := strings.TrimSpace(result.PaymentID)
	saleID := strings.TrimSpace(result.SaleID)

	if paymentID == "" {
		return nil, errors.New("mpesa payment id was not created")
	}

	if saleID == "" {
		return nil, errors.New("mpesa sale id was not created")
	}

	// ------------------------------------------------------------
	// 2. Start the Daraja STK Push.
	// ------------------------------------------------------------

	stkRequest, err := s.mpesaService.InitiateTerminalSTKPush(
		ctx,
		tenantID,
		terminalID,
		branchID,
		mpesa.InitiateSTKRequest{
			BranchID:               branchID,
			SaleID:                 saleID,
			PaymentID:              &paymentID,
			PhoneNumber:            req.PhoneNumber,
			Amount:                 req.Amount,
			AccountReference:       "CYBERCAFE",
			TransactionDescription: "Cyber café computer session",
		},
	)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3. Return the payment as pending/accepted.
	//
	// IMPORTANT:
	// "accepted" here means Safaricom accepted the STK request.
	// It does NOT mean the customer has paid.
	// ------------------------------------------------------------

	result.Status = "pending"

	if stkRequest.CheckoutRequestID != nil {
		result.CheckoutRequestID = *stkRequest.CheckoutRequestID
	}

	if stkRequest.MerchantRequestID != nil {
		result.MerchantRequestID = *stkRequest.MerchantRequestID
	}

	return result, nil
}

func (s *Service) GetMpesaPaymentStatus(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	paymentID string,
) (*MpesaPaymentStatusResponse, error) {
	tenantID = strings.TrimSpace(tenantID)
	terminalID = strings.TrimSpace(terminalID)
	branchID = strings.TrimSpace(branchID)
	paymentID = strings.TrimSpace(paymentID)

	if tenantID == "" {
		return nil, errors.New("tenant id is required")
	}

	if terminalID == "" {
		return nil, errors.New("terminal id is required")
	}

	if branchID == "" {
		return nil, errors.New("branch id is required")
	}

	if paymentID == "" {
		return nil, errors.New("payment id is required")
	}

	if _, err := uuid.Parse(paymentID); err != nil {
		return nil, errors.New("invalid payment id")
	}

	return s.repo.GetMpesaPaymentStatus(
		ctx,
		tenantID,
		terminalID,
		branchID,
		paymentID,
	)
}
