package report

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type OwnerReportService struct {
	repository *OwnerRepository
}

func NewOwnerReportService(
	repository *OwnerRepository,
) *OwnerReportService {
	return &OwnerReportService{
		repository: repository,
	}
}

func (s *OwnerReportService) GetSummary(
	ctx context.Context,
	tenantID string,
	branchID *string,
	periodStart time.Time,
	periodEnd time.Time,
) (OwnerReportResponse, error) {
	if tenantID == "" {
		return OwnerReportResponse{}, fmt.Errorf("tenant ID is required")
	}

	if periodStart.IsZero() {
		return OwnerReportResponse{}, fmt.Errorf("period start is required")
	}

	if periodEnd.IsZero() {
		return OwnerReportResponse{}, fmt.Errorf("period end is required")
	}

	if !periodEnd.After(periodStart) {
		return OwnerReportResponse{}, fmt.Errorf(
			"period end must be after period start",
		)
	}

	start := periodStart.UTC().Format(time.RFC3339)
	end := periodEnd.UTC().Format(time.RFC3339)

	// ------------------------------------------------------------
	// SUMMARY
	// ------------------------------------------------------------
	summary, err := s.repository.GetSummary(
		ctx,
		tenantID,
		branchID,
		start,
		end,
	)
	if err != nil {
		return OwnerReportResponse{}, fmt.Errorf(
			"get owner report summary: %w",
			err,
		)
	}

	// ------------------------------------------------------------
	// PAYMENT METHODS
	// ------------------------------------------------------------
	paymentMethods, err := s.repository.GetPaymentMethods(
		ctx,
		tenantID,
		branchID,
		start,
		end,
	)
	if err != nil {
		return OwnerReportResponse{}, fmt.Errorf(
			"get owner report payment methods: %w",
			err,
		)
	}

	// ------------------------------------------------------------
	// BRANCH TOTALS
	// ------------------------------------------------------------
	branches, err := s.repository.GetBranchTotals(
		ctx,
		tenantID,
		branchID,
		start,
		end,
	)
	if err != nil {
		return OwnerReportResponse{}, fmt.Errorf(
			"get owner report branch totals: %w",
			err,
		)
	}

	// ------------------------------------------------------------
	// NET REVENUE
	//
	// Total Revenue = Session Revenue + Sales Revenue
	// Net Revenue   = Total Revenue - Expenses
	//
	// Collections are kept separate because they represent
	// confirmed payment transactions and should not be added
	// to revenue again.
	// ------------------------------------------------------------
	summary.NetAmount = calculateMoneyDifference(
		summary.TotalRevenue,
		summary.TotalExpenses,
	)

	for i := range branches {
		branches[i].NetAmount = calculateMoneyDifference(
			branches[i].TotalRevenue,
			branches[i].TotalExpenses,
		)
	}

	return OwnerReportResponse{
		Summary:        summary,
		PaymentMethods: paymentMethods,
		Branches:       branches,
	}, nil
}

// DownloadReport creates the downloadable CSV version of the same
// owner report shown by GetSummary.
func (s *OwnerReportService) DownloadReport(
	ctx context.Context,
	tenantID string,
	branchID *string,
	periodStart time.Time,
	periodEnd time.Time,
) ([]byte, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}

	if periodStart.IsZero() {
		return nil, fmt.Errorf("period start is required")
	}

	if periodEnd.IsZero() {
		return nil, fmt.Errorf("period end is required")
	}

	if !periodEnd.After(periodStart) {
		return nil, fmt.Errorf(
			"period end must be after period start",
		)
	}

	start := periodStart.UTC().Format(time.RFC3339)
	end := periodEnd.UTC().Format(time.RFC3339)

	reportData, err := s.repository.GetDownloadReport(
		ctx,
		tenantID,
		branchID,
		start,
		end,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get owner report download data: %w",
			err,
		)
	}

	// ------------------------------------------------------------
	// NET REVENUE
	// ------------------------------------------------------------
	reportData.Summary.NetAmount = calculateMoneyDifference(
		reportData.Summary.TotalRevenue,
		reportData.Summary.TotalExpenses,
	)

	for i := range reportData.Branches {
		reportData.Branches[i].NetAmount = calculateMoneyDifference(
			reportData.Branches[i].TotalRevenue,
			reportData.Branches[i].TotalExpenses,
		)
	}

	return buildOwnerCSV(reportData)
}

