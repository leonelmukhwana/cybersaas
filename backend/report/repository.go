package report

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrReportNotFound = errors.New("report not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type CreateSnapshotParams struct {
	ID             string
	TenantID       string
	BranchID       *string
	GeneratedBy    *string
	ReportType     string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	RetentionUntil time.Time
	Snapshot       map[string]any
	SnapshotHash   *string
}

func (r *Repository) CreateSnapshot(
	ctx context.Context,
	params CreateSnapshotParams,
) (*ComplianceReportSnapshot, error) {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(params.TenantID)
	if err != nil {
		return nil, err
	}

	var branchID *uuid.UUID

	if params.BranchID != nil {
		value, err := uuid.Parse(*params.BranchID)
		if err != nil {
			return nil, err
		}

		branchID = &value
	}

	var generatedBy *uuid.UUID

	if params.GeneratedBy != nil {
		value, err := uuid.Parse(*params.GeneratedBy)
		if err != nil {
			return nil, err
		}

		generatedBy = &value
	}

	snapshotJSON, err := json.Marshal(params.Snapshot)
	if err != nil {
		return nil, err
	}

	if !json.Valid(snapshotJSON) {
		return nil, errors.New("generated compliance snapshot is not valid JSON")
	}

	snapshotText := string(snapshotJSON)

	var result ComplianceReportSnapshot
	var returnedSnapshot []byte

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO compliance_report_snapshots (
			id,
			tenant_id,
			branch_id,
			generated_by,
			report_type,
			period_start,
			period_end,
			retention_until,
			snapshot,
			snapshot_hash
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9::jsonb,
			$10
		)
		RETURNING
			id,
			tenant_id,
			branch_id,
			generated_by,
			report_type,
			period_start,
			period_end,
			generated_at,
			retention_until,
			snapshot,
			snapshot_hash,
			created_at
		`,
		id,
		tenantID,
		branchID,
		generatedBy,
		params.ReportType,
		params.PeriodStart,
		params.PeriodEnd,
		params.RetentionUntil,
		snapshotText,
		params.SnapshotHash,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.GeneratedBy,
		&result.ReportType,
		&result.PeriodStart,
		&result.PeriodEnd,
		&result.GeneratedAt,
		&result.RetentionUntil,
		&returnedSnapshot,
		&result.SnapshotHash,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(returnedSnapshot, &result.Snapshot); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) GetSnapshot(
	ctx context.Context,
	tenantID string,
	snapshotID string,
) (*ComplianceReportSnapshot, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}

	snapshotUUID, err := uuid.Parse(snapshotID)
	if err != nil {
		return nil, err
	}

	var result ComplianceReportSnapshot
	var snapshotJSON []byte

	err = r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			generated_by,
			report_type,
			period_start,
			period_end,
			generated_at,
			retention_until,
			snapshot,
			snapshot_hash,
			created_at
		FROM compliance_report_snapshots
		WHERE id = $1
		  AND tenant_id = $2
		`,
		snapshotUUID,
		tenantUUID,
	).Scan(
		&result.ID,
		&result.TenantID,
		&result.BranchID,
		&result.GeneratedBy,
		&result.ReportType,
		&result.PeriodStart,
		&result.PeriodEnd,
		&result.GeneratedAt,
		&result.RetentionUntil,
		&snapshotJSON,
		&result.SnapshotHash,
		&result.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReportNotFound
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(
		snapshotJSON,
		&result.Snapshot,
	); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) ListSnapshots(
	ctx context.Context,
	tenantID string,
	branchID string,
) ([]ComplianceReportSnapshot, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}

	var branchUUID *uuid.UUID

	if branchID != "" {
		parsed, err := uuid.Parse(branchID)
		if err != nil {
			return nil, err
		}

		branchUUID = &parsed
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			tenant_id,
			branch_id,
			generated_by,
			report_type,
			period_start,
			period_end,
			generated_at,
			retention_until,
			snapshot,
			snapshot_hash,
			created_at
		FROM compliance_report_snapshots
		WHERE tenant_id = $1
		  AND ($2::uuid IS NULL OR branch_id = $2::uuid)
		ORDER BY generated_at DESC
		`,
		tenantUUID,
		branchUUID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]ComplianceReportSnapshot, 0)

	for rows.Next() {
		var item ComplianceReportSnapshot
		var snapshotJSON []byte

		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.BranchID,
			&item.GeneratedBy,
			&item.ReportType,
			&item.PeriodStart,
			&item.PeriodEnd,
			&item.GeneratedAt,
			&item.RetentionUntil,
			&snapshotJSON,
			&item.SnapshotHash,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(
			snapshotJSON,
			&item.Snapshot,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) CreateGenerationLog(
	ctx context.Context,
	log ReportGenerationLog,
) error {
	id, err := uuid.Parse(log.ID)
	if err != nil {
		return err
	}

	var tenantID *uuid.UUID

	if log.TenantID != nil {
		value, err := uuid.Parse(*log.TenantID)
		if err != nil {
			return err
		}

		tenantID = &value
	}

	var branchID *uuid.UUID

	if log.BranchID != nil {
		value, err := uuid.Parse(*log.BranchID)
		if err != nil {
			return err
		}

		branchID = &value
	}

	var generatedBy *uuid.UUID

	if log.GeneratedBy != nil {
		value, err := uuid.Parse(*log.GeneratedBy)
		if err != nil {
			return err
		}

		generatedBy = &value
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO report_generation_logs (
			id,
			tenant_id,
			branch_id,
			generated_by,
			user_role,
			report_type,
			period_start,
			period_end,
			format
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`,
		id,
		tenantID,
		branchID,
		generatedBy,
		log.UserRole,
		log.ReportType,
		log.PeriodStart,
		log.PeriodEnd,
		log.Format,
	)

	return err
}

