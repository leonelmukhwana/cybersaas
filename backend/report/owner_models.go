package report

import "time"

type OwnerReportFilters struct {
	BranchID    *string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type OwnerReportSummary struct {
	TenantID    string    `json:"tenant_id"`
	BranchID    *string   `json:"branch_id,omitempty"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Revenue
	SessionRevenue string `json:"session_revenue"`
	SalesRevenue   string `json:"sales_revenue"`
	TotalRevenue   string `json:"total_revenue"`

	// Collections and expenses
	TotalSales       string `json:"total_sales"`
	TotalCollections string `json:"total_collections"`
	TotalExpenses    string `json:"total_expenses"`
	NetAmount        string `json:"net_amount"`

	// Counts
	SessionCount  int64 `json:"session_count"`
	SalesCount    int64 `json:"sales_count"`
	PaymentsCount int64 `json:"payments_count"`
	ExpensesCount int64 `json:"expenses_count"`
}

type OwnerPaymentMethodTotal struct {
	Method string `json:"method"`
	Amount string `json:"amount"`
	Count  int64  `json:"count"`
}

type OwnerBranchTotal struct {
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`

	// Revenue
	SessionRevenue string `json:"session_revenue"`
	SalesRevenue   string `json:"sales_revenue"`
	TotalRevenue   string `json:"total_revenue"`

	// Collections and expenses
	TotalSales       string `json:"total_sales"`
	TotalCollections string `json:"total_collections"`
	TotalExpenses    string `json:"total_expenses"`
	NetAmount        string `json:"net_amount"`

	// Counts
	SessionCount int64 `json:"session_count"`
	SalesCount   int64 `json:"sales_count"`
}

type OwnerDailyTotal struct {
	Date string `json:"date"`

	// Revenue
	SessionRevenue string `json:"session_revenue"`
	SalesRevenue   string `json:"sales_revenue"`
	TotalRevenue   string `json:"total_revenue"`

	// Collections and expenses
	TotalSales       string `json:"total_sales"`
	TotalCollections string `json:"total_collections"`
	TotalExpenses    string `json:"total_expenses"`
	NetAmount        string `json:"net_amount"`

	// Counts
	SessionCount int64 `json:"session_count"`
	SalesCount   int64 `json:"sales_count"`
}

type OwnerMonthlyTotal struct {
	Month string `json:"month"`

	// Revenue
	SessionRevenue string `json:"session_revenue"`
	SalesRevenue   string `json:"sales_revenue"`
	TotalRevenue   string `json:"total_revenue"`

	// Collections and expenses
	TotalSales       string `json:"total_sales"`
	TotalCollections string `json:"total_collections"`
	TotalExpenses    string `json:"total_expenses"`
	NetAmount        string `json:"net_amount"`

	// Counts
	SessionCount int64 `json:"session_count"`
	SalesCount   int64 `json:"sales_count"`
}

type OwnerSalesReport struct {
	Sales []OwnerSaleRow `json:"sales"`
	Total int64          `json:"total"`
}

type OwnerSaleRow struct {
	ID          string  `json:"id"`
	BranchID    string  `json:"branch_id"`
	CustomerID  *string `json:"customer_id,omitempty"`
	SessionID   *string `json:"session_id,omitempty"`
	TerminalID  *string `json:"terminal_id,omitempty"`
	AttendantID *string `json:"attendant_id,omitempty"`

	Status      string    `json:"status"`
	TotalAmount string    `json:"total_amount"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`

	Items []OwnerSaleItem `json:"items,omitempty"`
}

type OwnerSaleItem struct {
	ID          string  `json:"id"`
	ServiceID   *string `json:"service_id,omitempty"`
	Description string  `json:"description"`
	Quantity    string  `json:"quantity"`
	UnitPrice   string  `json:"unit_price"`
	LineTotal   string  `json:"line_total"`
}

type OwnerPaymentReport struct {
	Payments []OwnerPaymentRow `json:"payments"`
	Total    int64             `json:"total"`
}

type OwnerPaymentRow struct {
	ID                 string     `json:"id"`
	BranchID           string     `json:"branch_id"`
	SaleID             string     `json:"sale_id"`
	Method             string     `json:"method"`
	Status             string     `json:"status"`
	Amount             string     `json:"amount"`
	ExternalReference  *string    `json:"external_reference,omitempty"`
	MPesaReceiptNumber *string    `json:"mpesa_receipt_number,omitempty"`
	ConfirmedAt        *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type OwnerExpenseReport struct {
	Expenses []OwnerExpenseRow `json:"expenses"`
	Total    int64             `json:"total"`
}

type OwnerExpenseRow struct {
	ID            string    `json:"id"`
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
}

type OwnerReportResponse struct {
	Summary        OwnerReportSummary        `json:"summary"`
	PaymentMethods []OwnerPaymentMethodTotal `json:"payment_methods"`
	Branches       []OwnerBranchTotal        `json:"branches"`

	// Time-based report data
	DailyTotals   []OwnerDailyTotal   `json:"daily_totals,omitempty"`
	MonthlyTotals []OwnerMonthlyTotal `json:"monthly_totals,omitempty"`
}
