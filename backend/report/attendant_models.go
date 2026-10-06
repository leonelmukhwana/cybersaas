package report

type AttendantReportSummary struct {
	AttendantID string `json:"attendant_id"`

	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`

	// Revenue
	SessionRevenue string `json:"session_revenue"`
	SalesRevenue   string `json:"sales_revenue"`
	TotalRevenue   string `json:"total_revenue"`

	// Confirmed collections
	ConfirmedCollections string `json:"confirmed_collections"`

	// Activity counts
	SessionCount  int64 `json:"session_count"`
	SalesCount    int64 `json:"sales_count"`
	PaymentsCount int64 `json:"payments_count"`
}

type AttendantPaymentMethodTotal struct {
	Method string `json:"method"`
	Amount string `json:"amount"`
	Count  int64  `json:"count"`
}

type AttendantDailyTotal struct {
	Date string `json:"date"`

	SessionRevenue string `json:"session_revenue"`
	SalesRevenue   string `json:"sales_revenue"`
	TotalRevenue   string `json:"total_revenue"`

	SessionCount int64 `json:"session_count"`
	SalesCount   int64 `json:"sales_count"`
}

type AttendantReportResponse struct {
	Summary        AttendantReportSummary        `json:"summary"`
	PaymentMethods []AttendantPaymentMethodTotal `json:"payment_methods"`
	DailyTotals    []AttendantDailyTotal         `json:"daily_totals"`
}
