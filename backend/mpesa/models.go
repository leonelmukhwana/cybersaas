package mpesa

import "time"

type Configuration struct {
	ID                       string    `json:"id"`
	TenantID                 string    `json:"tenant_id,omitempty"`
	BranchID                 string    `json:"branch_id,omitempty"`
	Provider                 string    `json:"provider"`
	Environment              string    `json:"environment"`
	BusinessShortCode        *string   `json:"business_short_code,omitempty"`
	TillNumber               *string   `json:"till_number,omitempty"`
	PaybillNumber            *string   `json:"paybill_number,omitempty"`
	AccountReference         *string   `json:"account_reference,omitempty"`
	CallbackURL              *string   `json:"callback_url,omitempty"`
	Active                   bool      `json:"active"`
	ConsumerKeyConfigured    bool      `json:"consumer_key_configured"`
	ConsumerSecretConfigured bool      `json:"consumer_secret_configured"`
	PasskeyConfigured        bool      `json:"passkey_configured"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type SaveConfigurationRequest struct {
	BranchID          string  `json:"branch_id,omitempty"`
	Provider          string  `json:"provider"`
	Environment       string  `json:"environment"`
	BusinessShortCode *string `json:"business_short_code,omitempty"`
	TillNumber        *string `json:"till_number,omitempty"`
	PaybillNumber     *string `json:"paybill_number,omitempty"`

	ConsumerKey    *string `json:"consumer_key,omitempty"`
	ConsumerSecret *string `json:"consumer_secret,omitempty"`
	Passkey        *string `json:"passkey,omitempty"`

	AccountReference *string `json:"account_reference,omitempty"`
	CallbackURL      *string `json:"callback_url,omitempty"`
	Active           *bool   `json:"active,omitempty"`
}

type PlatformSaveConfigurationRequest struct {
	Provider          string  `json:"provider"`
	Environment       string  `json:"environment"`
	BusinessShortCode *string `json:"business_short_code,omitempty"`
	TillNumber        *string `json:"till_number,omitempty"`
	PaybillNumber     *string `json:"paybill_number,omitempty"`

	ConsumerKey    *string `json:"consumer_key,omitempty"`
	ConsumerSecret *string `json:"consumer_secret,omitempty"`
	Passkey        *string `json:"passkey,omitempty"`

	AccountReference *string `json:"account_reference,omitempty"`
	CallbackURL      *string `json:"callback_url,omitempty"`
	Active           *bool   `json:"active,omitempty"`
}

type STKRequest struct {
	ID                     string  `json:"id"`
	TenantID               string  `json:"tenant_id"`
	BranchID               string  `json:"branch_id"`
	SaleID                 string  `json:"sale_id"`
	PaymentID              *string `json:"payment_id,omitempty"`
	PhoneNumber            string  `json:"phone_number"`
	Amount                 string  `json:"amount"`
	AccountReference       string  `json:"account_reference"`
	TransactionDescription string  `json:"transaction_description"`

	MerchantRequestID   *string `json:"merchant_request_id,omitempty"`
	CheckoutRequestID   *string `json:"checkout_request_id,omitempty"`
	ResponseCode        *string `json:"response_code,omitempty"`
	ResponseDescription *string `json:"response_description,omitempty"`
	CustomerMessage     *string `json:"customer_message,omitempty"`

	Status            string  `json:"status"`
	ResultCode        *string `json:"result_code,omitempty"`
	ResultDescription *string `json:"result_description,omitempty"`

	MPesaReceiptNumber *string    `json:"mpesa_receipt_number,omitempty"`
	TransactionDate    *time.Time `json:"transaction_date,omitempty"`
	CallbackReceivedAt *time.Time `json:"callback_received_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InitiateSTKRequest struct {
	BranchID               string  `json:"branch_id" binding:"required"`
	SaleID                 string  `json:"sale_id" binding:"required"`
	PaymentID              *string `json:"payment_id,omitempty"`
	PhoneNumber            string  `json:"phone_number" binding:"required"`
	Amount                 string  `json:"amount" binding:"required"`
	AccountReference       string  `json:"account_reference" binding:"required"`
	TransactionDescription string  `json:"transaction_description" binding:"required"`
}

// STKCallbackRequest represents the complete payload sent by Safaricom.
type STKCallbackRequest struct {
	Body STKCallbackBody `json:"Body"`
}

type STKCallbackBody struct {
	StkCallback STKCallback `json:"stkCallback"`
}

type STKCallback struct {
	MerchantRequestID string            `json:"MerchantRequestID"`
	CheckoutRequestID string            `json:"CheckoutRequestID"`
	ResultCode        int64             `json:"ResultCode"`
	ResultDesc        string            `json:"ResultDesc"`
	CallbackMetadata  *CallbackMetadata `json:"CallbackMetadata,omitempty"`
}

// CallbackMetadata contains the values returned after a successful
// M-Pesa STK transaction.
type CallbackMetadata struct {
	Item []CallbackMetadataItem `json:"Item"`
}

type CallbackMetadataItem struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value,omitempty"`
}

// STKCallbackResult is the normalized callback information that can be
// used internally by the service layer.
type STKCallbackResult struct {
	CheckoutRequestID  string     `json:"checkout_request_id"`
	ResultCode         int64      `json:"result_code"`
	ResultDescription  string     `json:"result_description"`
	MPesaReceiptNumber *string    `json:"mpesa_receipt_number,omitempty"`
	TransactionDate    *time.Time `json:"transaction_date,omitempty"`
	PhoneNumber        *string    `json:"phone_number,omitempty"`
	Amount             *string    `json:"amount,omitempty"`
}

type SubscriptionSTKRequest struct {
	ID                     string `json:"id"`
	TenantID               string `json:"tenant_id"`
	SubscriptionID         string `json:"subscription_id"`
	SubscriptionPaymentID  string `json:"subscription_payment_id"`
	PhoneNumber            string `json:"phone_number"`
	Amount                 string `json:"amount"`
	AccountReference       string `json:"account_reference"`
	TransactionDescription string `json:"transaction_description"`

	MerchantRequestID   *string `json:"merchant_request_id,omitempty"`
	CheckoutRequestID   *string `json:"checkout_request_id,omitempty"`
	ResponseCode        *string `json:"response_code,omitempty"`
	ResponseDescription *string `json:"response_description,omitempty"`
	CustomerMessage     *string `json:"customer_message,omitempty"`

	Status string `json:"status"`

	ResultCode         *string    `json:"result_code,omitempty"`
	ResultDescription  *string    `json:"result_description,omitempty"`
	MPesaReceiptNumber *string    `json:"mpesa_receipt_number,omitempty"`
	TransactionDate    *time.Time `json:"transaction_date,omitempty"`
	CallbackReceivedAt *time.Time `json:"callback_received_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