// ---------------------------------------------------------------------
// CAK operational report queries
// ---------------------------------------------------------------------

type CAKBranchRecord struct {
	ID      string
	Name    string
	Address *string
	Phone   *string
	Status  string
}

type CAKCustomerRecord struct {
	ID                    string
	BranchID              string
	CustomerType          string
	FullName              string
	Phone                 *string
	IDNumberEncrypted     *string
	ParentName            *string
	ParentPhone           *string
	ParentIDNumberEncrypt *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CAKTerminalRecord struct {
	ID               string
	BranchID         string
	TerminalCode     string
	MachineName      *string
	DeviceIdentifier *string
	Status           string
	RegisteredAt     *time.Time
	LastSeenAt       *time.Time
	DisabledAt       *time.Time
	CreatedAt        time.Time
}

type CAKSessionRecordData struct {
	ID                string
	BranchID          string
	TerminalID        string
	TerminalCode      string
	MachineName       *string
	CustomerID        string
	CustomerName      string
	CustomerType      string
	IDNumberEncrypted *string
	AttendantID       *string
	AttendantName     *string
	SessionType       string
	Status            string
	StartedAt         time.Time
	EndedAt           *time.Time
	PrepaidAmount     *float64
	AllowedMinutes    *int
	RatePerMinute     *float64
	MinimumCharge     *float64
	FinalAmount       *float64
}

type CAKSaleRecord struct {
	ID               string
	BranchID         string
	CustomerID       *string
	CustomerName     *string
	SessionID        *string
	SessionStartedAt *time.Time
	TerminalID       *string
	AttendantID      *string
	AttendantName    *string
	Subtotal         float64
	DiscountType     *string
	DiscountValue    float64
	DiscountAmount   float64
	TotalAmount      float64
	Status           string
	VoidReason       *string
	VoidedBy         *string
	VoidedAt         *time.Time
	CreatedAt        time.Time
}

type CAKSaleItemRecord struct {
	ID          string
	SaleID      string
	ServiceID   *string
	ServiceName *string
	Description string
	Quantity    float64
	UnitPrice   float64
	LineTotal   float64
	CreatedAt   time.Time
}

type CAKPaymentRecord struct {
	ID                    string
	BranchID              string
	SaleID                string
	Method                string
	Status                string
	Amount                float64
	Phone                 *string
	ExternalReference     *string
	MPesaReceiptNumber    *string
	ProviderRequestID     *string
	ProviderTransactionID *string
	FailureReason         *string
	ConfirmedAt           *time.Time
	CreatedAt             time.Time
}

type CAKReceiptRecord struct {
	ID            string
	BranchID      string
	SaleID        string
	PaymentID     *string
	ReceiptNumber string
	ReceiptType   string
	IssuedAt      time.Time
	PrintedAt     *time.Time
	ReprintCount  int
	CreatedAt     time.Time
}

type CAKReceiptReprintRecord struct {
	ID           string
	ReceiptID    string
	UserID       *string
	UserName     *string
	TerminalID   *string
	TerminalCode *string
	Reason       *string
	CreatedAt    time.Time
}

type CAKExpenseRecord struct {
	ID            string
	BranchID      string
	RecordedBy    string
	Category      string
	Description   *string
	Amount        float64
	Currency      string
	PaymentMethod string
	Reference     *string
	ExpenseDate   time.Time
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CAKAuditRecord struct {
	ID         string
	BranchID   *string
	UserID     *string
	UserName   *string
	UserRole   *string
	Action     string
	EntityType *string
	EntityID   *string
	OldValue   map[string]any
	NewValue   map[string]any
	Reason     *string
	IP         *string
	DeviceID   *string
	CreatedAt  time.Time
}

func (r *Repository) GetBranchForCompliance(
	ctx context.Context,
	tenantID string,
	branchID string,
) (*CAKBranchRecord, error) {
	var result CAKBranchRecord

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			address,
			phone,
			status
		FROM branches
		WHERE id = $1
		  AND tenant_id = $2
		`,
		branchID,
		tenantID,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Address,
		&result.Phone,
		&result.Status,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReportNotFound
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) ListComplianceCustomers(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKCustomerRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			c.id,
			c.branch_id,
			c.customer_type,
			c.full_name,
			c.phone,
			c.id_number_encrypted,
			c.parent_name,
			c.parent_phone,
			c.parent_id_number_encrypted,
			c.created_at,
			c.updated_at
		FROM customers c
		WHERE c.tenant_id = $1
		  AND c.branch_id = $2
		  AND c.created_at < $3
		  AND (
				c.created_at >= $4
				OR EXISTS (
					SELECT 1
					FROM sessions s
					WHERE s.customer_id = c.id
					  AND s.tenant_id = $1
					  AND s.branch_id = $2
					  AND s.started_at < $3
					  AND (
							s.ended_at IS NULL
							OR s.ended_at >= $4
					  )
				)
				OR EXISTS (
					SELECT 1
					FROM sales sa
					WHERE sa.customer_id = c.id
					  AND sa.tenant_id = $1
					  AND sa.branch_id = $2
					  AND sa.created_at >= $4
					  AND sa.created_at < $3
				)
		  )
		ORDER BY c.created_at ASC, c.id
		`,
		tenantID,
		branchID,
		end,
		start,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKCustomerRecord, 0)

	for rows.Next() {
		var item CAKCustomerRecord

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.CustomerType,
			&item.FullName,
			&item.Phone,
			&item.IDNumberEncrypted,
			&item.ParentName,
			&item.ParentPhone,
			&item.ParentIDNumberEncrypt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceTerminals(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKTerminalRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			branch_id,
			terminal_code,
			machine_name,
			device_identifier,
			status,
			registered_at,
			last_seen_at,
			disabled_at,
			created_at
		FROM terminals
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND created_at < $3
		  AND (
				disabled_at IS NULL
				OR disabled_at >= $4
		  )
		ORDER BY created_at ASC, terminal_code ASC
		`,
		tenantID,
		branchID,
		end,
		start,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKTerminalRecord, 0)

	for rows.Next() {
		var item CAKTerminalRecord

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.TerminalCode,
			&item.MachineName,
			&item.DeviceIdentifier,
			&item.Status,
			&item.RegisteredAt,
			&item.LastSeenAt,
			&item.DisabledAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceSessions(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKSessionRecordData, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			s.id,
			s.branch_id,
			s.terminal_id,
			t.terminal_code,
			t.machine_name,
			s.customer_id,
			c.full_name,
			c.customer_type,
			c.id_number_encrypted,
			s.attendant_id,
			u.full_name,
			s.session_type,
			s.status,
			s.started_at,
			s.ended_at,
			s.prepaid_amount,
			s.allowed_minutes,
			s.rate_per_minute,
			s.minimum_charge,
			s.final_amount
		FROM sessions s
		INNER JOIN terminals t
			ON t.id = s.terminal_id
		   AND t.branch_id = s.branch_id
		   AND t.tenant_id = s.tenant_id
		INNER JOIN customers c
			ON c.id = s.customer_id
		   AND c.branch_id = s.branch_id
		   AND c.tenant_id = s.tenant_id
		LEFT JOIN users u
			ON u.id = s.attendant_id
		WHERE s.tenant_id = $1
		  AND s.branch_id = $2
		  AND s.started_at < $4
		  AND (
				s.ended_at IS NULL
				OR s.ended_at >= $3
		  )
		ORDER BY s.started_at ASC, s.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKSessionRecordData, 0)

	for rows.Next() {
		var item CAKSessionRecordData

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.TerminalID,
			&item.TerminalCode,
			&item.MachineName,
			&item.CustomerID,
			&item.CustomerName,
			&item.CustomerType,
			&item.IDNumberEncrypted,
			&item.AttendantID,
			&item.AttendantName,
			&item.SessionType,
			&item.Status,
			&item.StartedAt,
			&item.EndedAt,
			&item.PrepaidAmount,
			&item.AllowedMinutes,
			&item.RatePerMinute,
			&item.MinimumCharge,
			&item.FinalAmount,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceSales(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKSaleRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			sa.id,
			sa.branch_id,
			sa.customer_id,
			c.full_name,
			sa.session_id,
			sa.session_started_at,
			sa.terminal_id,
			sa.attendant_id,
			u.full_name,
			sa.subtotal,
			sa.discount_type::text,
			sa.discount_value,
			sa.discount_amount,
			sa.total_amount,
			sa.status::text,
			sa.void_reason,
			sa.voided_by,
			sa.voided_at,
			sa.created_at
		FROM sales sa
		LEFT JOIN customers c
			ON c.id = sa.customer_id
		   AND c.tenant_id = sa.tenant_id
		   AND c.branch_id = sa.branch_id
		LEFT JOIN users u
			ON u.id = sa.attendant_id
		WHERE sa.tenant_id = $1
		  AND sa.branch_id = $2
		  AND sa.created_at >= $3
		  AND sa.created_at < $4
		ORDER BY sa.created_at ASC, sa.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKSaleRecord, 0)

	for rows.Next() {
		var item CAKSaleRecord

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.CustomerID,
			&item.CustomerName,
			&item.SessionID,
			&item.SessionStartedAt,
			&item.TerminalID,
			&item.AttendantID,
			&item.AttendantName,
			&item.Subtotal,
			&item.DiscountType,
			&item.DiscountValue,
			&item.DiscountAmount,
			&item.TotalAmount,
			&item.Status,
			&item.VoidReason,
			&item.VoidedBy,
			&item.VoidedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceSaleItems(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKSaleItemRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			si.id,
			si.sale_id,
			si.service_id,
			s.name,
			si.description,
			si.quantity,
			si.unit_price,
			si.line_total,
			si.created_at
		FROM sale_items si
		INNER JOIN sales sa
			ON sa.id = si.sale_id
		   AND sa.tenant_id = $1
		   AND sa.branch_id = $2
		LEFT JOIN services s
			ON s.id = si.service_id
		WHERE sa.created_at >= $3
		  AND sa.created_at < $4
		ORDER BY si.created_at ASC, si.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKSaleItemRecord, 0)

	for rows.Next() {
		var item CAKSaleItemRecord

		if err := rows.Scan(
			&item.ID,
			&item.SaleID,
			&item.ServiceID,
			&item.ServiceName,
			&item.Description,
			&item.Quantity,
			&item.UnitPrice,
			&item.LineTotal,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListCompliancePayments(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKPaymentRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			p.id,
			p.branch_id,
			p.sale_id,
			p.method::text,
			p.status::text,
			p.amount,
			p.phone,
			p.external_reference,
			p.mpesa_receipt_number,
			p.provider_request_id,
			p.provider_transaction_id,
			p.failure_reason,
			p.confirmed_at,
			p.created_at
		FROM payments p
		WHERE p.tenant_id = $1
		  AND p.branch_id = $2
		  AND p.created_at >= $3
		  AND p.created_at < $4
		ORDER BY p.created_at ASC, p.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKPaymentRecord, 0)

	for rows.Next() {
		var item CAKPaymentRecord

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.SaleID,
			&item.Method,
			&item.Status,
			&item.Amount,
			&item.Phone,
			&item.ExternalReference,
			&item.MPesaReceiptNumber,
			&item.ProviderRequestID,
			&item.ProviderTransactionID,
			&item.FailureReason,
			&item.ConfirmedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceReceipts(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKReceiptRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			branch_id,
			sale_id,
			payment_id,
			receipt_number,
			receipt_type,
			issued_at,
			printed_at,
			reprint_count,
			created_at
		FROM receipts
		WHERE tenant_id = $1
		  AND branch_id = $2
		  AND issued_at >= $3
		  AND issued_at < $4
		ORDER BY issued_at ASC, id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKReceiptRecord, 0)

	for rows.Next() {
		var item CAKReceiptRecord

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.SaleID,
			&item.PaymentID,
			&item.ReceiptNumber,
			&item.ReceiptType,
			&item.IssuedAt,
			&item.PrintedAt,
			&item.ReprintCount,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceReceiptReprints(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKReceiptReprintRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			rr.id,
			rr.receipt_id,
			rr.user_id,
			u.full_name,
			rr.terminal_id,
			t.terminal_code,
			rr.reason,
			rr.created_at
		FROM receipt_reprints rr
		INNER JOIN receipts rcp
			ON rcp.id = rr.receipt_id
		   AND rcp.tenant_id = $1
		   AND rcp.branch_id = $2
		LEFT JOIN users u
			ON u.id = rr.user_id
		LEFT JOIN terminals t
			ON t.id = rr.terminal_id
		WHERE rr.created_at >= $3
		  AND rr.created_at < $4
		ORDER BY rr.created_at ASC, rr.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKReceiptReprintRecord, 0)

	for rows.Next() {
		var item CAKReceiptReprintRecord

		if err := rows.Scan(
			&item.ID,
			&item.ReceiptID,
			&item.UserID,
			&item.UserName,
			&item.TerminalID,
			&item.TerminalCode,
			&item.Reason,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceExpenses(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKExpenseRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			e.id,
			e.branch_id,
			e.recorded_by,
			e.category,
			e.description,
			e.amount,
			e.currency,
			e.payment_method::text,
			e.reference,
			e.expense_date,
			e.status::text,
			e.created_at,
			e.updated_at
		FROM expenses e
		WHERE e.tenant_id = $1
		  AND e.branch_id = $2
		  AND e.expense_date >= $3::date
		  AND e.expense_date < $4::date
		ORDER BY e.expense_date ASC, e.created_at ASC, e.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKExpenseRecord, 0)

	for rows.Next() {
		var item CAKExpenseRecord

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.RecordedBy,
			&item.Category,
			&item.Description,
			&item.Amount,
			&item.Currency,
			&item.PaymentMethod,
			&item.Reference,
			&item.ExpenseDate,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) ListComplianceAuditLogs(
	ctx context.Context,
	tenantID string,
	branchID string,
	start time.Time,
	end time.Time,
) ([]CAKAuditRecord, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			a.id,
			a.branch_id,
			a.user_id,
			u.full_name,
			a.user_role::text,
			a.action,
			a.entity_type,
			a.entity_id,
			a.old_value,
			a.new_value,
			a.reason,
			a.ip_address::text,
			a.device_id,
			a.created_at
		FROM audit_logs a
		LEFT JOIN users u
			ON u.id = a.user_id
		WHERE a.tenant_id = $1
		  AND a.branch_id = $2
		  AND a.created_at >= $3
		  AND a.created_at < $4
		ORDER BY a.created_at ASC, a.id
		`,
		tenantID,
		branchID,
		start,
		end,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results := make([]CAKAuditRecord, 0)

	for rows.Next() {
		var item CAKAuditRecord
		var oldValue []byte
		var newValue []byte

		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.UserID,
			&item.UserName,
			&item.UserRole,
			&item.Action,
			&item.EntityType,
			&item.EntityID,
			&oldValue,
			&newValue,
			&item.Reason,
			&item.IP,
			&item.DeviceID,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		if len(oldValue) > 0 {
			if err := json.Unmarshal(
				oldValue,
				&item.OldValue,
			); err != nil {
				return nil, err
			}
		}

		if len(newValue) > 0 {
			if err := json.Unmarshal(
				newValue,
				&item.NewValue,
			); err != nil {
				return nil, err
			}
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
