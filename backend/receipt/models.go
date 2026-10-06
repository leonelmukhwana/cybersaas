package receipt

import "time"

type Receipt struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	BranchID      string     `json:"branch_id"`
	SaleID        string     `json:"sale_id"`
	PaymentID     *string    `json:"payment_id,omitempty"`
	ReceiptNumber string     `json:"receipt_number"`
	ReceiptType   string     `json:"receipt_type"`
	IssuedAt      time.Time  `json:"issued_at"`
	PrintedAt     *time.Time `json:"printed_at,omitempty"`
	ReprintCount  int        `json:"reprint_count"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ReceiptReprint struct {
	ID         string    `json:"id"`
	ReceiptID  string    `json:"receipt_id"`
	UserID     *string   `json:"user_id,omitempty"`
	TerminalID *string   `json:"terminal_id,omitempty"`
	Reason     *string   `json:"reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type ReceiptWithSale struct {
	Receipt

	SaleID        string  `json:"sale_id"`
	PaymentID     *string `json:"payment_id,omitempty"`
	SaleTotal     string  `json:"sale_total"`
	PaymentAmount string  `json:"payment_amount"`
	PaymentMethod string  `json:"payment_method"`
	PaymentStatus string  `json:"payment_status"`
}

type ReceiptListResponse struct {
	Receipts []Receipt `json:"receipts"`
	Total    int64     `json:"total"`
}

type ReprintReceiptRequest struct {
	Reason     string  `json:"reason"`
	TerminalID *string `json:"terminal_id"`
}
