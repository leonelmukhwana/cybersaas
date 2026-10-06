package expenses

import "time"

type Expense struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	BranchID      string    `json:"branch_id"`
	RecordedBy    string    `json:"recorded_by"`
	Category      string    `json:"category"`
	Description   *string   `json:"description,omitempty"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"payment_method"`
	Reference     *string   `json:"reference,omitempty"`
	ExpenseDate   time.Time `json:"expense_date"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateExpenseRequest struct {
	BranchID      string  `json:"branch_id" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	Description   *string `json:"description"`
	Amount        string  `json:"amount" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	Reference     *string `json:"reference"`

	// Frontend sends YYYY-MM-DD.
	// We keep this as a string at the API boundary
	// and parse it in the service.
	ExpenseDate string `json:"expense_date" binding:"required"`
}

type ExpenseListResponse struct {
	Expenses []Expense `json:"expenses"`
	Total    int64     `json:"total"`
}
