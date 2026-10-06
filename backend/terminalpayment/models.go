package terminalpayment

import "time"

type CreateCashPaymentRequest struct {
	SessionID         string `json:"session_id" binding:"required"`
	Amount            string `json:"amount" binding:"required"`
	ClientOperationID string `json:"client_operation_id" binding:"required"`
}

type CreateCashPaymentResponse struct {
	PaymentID   string     `json:"payment_id"`
	SaleID      string     `json:"sale_id"`
	SessionID   string     `json:"session_id"`
	Amount      string     `json:"amount"`
	Method      string     `json:"method"`
	Status      string     `json:"status"`
	Currency    string     `json:"currency"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

// ------------------------------------------------------------
// Terminal M-Pesa
// ------------------------------------------------------------

type CreateMpesaPaymentRequest struct {
	SessionID         string `json:"session_id" binding:"required"`
	Amount            string `json:"amount" binding:"required"`
	PhoneNumber       string `json:"phone_number" binding:"required"`
	ClientOperationID string `json:"client_operation_id" binding:"required"`
}

type CreateMpesaPaymentResponse struct {
	PaymentID         string     `json:"payment_id"`
	SaleID            string     `json:"sale_id"`
	SessionID         string     `json:"session_id"`
	Amount            string     `json:"amount"`
	Method            string     `json:"method"`
	Status            string     `json:"status"`
	Currency          string     `json:"currency"`
	PhoneNumber       string     `json:"phone_number"`
	CheckoutRequestID string     `json:"checkout_request_id,omitempty"`
	MerchantRequestID string     `json:"merchant_request_id,omitempty"`
	ConfirmedAt       *time.Time `json:"confirmed_at,omitempty"`
}

type MpesaPaymentStatusResponse struct {
	PaymentID     string     `json:"payment_id"`
	SaleID        string     `json:"sale_id"`
	SessionID     string     `json:"session_id"`
	Amount        string     `json:"amount"`
	Method        string     `json:"method"`
	Status        string     `json:"status"`
	Currency      string     `json:"currency"`
	ReceiptNumber string     `json:"receipt_number,omitempty"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
}
