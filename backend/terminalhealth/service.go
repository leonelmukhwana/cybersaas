package terminalhealth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTerminalID = errors.New("invalid terminal id")
	ErrInvalidHealth     = errors.New("invalid terminal health report")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// ============================================================
// REPORT HEALTH
// ============================================================

func (s *Service) ReportHealth(
	ctx context.Context,
	terminalID string,
	req ReportHealthRequest,
) (*TerminalHealth, error) {

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	if err := validateHealthRequest(req); err != nil {
		return nil, err
	}

	// The backend determines the final health status.
	// The status sent by the PC Client is not trusted.
	req.Status = determineHealthStatus(req)

	health, err := s.repository.UpsertHealth(
		ctx,
		terminalID,
		req,
	)
	if err != nil {
		return nil, err
	}

	return health, nil
}

// ============================================================
// GET SINGLE TERMINAL HEALTH
// ============================================================

func (s *Service) GetTerminalHealth(
	ctx context.Context,
	tenantID string,
	terminalID string,
) (*TerminalHealth, error) {

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, ErrInvalidTerminalID
	}

	health, err := s.repository.GetTerminalHealth(
		ctx,
		tenantID,
		terminalID,
	)
	if err != nil {
		return nil, err
	}

	health.Status = HealthState(*health)

	return health, nil
}

// ============================================================
// LIST BRANCH TERMINAL HEALTH
// ============================================================

func (s *Service) ListBranchHealth(
	ctx context.Context,
	tenantID string,
	branchID string,
) ([]TerminalHealth, int64, error) {

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, 0, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, 0, errors.New("invalid branch id")
	}

	terminals, total, err := s.repository.ListBranchHealth(
		ctx,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, 0, err
	}

	for i := range terminals {
		terminals[i].Status = HealthState(terminals[i])
	}

	return terminals, total, nil
}

// ============================================================
// HEALTH STATE
// ============================================================

// HealthState returns the effective health state of a terminal.
//
// A terminal that has stopped reporting must not continue to
// appear healthy simply because its last stored report was healthy.
func HealthState(health TerminalHealth) string {
	if !IsHealthFresh(health.LastSeenAt) {
		return "offline"
	}

	return health.Status
}

// ============================================================
// VALIDATION
// ============================================================

