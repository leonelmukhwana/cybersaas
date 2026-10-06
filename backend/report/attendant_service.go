package report

import (
	"context"
	"fmt"
	"time"
)

type AttendantReportService struct {
	repository *AttendantRepository
}

func NewAttendantReportService(
	repository *AttendantRepository,
) *AttendantReportService {
	return &AttendantReportService{
		repository: repository,
	}
}

func (s *AttendantReportService) GetSummary(
	ctx context.Context,
	tenantID string,
	attendantID string,
	periodStart time.Time,
	periodEnd time.Time,
) (AttendantReportResponse, error) {
	if tenantID == "" {
		return AttendantReportResponse{}, fmt.Errorf(
			"tenant ID is required",
		)
	}

	if attendantID == "" {
		return AttendantReportResponse{}, fmt.Errorf(
			"attendant ID is required",
		)
	}

	if periodStart.IsZero() {
		return AttendantReportResponse{}, fmt.Errorf(
			"period start is required",
		)
	}

	if periodEnd.IsZero() {
		return AttendantReportResponse{}, fmt.Errorf(
			"period end is required",
		)
	}

	if !periodEnd.After(periodStart) {
		return AttendantReportResponse{}, fmt.Errorf(
			"period end must be after period start",
		)
	}

	start := periodStart.UTC().Format(time.RFC3339)
	end := periodEnd.UTC().Format(time.RFC3339)

	summary, err := s.repository.GetSummary(
		ctx,
		tenantID,
		attendantID,
		start,
		end,
	)
	if err != nil {
		return AttendantReportResponse{}, fmt.Errorf(
			"get attendant report summary: %w",
			err,
		)
	}

	paymentMethods, err := s.repository.GetPaymentMethods(
		ctx,
		tenantID,
		attendantID,
		start,
		end,
	)
	if err != nil {
		return AttendantReportResponse{}, fmt.Errorf(
			"get attendant payment methods: %w",
			err,
		)
	}

	dailyTotals, err := s.repository.GetDailyTotals(
		ctx,
		tenantID,
		attendantID,
		start,
		end,
	)
	if err != nil {
		return AttendantReportResponse{}, fmt.Errorf(
			"get attendant daily totals: %w",
			err,
		)
	}

	return AttendantReportResponse{
		Summary:        summary,
		PaymentMethods: paymentMethods,
		DailyTotals:    dailyTotals,
	}, nil
}
