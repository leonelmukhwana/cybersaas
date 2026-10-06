package sale

import "time"

type Sale struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	BranchID         string     `json:"branch_id"`
	CustomerID       *string    `json:"customer_id,omitempty"`
	SessionID        *string    `json:"session_id,omitempty"`
	SessionStartedAt *time.Time `json:"session_started_at,omitempty"`
	TerminalID       *string    `json:"terminal_id,omitempty"`
	AttendantID      *string    `json:"attendant_id,omitempty"`

	Status string `json:"status"`

	Subtotal       string  `json:"subtotal"`
	DiscountType   *string `json:"discount_type,omitempty"`
	DiscountValue  string  `json:"discount_value"`
	DiscountAmount string  `json:"discount_amount"`
	TotalAmount    string  `json:"total_amount"`
	Currency       string  `json:"currency"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Items []SaleItem `json:"items,omitempty"`
}

type SaleItem struct {
	ID          string  `json:"id"`
	SaleID      string  `json:"sale_id"`
	ServiceID   *string `json:"service_id,omitempty"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   string  `json:"unit_price"`
	LineTotal   string  `json:"line_total"`
}

type CreateSaleItemRequest struct {
	ServiceID   *string `json:"service_id"`
	Description string  `json:"description" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required"`
	UnitPrice   string  `json:"unit_price" binding:"required"`
}

type CreateSaleRequest struct {
	BranchID          string                  `json:"branch_id" binding:"required"`
	CustomerID        *string                 `json:"customer_id"`
	SessionID         *string                 `json:"session_id"`
	SessionStartedAt  *time.Time              `json:"session_started_at"`
	TerminalID        *string                 `json:"terminal_id"`
	AttendantID       *string                 `json:"attendant_id"`
	DiscountType      *string                 `json:"discount_type"`
	DiscountValue     string                  `json:"discount_value"`
	ClientOperationID string                  `json:"client_operation_id" binding:"required"`
	Items             []CreateSaleItemRequest `json:"items" binding:"required,min=1"`
}

type UpdateSaleRequest struct {
	DiscountType  *string `json:"discount_type"`
	DiscountValue string  `json:"discount_value"`
}

type SaleListResponse struct {
	Sales []Sale `json:"sales"`
	Total int64  `json:"total"`
}

type VoidSaleRequest struct {
	Reason string `json:"reason" binding:"required"`
}