func validateHealthRequest(req ReportHealthRequest) error {
	if len(req.NetworkIssue) > 1000 {
		return fmt.Errorf(
			"%w: network issue is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.HealthMessage) > 2000 {
		return fmt.Errorf(
			"%w: health message is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.Issues) > 50 {
		return fmt.Errorf(
			"%w: too many issues",
			ErrInvalidHealth,
		)
	}

	for _, issue := range req.Issues {
		if len(issue) > 500 {
			return fmt.Errorf(
				"%w: issue is too long",
				ErrInvalidHealth,
			)
		}
	}

	if req.CPUUsagePercent < 0 ||
		req.CPUUsagePercent > 100 {
		return fmt.Errorf(
			"%w: cpu usage must be between 0 and 100",
			ErrInvalidHealth,
		)
	}

	if req.MemoryTotal < 0 ||
		req.MemoryUsed < 0 ||
		req.MemoryFree < 0 {
		return fmt.Errorf(
			"%w: memory values cannot be negative",
			ErrInvalidHealth,
		)
	}

	if req.MemoryTotal > 0 {
		if req.MemoryUsed > req.MemoryTotal {
			return fmt.Errorf(
				"%w: memory used cannot exceed total memory",
				ErrInvalidHealth,
			)
		}

		if req.MemoryFree > req.MemoryTotal {
			return fmt.Errorf(
				"%w: memory free cannot exceed total memory",
				ErrInvalidHealth,
			)
		}
	}

	if req.DiskTotal < 0 ||
		req.DiskFree < 0 {
		return fmt.Errorf(
			"%w: disk values cannot be negative",
			ErrInvalidHealth,
		)
	}

	if req.DiskTotal > 0 &&
		req.DiskFree > req.DiskTotal {
		return fmt.Errorf(
			"%w: disk free cannot exceed total disk",
			ErrInvalidHealth,
		)
	}

	if req.BackendLatencyMS < 0 {
		return fmt.Errorf(
			"%w: backend latency cannot be negative",
			ErrInvalidHealth,
		)
	}

	if req.OfflineQueueCount < 0 ||
		req.SyncPendingCount < 0 ||
		req.SyncFailedCount < 0 {
		return fmt.Errorf(
			"%w: synchronization counts cannot be negative",
			ErrInvalidHealth,
		)
	}

	if req.UptimeSeconds < 0 {
		return fmt.Errorf(
			"%w: uptime cannot be negative",
			ErrInvalidHealth,
		)
	}

	if !validClientStatus(req.ClientStatus) {
		return fmt.Errorf(
			"%w: invalid client status",
			ErrInvalidHealth,
		)
	}

	if !validSyncStatus(req.SyncStatus) {
		return fmt.Errorf(
			"%w: invalid sync status",
			ErrInvalidHealth,
		)
	}

	if len(req.SoundIssue) > 1000 {
		return fmt.Errorf(
			"%w: sound issue is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.DriverIssue) > 1000 {
		return fmt.Errorf(
			"%w: driver issue is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.SecurityIssue) > 1000 {
		return fmt.Errorf(
			"%w: security issue is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.OSName) > 100 {
		return fmt.Errorf(
			"%w: operating system name is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.OSVersion) > 100 {
		return fmt.Errorf(
			"%w: operating system version is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.OSArchitecture) > 50 {
		return fmt.Errorf(
			"%w: operating system architecture is too long",
			ErrInvalidHealth,
		)
	}

	if len(req.ClientVersion) > 50 {
		return fmt.Errorf(
			"%w: client version is too long",
			ErrInvalidHealth,
		)
	}

	return nil
}

// ============================================================
// HEALTH STATUS
// ============================================================

func determineHealthStatus(req ReportHealthRequest) string {

	// No network adapter and no database connection means
	// the terminal is effectively offline.
	if !req.NetworkAdapterAvailable &&
		!req.DatabaseAvailable {
		return "offline"
	}

	// No network adapter means the terminal cannot communicate
	// with the local network/backend.
	if !req.NetworkAdapterAvailable {
		return "offline"
	}

	// Backend/database problems require attention.
	if !req.APIAvailable ||
		!req.DatabaseAvailable {
		return "attention"
	}

	// DNS failure while internet is expected to be available.
	if req.InternetAvailable &&
		!req.DNSAvailable {
		return "attention"
	}

	// Internet may be unavailable while the terminal can still
	// operate offline.
	if !req.InternetAvailable {
		return "attention"
	}

	// Core hardware/service availability.
	if !req.DiskAvailable ||
		!req.MemoryAvailable ||
		!req.PrinterAvailable {
		return "attention"
	}

	// Sound and drivers.
	if !req.SoundAvailable ||
		!req.DriversAvailable {
		return "attention"
	}

	// Windows security.
	if !req.SecurityAvailable ||
		!req.AntivirusEnabled {
		return "attention"
	}

	// Extremely high CPU usage.
	if req.CPUUsagePercent >= 95 {
		return "attention"
	}

	// Disk nearly full.
	if req.DiskTotal > 0 {
		diskUsage := float64(
			req.DiskTotal-req.DiskFree,
		) / float64(req.DiskTotal)

		if diskUsage >= 0.95 {
			return "attention"
		}
	}

	// Memory nearly full.
	if req.MemoryTotal > 0 {
		memoryUsage := float64(
			req.MemoryUsed,
		) / float64(req.MemoryTotal)

		if memoryUsage >= 0.95 {
			return "attention"
		}
	}

	// Offline/synchronization backlog.
	if req.OfflineQueueCount > 0 ||
		req.SyncPendingCount > 0 ||
		req.SyncFailedCount > 0 {
		return "attention"
	}

	if req.SyncStatus == "failed" ||
		req.SyncStatus == "offline" {
		return "attention"
	}

	// PC Client itself reports a problem.
	if req.ClientStatus == "error" ||
		req.ClientStatus == "stopped" {
		return "attention"
	}

	// Explicit diagnostic messages.
	if strings.TrimSpace(req.NetworkIssue) != "" ||
		strings.TrimSpace(req.SoundIssue) != "" ||
		strings.TrimSpace(req.DriverIssue) != "" ||
		strings.TrimSpace(req.SecurityIssue) != "" {
		return "attention"
	}

	if len(req.Issues) > 0 {
		return "attention"
	}

	return "healthy"
}

// ============================================================
// CLIENT STATUS VALIDATION
// ============================================================

func validClientStatus(status string) bool {
	switch status {
	case "running",
		"stopped",
		"error",
		"unknown":
		return true
	default:
		return false
	}
}

// ============================================================
// SYNC STATUS VALIDATION
// ============================================================

func validSyncStatus(status string) bool {
	switch status {
	case "synced",
		"pending",
		"failed",
		"offline":
		return true
	default:
		return false
	}
}

func (s *Service) GetAttendantHealthSummary(
	ctx context.Context,
	tenantID string,
	attendantID string,
) (*TerminalHealthSummary, error) {

	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(attendantID); err != nil {
		return nil, errors.New("invalid attendant id")
	}

	return s.repository.GetAttendantHealthSummary(
		ctx,
		tenantID,
		attendantID,
	)
}

func (s *Service) ListAttendantTerminals(
	ctx context.Context,
	tenantID string,
	attendantID string,
) ([]AttendantTerminal, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(attendantID); err != nil {
		return nil, errors.New("invalid attendant id")
	}

	return s.repository.ListAttendantTerminals(
		ctx,
		tenantID,
		attendantID,
	)
}

func (s *Service) GetAttendantTerminalHealth(
	ctx context.Context,
	tenantID string,
	attendantID string,
	terminalID string,
) (*TerminalHealth, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(attendantID); err != nil {
		return nil, errors.New("invalid attendant id")
	}

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, errors.New("invalid terminal id")
	}

	return s.repository.GetAttendantTerminalHealth(
		ctx,
		tenantID,
		attendantID,
		terminalID,
	)
}

func IsHealthFresh(lastSeen time.Time) bool {
	return time.Since(lastSeen) <= 2*time.Minute
}
