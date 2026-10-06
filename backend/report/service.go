package report

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTenantID   = errors.New("invalid tenant id")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidBranchID   = errors.New("invalid branch id")
	ErrInvalidPeriod     = errors.New("invalid report period")
	ErrInvalidReportType = errors.New("invalid report type")

	ErrReportAccessDenied = errors.New(
		"report access denied",
	)

	ErrInvalidReportRole = errors.New(
		"invalid report role",
	)
)

type Service struct {
	repository    *Repository
	encryptionKey []byte
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository:    repository,
		encryptionKey: loadComplianceEncryptionKey(),
	}
}

func (s *Service) GenerateComplianceReport(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	request GenerateComplianceReportRequest,
) (*ComplianceReportSnapshot, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, ErrInvalidTenantID
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	role := strings.TrimSpace(userRole)

	if role != "owner" && role != "attendant" {
		return nil, ErrInvalidReportRole
	}

	branchID := strings.TrimSpace(request.BranchID)

	/*
		ATTENDANT SECURITY RULE

		Never trust branch_id supplied by the attendant.

		The attendant's branch is resolved from
		attendant_branch_assignments using the authenticated user.
	*/
	if role == "attendant" {
		assignedBranchID, err := s.getAttendantBranchID(
			ctx,
			tenantID,
			userID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"resolve attendant branch: %w",
				err,
			)
		}

		branchID = assignedBranchID
	}

	if branchID == "" {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	reportType := strings.TrimSpace(request.ReportType)

	if reportType != "cak_compliance" {
		return nil, ErrInvalidReportType
	}

	periodStart, err := time.Parse(
		"2006-01-02",
		strings.TrimSpace(request.PeriodStart),
	)
	if err != nil {
		return nil, ErrInvalidPeriod
	}

	periodEndDate, err := time.Parse(
		"2006-01-02",
		strings.TrimSpace(request.PeriodEnd),
	)
	if err != nil {
		return nil, ErrInvalidPeriod
	}

	if periodEndDate.Before(periodStart) {
		return nil, ErrInvalidPeriod
	}

	/*
		Make the end date inclusive.

		For example:

		period_start = 2026-01-01
		period_end   = 2026-01-31

		Queries use:

		>= 2026-01-01
		<  2026-02-01
	*/
	periodEndExclusive := periodEndDate.AddDate(0, 0, 1)

	/*
		For attendants, verify the assignment again immediately
		before reading branch data.
	*/
	if role == "attendant" {
		if err := s.verifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			branchID,
		); err != nil {
			return nil, fmt.Errorf(
				"verify attendant branch access: %w",
				err,
			)
		}
	}

	/*
		------------------------------------------------------------
		BRANCH
		------------------------------------------------------------
	*/

	branch, err := s.repository.GetBranchForCompliance(
		ctx,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get compliance branch: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		CUSTOMERS
		------------------------------------------------------------
	*/

	customers, err := s.repository.ListComplianceCustomers(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance customers: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		TERMINALS
		------------------------------------------------------------
	*/

	terminals, err := s.repository.ListComplianceTerminals(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance terminals: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		SESSIONS
		------------------------------------------------------------
	*/

	sessions, err := s.repository.ListComplianceSessions(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance sessions: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		SALES
		------------------------------------------------------------
	*/

	sales, err := s.repository.ListComplianceSales(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance sales: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		SALE ITEMS
		------------------------------------------------------------
	*/

	saleItems, err := s.repository.ListComplianceSaleItems(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance sale items: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		PAYMENTS
		------------------------------------------------------------
	*/

	payments, err := s.repository.ListCompliancePayments(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance payments: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		RECEIPTS
		------------------------------------------------------------
	*/

	receipts, err := s.repository.ListComplianceReceipts(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance receipts: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		RECEIPT REPRINTS
		------------------------------------------------------------
	*/

	reprints, err := s.repository.ListComplianceReceiptReprints(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance receipt reprints: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		EXPENSES
		------------------------------------------------------------
	*/

	expenses, err := s.repository.ListComplianceExpenses(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance expenses: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		AUDIT LOGS
		------------------------------------------------------------
	*/

	auditEvents, err := s.repository.ListComplianceAuditLogs(
		ctx,
		tenantID,
		branchID,
		periodStart,
		periodEndExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance audit logs: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		CUSTOMER RECORDS
		------------------------------------------------------------
	*/

	customerRecords := make([]map[string]any, 0, len(customers))

	for _, customer := range customers {
		record := map[string]any{
			"id":            customer.ID,
			"branch_id":     customer.BranchID,
			"customer_type": customer.CustomerType,
			"full_name":     customer.FullName,
			"phone":         customer.Phone,
			"created_at":    customer.CreatedAt,
			"updated_at":    customer.UpdatedAt,
		}

		if customer.IDNumberEncrypted != nil {
			idNumber, err := s.decryptID(
				*customer.IDNumberEncrypted,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"decrypt customer ID %s: %w",
					customer.ID,
					err,
				)
			}

			record["id_number"] = idNumber
		}

		if customer.ParentName != nil {
			record["parent_name"] = *customer.ParentName
		}

		if customer.ParentPhone != nil {
			record["parent_phone"] = *customer.ParentPhone
		}

		if customer.ParentIDNumberEncrypt != nil {
			parentID, err := s.decryptID(
				*customer.ParentIDNumberEncrypt,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"decrypt parent ID for customer %s: %w",
					customer.ID,
					err,
				)
			}

			record["parent_id_number"] = parentID
		}

		customerRecords = append(
			customerRecords,
			record,
		)
	}

	/*
		------------------------------------------------------------
		TERMINAL RECORDS
		------------------------------------------------------------
	*/

	terminalRecords := make([]map[string]any, 0, len(terminals))

	for _, terminal := range terminals {
		terminalRecords = append(
			terminalRecords,
			map[string]any{
				"id":                terminal.ID,
				"branch_id":         terminal.BranchID,
				"terminal_code":     terminal.TerminalCode,
				"machine_name":      terminal.MachineName,
				"device_identifier": terminal.DeviceIdentifier,
				"status":            terminal.Status,
				"registered_at":     terminal.RegisteredAt,
				"last_seen_at":      terminal.LastSeenAt,
				"disabled_at":       terminal.DisabledAt,
				"created_at":        terminal.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		SESSION RECORDS
		------------------------------------------------------------
	*/

	sessionRecords := make([]map[string]any, 0, len(sessions))

	for _, session := range sessions {
		usedMinutes, remainingMinutes := calculateSessionMinutes(
			session.StartedAt,
			session.EndedAt,
			session.AllowedMinutes,
			time.Now().UTC(),
		)

		record := map[string]any{
			"session_id":        session.ID,
			"branch_id":         session.BranchID,
			"terminal_id":       session.TerminalID,
			"terminal_code":     session.TerminalCode,
			"machine_name":      session.MachineName,
			"customer_id":       session.CustomerID,
			"customer_name":     session.CustomerName,
			"customer_type":     session.CustomerType,
			"attendant_id":      session.AttendantID,
			"attendant_name":    session.AttendantName,
			"session_type":      session.SessionType,
			"status":            session.Status,
			"started_at":        session.StartedAt,
			"ended_at":          session.EndedAt,
			"prepaid_amount":    session.PrepaidAmount,
			"allowed_minutes":   session.AllowedMinutes,
			"used_minutes":      usedMinutes,
			"remaining_minutes": remainingMinutes,
			"rate_per_minute":   session.RatePerMinute,
			"minimum_charge":    session.MinimumCharge,
			"final_amount":      session.FinalAmount,
		}

		/*
			The customer's identification number is included directly
			in the session record so an investigator does not have to
			cross-reference the customer table manually.
		*/
		if session.IDNumberEncrypted != nil {
			idNumber, err := s.decryptID(
				*session.IDNumberEncrypted,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"decrypt session customer ID %s: %w",
					session.CustomerID,
					err,
				)
			}

			record["identification_number"] = idNumber
		}

		sessionRecords = append(
			sessionRecords,
			record,
		)
	}

	/*
		------------------------------------------------------------
		SALE RECORDS
		------------------------------------------------------------
	*/

	saleRecords := make([]map[string]any, 0, len(sales))

	for _, sale := range sales {
		saleRecords = append(
			saleRecords,
			map[string]any{
				"id":                 sale.ID,
				"branch_id":          sale.BranchID,
				"customer_id":        sale.CustomerID,
				"customer_name":      sale.CustomerName,
				"session_id":         sale.SessionID,
				"session_started_at": sale.SessionStartedAt,
				"terminal_id":        sale.TerminalID,
				"attendant_id":       sale.AttendantID,
				"attendant_name":     sale.AttendantName,
				"subtotal":           sale.Subtotal,
				"discount_type":      sale.DiscountType,
				"discount_value":     sale.DiscountValue,
				"discount_amount":    sale.DiscountAmount,
				"total_amount":       sale.TotalAmount,
				"status":             sale.Status,
				"void_reason":        sale.VoidReason,
				"voided_by":          sale.VoidedBy,
				"voided_at":          sale.VoidedAt,
				"created_at":         sale.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		SALE ITEM RECORDS
		------------------------------------------------------------
	*/

	saleItemRecords := make([]map[string]any, 0, len(saleItems))

	for _, item := range saleItems {
		saleItemRecords = append(
			saleItemRecords,
			map[string]any{
				"id":           item.ID,
				"sale_id":      item.SaleID,
				"service_id":   item.ServiceID,
				"service_name": item.ServiceName,
				"description":  item.Description,
				"quantity":     item.Quantity,
				"unit_price":   item.UnitPrice,
				"line_total":   item.LineTotal,
				"created_at":   item.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		PAYMENT RECORDS
		------------------------------------------------------------
	*/

	paymentRecords := make([]map[string]any, 0, len(payments))

	for _, payment := range payments {
		paymentRecords = append(
			paymentRecords,
			map[string]any{
				"id":                      payment.ID,
				"branch_id":               payment.BranchID,
				"sale_id":                 payment.SaleID,
				"method":                  payment.Method,
				"status":                  payment.Status,
				"amount":                  payment.Amount,
				"phone":                   payment.Phone,
				"external_reference":      payment.ExternalReference,
				"mpesa_receipt_number":    payment.MPesaReceiptNumber,
				"provider_request_id":     payment.ProviderRequestID,
				"provider_transaction_id": payment.ProviderTransactionID,
				"failure_reason":          payment.FailureReason,
				"confirmed_at":            payment.ConfirmedAt,
				"created_at":              payment.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		RECEIPT RECORDS
		------------------------------------------------------------
	*/

	receiptRecords := make([]map[string]any, 0, len(receipts))

	for _, receipt := range receipts {
		receiptRecords = append(
			receiptRecords,
			map[string]any{
				"id":             receipt.ID,
				"branch_id":      receipt.BranchID,
				"sale_id":        receipt.SaleID,
				"payment_id":     receipt.PaymentID,
				"receipt_number": receipt.ReceiptNumber,
				"receipt_type":   receipt.ReceiptType,
				"issued_at":      receipt.IssuedAt,
				"printed_at":     receipt.PrintedAt,
				"reprint_count":  receipt.ReprintCount,
				"created_at":     receipt.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		RECEIPT REPRINT RECORDS
		------------------------------------------------------------
	*/

	reprintRecords := make([]map[string]any, 0, len(reprints))

	for _, reprint := range reprints {
		reprintRecords = append(
			reprintRecords,
			map[string]any{
				"id":            reprint.ID,
				"receipt_id":    reprint.ReceiptID,
				"user_id":       reprint.UserID,
				"user_name":     reprint.UserName,
				"terminal_id":   reprint.TerminalID,
				"terminal_code": reprint.TerminalCode,
				"reason":        reprint.Reason,
				"created_at":    reprint.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		EXPENSE RECORDS
		------------------------------------------------------------
	*/

	expenseRecords := make([]map[string]any, 0, len(expenses))

	for _, expense := range expenses {
		expenseRecords = append(
			expenseRecords,
			map[string]any{
				"id":             expense.ID,
				"branch_id":      expense.BranchID,
				"recorded_by":    expense.RecordedBy,
				"category":       expense.Category,
				"description":    expense.Description,
				"amount":         expense.Amount,
				"currency":       expense.Currency,
				"payment_method": expense.PaymentMethod,
				"reference":      expense.Reference,
				"expense_date":   expense.ExpenseDate,
				"status":         expense.Status,
				"created_at":     expense.CreatedAt,
				"updated_at":     expense.UpdatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		AUDIT RECORDS
		------------------------------------------------------------
	*/

	auditRecords := make([]map[string]any, 0, len(auditEvents))

	for _, event := range auditEvents {
		auditRecords = append(
			auditRecords,
			map[string]any{
				"id":          event.ID,
				"branch_id":   event.BranchID,
				"user_id":     event.UserID,
				"user_name":   event.UserName,
				"user_role":   event.UserRole,
				"action":      event.Action,
				"entity_type": event.EntityType,
				"entity_id":   event.EntityID,
				"old_value":   event.OldValue,
				"new_value":   event.NewValue,
				"reason":      event.Reason,
				"ip_address":  event.IP,
				"device_id":   event.DeviceID,
				"created_at":  event.CreatedAt,
			},
		)
	}

	/*
		------------------------------------------------------------
		BUILD SNAPSHOT
		------------------------------------------------------------
	*/

	generatedAt := time.Now().UTC()

	retentionYears := loadReportRetentionYears()

	snapshot := map[string]any{
		"report_type": "cak_compliance",
		"tenant_id":   tenantID,

		"branch": map[string]any{
			"id":      branch.ID,
			"name":    branch.Name,
			"address": branch.Address,
			"phone":   branch.Phone,
			"status":  branch.Status,
		},

		"period": map[string]any{
			"start": periodStart,
			"end":   periodEndDate,
		},

		"generated_at": generatedAt,

		"retention": map[string]any{
			"years": retentionYears,
		},

		"records": map[string]any{
			"customers":        customerRecords,
			"terminals":        terminalRecords,
			"sessions":         sessionRecords,
			"sales":            saleRecords,
			"sale_items":       saleItemRecords,
			"payments":         paymentRecords,
			"receipts":         receiptRecords,
			"receipt_reprints": reprintRecords,
			"expenses":         expenseRecords,
			"audit_events":     auditRecords,
		},

		/*
			We deliberately do not collect or include browsing history.
		*/
		"excluded": []string{
			"browsing_history",
			"visited_urls",
			"search_history",
			"search_queries",
		},

		"limitations": []string{
			"Pause and resume events are not included because the current database schema does not store separate pause/resume events.",
			"Session extensions are not included because the current database schema does not store separate extension events.",
			"Printing records are not included until the existing printing database schema is connected to the report module.",
		},
	}

	/*
		------------------------------------------------------------
		HASH SNAPSHOT
		------------------------------------------------------------
	*/

	snapshotHash, err := hashSnapshot(snapshot)
	if err != nil {
		return nil, fmt.Errorf(
			"hash compliance snapshot: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		RETENTION
		------------------------------------------------------------
	*/

	retentionUntil := generatedAt.AddDate(
		retentionYears,
		0,
		0,
	)

	generatedBy := userID

	/*
		------------------------------------------------------------
		SAVE SNAPSHOT
		------------------------------------------------------------
	*/

	result, err := s.repository.CreateSnapshot(
		ctx,
		CreateSnapshotParams{
			ID:             uuid.New().String(),
			TenantID:       tenantID,
			BranchID:       &branchID,
			GeneratedBy:    &generatedBy,
			ReportType:     "cak_compliance",
			PeriodStart:    periodStart,
			PeriodEnd:      periodEndDate,
			RetentionUntil: retentionUntil,
			Snapshot:       snapshot,
			SnapshotHash:   &snapshotHash,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create compliance snapshot: %w",
			err,
		)
	}

	/*
		------------------------------------------------------------
		REPORT GENERATION LOG
		------------------------------------------------------------
	*/

	log := ReportGenerationLog{
		ID:          uuid.New().String(),
		TenantID:    &tenantID,
		BranchID:    &branchID,
		GeneratedBy: &generatedBy,
		UserRole:    stringPtr(role),
		ReportType:  "cak_compliance",
		PeriodStart: periodStart,
		PeriodEnd:   periodEndDate,
		Format:      "json",
	}

	if err := s.repository.CreateGenerationLog(ctx, log); err != nil {
		return nil, fmt.Errorf(
			"create report generation log: %w",
			err,
		)
	}

	return result, nil
}

func (s *Service) GetComplianceReport(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	snapshotID string,
) (*ComplianceReportSnapshot, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, ErrInvalidTenantID
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	if _, err := uuid.Parse(snapshotID); err != nil {
		return nil, ErrReportNotFound
	}

	role := strings.TrimSpace(userRole)

	if role != "owner" && role != "attendant" {
		return nil, ErrInvalidReportRole
	}

	report, err := s.repository.GetSnapshot(
		ctx,
		tenantID,
		snapshotID,
	)
	if err != nil {
		return nil, err
	}

	if role == "attendant" {
		if report.BranchID == nil {
			return nil, ErrReportAccessDenied
		}

		if err := s.verifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			*report.BranchID,
		); err != nil {
			return nil, err
		}
	}

	return report, nil
}

func (s *Service) ListComplianceReports(
	ctx context.Context,
	tenantID string,
	userID string,
	userRole string,
	branchID string,
) ([]ComplianceReportSnapshot, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, ErrInvalidTenantID
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	role := strings.TrimSpace(userRole)

	if role != "owner" && role != "attendant" {
		return nil, ErrInvalidReportRole
	}

	/*
		ATTENDANT SECURITY RULE

		Ignore any branch supplied by the attendant and resolve
		the branch from the authenticated assignment.
	*/
	if role == "attendant" {
		assignedBranchID, err := s.getAttendantBranchID(
			ctx,
			tenantID,
			userID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"resolve attendant branch for report list: %w",
				err,
			)
		}

		branchID = assignedBranchID
	}

	if branchID == "" {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if role == "attendant" {
		if err := s.verifyAttendantBranchAccess(
			ctx,
			tenantID,
			userID,
			branchID,
		); err != nil {
			return nil, fmt.Errorf(
				"verify attendant branch access for report list: %w",
				err,
			)
		}
	}

	results, err := s.repository.ListSnapshots(
		ctx,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list compliance report snapshots: %w",
			err,
		)
	}

	return results, nil
}

func (s *Service) getAttendantBranchID(
	ctx context.Context,
	tenantID string,
	userID string,
) (string, error) {
	var branchID string

	err := s.repository.db.QueryRow(
		ctx,
		`
		SELECT aba.branch_id::text
		FROM attendant_branch_assignments aba
		INNER JOIN users u
			ON u.id = aba.user_id
		INNER JOIN branches b
			ON b.id = aba.branch_id
		WHERE aba.user_id = $1
		  AND aba.tenant_id = $2
		  AND aba.unassigned_at IS NULL
		  AND u.role = 'attendant'
		  AND u.status = 'active'
		  AND b.status = 'active'
		  AND b.tenant_id = $2
		ORDER BY aba.assigned_at DESC
		LIMIT 1
		`,
		userID,
		tenantID,
	).Scan(&branchID)

	if err != nil {
		if strings.Contains(
			strings.ToLower(err.Error()),
			"no rows",
		) {
			return "", ErrReportAccessDenied
		}

		return "", err
	}

	if branchID == "" {
		return "", ErrReportAccessDenied
	}

	return branchID, nil
}

func (s *Service) verifyAttendantBranchAccess(
	ctx context.Context,
	tenantID string,
	userID string,
	branchID string,
) error {
	var exists bool

	err := s.repository.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM attendant_branch_assignments aba
			INNER JOIN users u
				ON u.id = aba.user_id
			INNER JOIN branches b
				ON b.id = aba.branch_id
			WHERE aba.user_id = $1
			  AND aba.tenant_id = $2
			  AND aba.branch_id = $3
			  AND aba.unassigned_at IS NULL
			  AND u.role = 'attendant'
			  AND u.status = 'active'
			  AND b.status = 'active'
			  AND b.tenant_id = $2
		)
		`,
		userID,
		tenantID,
		branchID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if !exists {
		return ErrReportAccessDenied
	}

	return nil
}

func hashSnapshot(snapshot map[string]any) (string, error) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:]), nil
}

func calculateSessionMinutes(
	start time.Time,
	end *time.Time,
	allowedMinutes *int,
	now time.Time,
) (int64, int64) {
	finish := now

	if end != nil && end.Before(finish) {
		finish = *end
	}

	if finish.Before(start) {
		return 0, 0
	}

	used := int64(finish.Sub(start) / time.Minute)

	if used < 0 {
		used = 0
	}

	if allowedMinutes == nil {
		return used, 0
	}

	remaining := int64(*allowedMinutes) - used

	if remaining < 0 {
		remaining = 0
	}

	return used, remaining
}

func loadComplianceEncryptionKey() []byte {
	value := strings.TrimSpace(
		os.Getenv("CUSTOMER_ID_ENCRYPTION_KEY"),
	)

	if value == "" {
		panic("CUSTOMER_ID_ENCRYPTION_KEY is required")
	}

	decoded, err := base64.StdEncoding.DecodeString(value)

	if err == nil && len(decoded) == 32 {
		return decoded
	}

	if len(value) == 64 {
		decoded, err := base64.RawStdEncoding.DecodeString(value)

		if err == nil && len(decoded) == 32 {
			return decoded
		}
	}

	panic(
		"CUSTOMER_ID_ENCRYPTION_KEY must contain a base64-encoded 32-byte AES-256 key",
	)
}

func (s *Service) decryptID(value string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	encoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()

	if len(encoded) < nonceSize {
		return "", errors.New("encrypted ID is too short")
	}

	nonce := encoded[:nonceSize]
	ciphertext := encoded[nonceSize:]

	plaintext, err := gcm.Open(
		nil,
		nonce,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func loadReportRetentionYears() int {
	const defaultRetentionYears = 5

	value := strings.TrimSpace(
		os.Getenv("COMPLIANCE_REPORT_RETENTION_YEARS"),
	)

	if value == "" {
		return defaultRetentionYears
	}

	years, err := strconv.Atoi(value)

	if err != nil || years < 1 {
		return defaultRetentionYears
	}

	return years
}

func stringPtr(value string) *string {
	return &value
}
