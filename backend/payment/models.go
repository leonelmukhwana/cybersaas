package payment

import "time"

type Payment struct {
	ID                    string     `json:"id"`
	TenantID              string     `json:"tenant_id"`
	BranchID              string     `json:"branch_id"`
	SaleID                string     `json:"sale_id"`
	Method                string     `json:"method"`
	Status                string     `json:"status"`
	Amount                string     `json:"amount"`
	Phone                 *string    `json:"phone,omitempty"`
	ExternalReference     *string    `json:"external_reference,omitempty"`
	MPesaReceiptNumber    *string    `json:"mpesa_receipt_number,omitempty"`
	ProviderRequestID     *string    `json:"provider_request_id,omitempty"`
	ProviderTransactionID *string    `json:"provider_transaction_id,omitempty"`
	FailureReason         *string    `json:"failure_reason,omitempty"`
	ConfirmedAt           *time.Time `json:"confirmed_at,omitempty"`
	ClientOperationID     *string    `json:"client_operation_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CreatePaymentRequest struct {
	BranchID          string  `json:"branch_id" binding:"required"`
	SaleID            string  `json:"sale_id" binding:"required"`
	Method            string  `json:"method" binding:"required"`
	Amount            string  `json:"amount" binding:"required"`
	Phone             *string `json:"phone"`
	ExternalReference *string `json:"external_reference"`
	ClientOperationID *string `json:"client_operation_id"`
}

type ConfirmPaymentRequest struct {
	ExternalReference     *string `json:"external_reference"`
	MPesaReceiptNumber    *string `json:"mpesa_receipt_number"`
	ProviderRequestID     *string `json:"provider_request_id"`
	ProviderTransactionID *string `json:"provider_transaction_id"`
}

type FailPaymentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RefundPaymentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type PaymentListResponse struct {
	Payments []Payment `json:"payments"`
	Total    int64     `json:"total"`
}