func buildOwnerCSV(
	report OwnerReportResponse,
) ([]byte, error) {
	var buffer bytes.Buffer

	writer := csv.NewWriter(&buffer)

	// ------------------------------------------------------------
	// REPORT TITLE
	// ------------------------------------------------------------
	if err := writer.Write([]string{
		"CYBERSAAS - CYBER OWNER REPORT",
	}); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// PERIOD
	// ------------------------------------------------------------
	if err := writer.Write([]string{
		"Period Start",
		report.Summary.PeriodStart.Format(time.RFC3339),
	}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{
		"Period End",
		report.Summary.PeriodEnd.Format(time.RFC3339),
	}); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// BRANCH
	// ------------------------------------------------------------
	branchName := "All Branches"

	if report.Summary.BranchID != nil {
		branchName = *report.Summary.BranchID
	}

	if err := writer.Write([]string{
		"Branch",
		branchName,
	}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// SUMMARY
	// ------------------------------------------------------------
	if err := writer.Write([]string{
		"SUMMARY",
	}); err != nil {
		return nil, err
	}

	summaryRows := [][]string{
		{"Session Revenue", report.Summary.SessionRevenue},
		{"Sales Revenue", report.Summary.SalesRevenue},
		{"Total Revenue", report.Summary.TotalRevenue},
		{"Total Collections", report.Summary.TotalCollections},
		{"Total Expenses", report.Summary.TotalExpenses},
		{"Net Revenue", report.Summary.NetAmount},

		{"Session Count", strconv.FormatInt(
			report.Summary.SessionCount,
			10,
		)},
		{"Sales Count", strconv.FormatInt(
			report.Summary.SalesCount,
			10,
		)},
		{"Payments Count", strconv.FormatInt(
			report.Summary.PaymentsCount,
			10,
		)},
		{"Expenses Count", strconv.FormatInt(
			report.Summary.ExpensesCount,
			10,
		)},
	}

	for _, row := range summaryRows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// PAYMENT METHODS
	// ------------------------------------------------------------
	if err := writer.Write([]string{
		"PAYMENT METHODS",
	}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{
		"Method",
		"Amount",
		"Count",
	}); err != nil {
		return nil, err
	}

	for _, item := range report.PaymentMethods {
		if err := writer.Write([]string{
			item.Method,
			item.Amount,
			strconv.FormatInt(item.Count, 10),
		}); err != nil {
			return nil, err
		}
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// BRANCH TOTALS
	// ------------------------------------------------------------
	if err := writer.Write([]string{
		"BRANCH TOTALS",
	}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{
		"Branch",
		"Session Revenue",
		"Sales Revenue",
		"Total Revenue",
		"Collections",
		"Expenses",
		"Net Revenue",
		"Session Count",
		"Sales Count",
	}); err != nil {
		return nil, err
	}

	for _, branch := range report.Branches {
		if err := writer.Write([]string{
			branch.BranchName,
			branch.SessionRevenue,
			branch.SalesRevenue,
			branch.TotalRevenue,
			branch.TotalCollections,
			branch.TotalExpenses,
			branch.NetAmount,
			strconv.FormatInt(branch.SessionCount, 10),
			strconv.FormatInt(branch.SalesCount, 10),
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf(
			"write owner report csv: %w",
			err,
		)
	}

	return buffer.Bytes(), nil
}

// calculateMoneyDifference performs exact decimal subtraction.
//
// This uses shopspring/decimal instead of float64 because these
// values represent financial amounts.
func calculateMoneyDifference(
	first string,
	second string,
) string {
	firstDecimal, err := decimal.NewFromString(first)
	if err != nil {
		return "0.00"
	}

	secondDecimal, err := decimal.NewFromString(second)
	if err != nil {
		return firstDecimal.StringFixed(2)
	}

	return firstDecimal.
		Sub(secondDecimal).
		StringFixed(2)
}
