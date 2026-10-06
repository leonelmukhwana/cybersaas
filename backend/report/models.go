package report

import "time"

type ComplianceReportSnapshot struct {
	ID             string         `json:"id"`
	TenantID       string         `json:"tenant_id"`
	BranchID       *string        `json:"branch_id,omitempty"`
	GeneratedBy    *string        `json:"generated_by,omitempty"`
	ReportType     string         `json:"report_type"`
	PeriodStart    time.Time      `json:"period_start"`
	PeriodEnd      time.Time      `json:"period_end"`
	GeneratedAt    time.Time      `json:"generated_at"`
	RetentionUntil time.Time      `json:"retention_until"`
	Snapshot       map[string]any `json:"snapshot"`
	SnapshotHash   *string        `json:"snapshot_hash,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type ReportGenerationLog struct {
	ID          string    `json:"id"`
	TenantID    *string   `json:"tenant_id,omitempty"`
	BranchID    *string   `json:"branch_id,omitempty"`
	GeneratedBy *string   `json:"generated_by,omitempty"`
	UserRole    *string   `json:"user_role,omitempty"`
	ReportType  string    `json:"report_type"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Format      string    `json:"format"`
	CreatedAt   time.Time `json:"created_at"`
}

type GenerateComplianceReportRequest struct {
	BranchID    string `json:"branch_id" binding:"required"`
	ReportType  string `json:"report_type" binding:"required"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

type ComplianceReportResponse struct {
	Snapshot *ComplianceReportSnapshot `json:"snapshot"`
}
