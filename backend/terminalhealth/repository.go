package terminalhealth

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ============================================================
// REPORT HEALTH
// ============================================================

// UpsertHealth stores the latest health state for a terminal.
//
// There is intentionally only one row per terminal in
// terminal_health. Health reporting updates the latest state
// instead of creating a new row every few seconds.
func (r *Repository) UpsertHealth(
	ctx context.Context,
	terminalID string,
	req ReportHealthRequest,
) (*TerminalHealth, error) {

	issuesJSON, err := json.Marshal(req.Issues)
	if err != nil {
		return nil, err
	}

	var health TerminalHealth
	var returnedIssuesJSON []byte

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO terminal_health (
			terminal_id,
			status,
			health_message,
			issues,

			cpu_usage_percent,

			memory_available,
			memory_total_bytes,
			memory_used_bytes,
			memory_free_bytes,

			disk_available,
			disk_total_bytes,
			disk_free_bytes,

			network_adapter_available,
			lan_available,
			internet_available,
			dns_available,
			api_available,
			backend_latency_ms,
			network_issue,

			database_available,
			printer_available,

			sound_available,
			sound_issue,

			drivers_available,
			driver_issue,

			security_available,
			antivirus_enabled,
			security_issue,

			os_name,
			os_version,
			os_architecture,

			client_version,
			client_status,

			offline_queue_count,
			sync_pending_count,
			sync_failed_count,
			sync_status,
			last_sync_at,

			uptime_seconds,

			last_health_check,
			last_seen_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4::jsonb,

			$5,

			$6,
			$7,
			$8,
			$9,

			$10,
			$11,
			$12,

			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19,

			$20,
			$21,

			$22,
			$23,

			$24,
			$25,

			$26,
			$27,
			$28,

			$29,
			$30,
			$31,

			$32,
			$33,

			$34,
			$35,
			$36,
			$37,
			$38,

			$39,

			NOW(),
			NOW()
		)
		ON CONFLICT (terminal_id)
		DO UPDATE SET
			status = EXCLUDED.status,
			health_message = EXCLUDED.health_message,
			issues = EXCLUDED.issues,

			cpu_usage_percent =
				EXCLUDED.cpu_usage_percent,

			memory_available =
				EXCLUDED.memory_available,
			memory_total_bytes =
				EXCLUDED.memory_total_bytes,
			memory_used_bytes =
				EXCLUDED.memory_used_bytes,
			memory_free_bytes =
				EXCLUDED.memory_free_bytes,

			disk_available =
				EXCLUDED.disk_available,
			disk_total_bytes =
				EXCLUDED.disk_total_bytes,
			disk_free_bytes =
				EXCLUDED.disk_free_bytes,

			network_adapter_available =
				EXCLUDED.network_adapter_available,
			lan_available =
				EXCLUDED.lan_available,
			internet_available =
				EXCLUDED.internet_available,
			dns_available =
				EXCLUDED.dns_available,
			api_available =
				EXCLUDED.api_available,
			backend_latency_ms =
				EXCLUDED.backend_latency_ms,
			network_issue =
				EXCLUDED.network_issue,

			database_available =
				EXCLUDED.database_available,
			printer_available =
				EXCLUDED.printer_available,

			sound_available =
				EXCLUDED.sound_available,
			sound_issue =
				EXCLUDED.sound_issue,

			drivers_available =
				EXCLUDED.drivers_available,
			driver_issue =
				EXCLUDED.driver_issue,

			security_available =
				EXCLUDED.security_available,
			antivirus_enabled =
				EXCLUDED.antivirus_enabled,
			security_issue =
				EXCLUDED.security_issue,

			os_name =
				EXCLUDED.os_name,
			os_version =
				EXCLUDED.os_version,
			os_architecture =
				EXCLUDED.os_architecture,

			client_version =
				EXCLUDED.client_version,
			client_status =
				EXCLUDED.client_status,

			offline_queue_count =
				EXCLUDED.offline_queue_count,
			sync_pending_count =
				EXCLUDED.sync_pending_count,
			sync_failed_count =
				EXCLUDED.sync_failed_count,
			sync_status =
				EXCLUDED.sync_status,
			last_sync_at =
				EXCLUDED.last_sync_at,

			uptime_seconds =
				EXCLUDED.uptime_seconds,

			last_health_check =
				NOW(),
			last_seen_at =
				NOW(),
			updated_at =
				NOW()
		RETURNING
			terminal_id,
			status,
			health_message,
			issues,

			cpu_usage_percent,

			memory_available,
			memory_total_bytes,
			memory_used_bytes,
			memory_free_bytes,

			disk_available,
			disk_total_bytes,
			disk_free_bytes,

			network_adapter_available,
			lan_available,
			internet_available,
			dns_available,
			api_available,
			backend_latency_ms,
			network_issue,

			database_available,
			printer_available,

			sound_available,
			sound_issue,

			drivers_available,
			driver_issue,

			security_available,
			antivirus_enabled,
			security_issue,

			os_name,
			os_version,
			os_architecture,

			client_version,
			client_status,

			offline_queue_count,
			sync_pending_count,
			sync_failed_count,
			sync_status,
			last_sync_at,

			uptime_seconds,

			last_health_check,
			last_seen_at,
			created_at,
			updated_at
		`,
		terminalID,
		req.Status,
		req.HealthMessage,
		string(issuesJSON),

		req.CPUUsagePercent,

		req.MemoryAvailable,
		req.MemoryTotal,
		req.MemoryUsed,
		req.MemoryFree,

		req.DiskAvailable,
		req.DiskTotal,
		req.DiskFree,

		req.NetworkAdapterAvailable,
		req.LANAvailable,
		req.InternetAvailable,
		req.DNSAvailable,
		req.APIAvailable,
		req.BackendLatencyMS,
		req.NetworkIssue,

		req.DatabaseAvailable,
		req.PrinterAvailable,

		req.SoundAvailable,
		req.SoundIssue,

		req.DriversAvailable,
		req.DriverIssue,

		req.SecurityAvailable,
		req.AntivirusEnabled,
		req.SecurityIssue,

		req.OSName,
		req.OSVersion,
		req.OSArchitecture,

		req.ClientVersion,
		req.ClientStatus,

		req.OfflineQueueCount,
		req.SyncPendingCount,
		req.SyncFailedCount,
		req.SyncStatus,
		req.LastSyncAt,

		req.UptimeSeconds,
	).Scan(
		&health.TerminalID,
		&health.Status,
		&health.HealthMessage,
		&returnedIssuesJSON,

		&health.CPUUsagePercent,

		&health.MemoryAvailable,
		&health.MemoryTotal,
		&health.MemoryUsed,
		&health.MemoryFree,

		&health.DiskAvailable,
		&health.DiskTotal,
		&health.DiskFree,

		&health.NetworkAdapterAvailable,
		&health.LANAvailable,
		&health.InternetAvailable,
		&health.DNSAvailable,
		&health.APIAvailable,
		&health.BackendLatencyMS,
		&health.NetworkIssue,

		&health.DatabaseAvailable,
		&health.PrinterAvailable,

		&health.SoundAvailable,
		&health.SoundIssue,

		&health.DriversAvailable,
		&health.DriverIssue,

		&health.SecurityAvailable,
		&health.AntivirusEnabled,
		&health.SecurityIssue,

		&health.OSName,
		&health.OSVersion,
		&health.OSArchitecture,

		&health.ClientVersion,
		&health.ClientStatus,

		&health.OfflineQueueCount,
		&health.SyncPendingCount,
		&health.SyncFailedCount,
		&health.SyncStatus,
		&health.LastSyncAt,

		&health.UptimeSeconds,

		&health.LastHealthCheck,
		&health.LastSeenAt,
		&health.CreatedAt,
		&health.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(
		returnedIssuesJSON,
		&health.Issues,
	); err != nil {
		return nil, err
	}

	// Health reporting also updates the terminal's general
	// presence timestamp.
	_, err = r.db.Exec(
		ctx,
		`
		UPDATE terminals
		SET
			last_seen_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		`,
		terminalID,
	)

	if err != nil {
		return nil, err
	}

	return &health, nil
}

// ============================================================
// SINGLE TERMINAL
// ============================================================

func (r *Repository) GetTerminalHealth(
	ctx context.Context,
	tenantID string,
	terminalID string,
) (*TerminalHealth, error) {

	var health TerminalHealth
	var issuesJSON []byte

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			th.terminal_id,
			th.status,
			th.health_message,
			th.issues,

			th.cpu_usage_percent,

			th.memory_available,
			th.memory_total_bytes,
			th.memory_used_bytes,
			th.memory_free_bytes,

			th.disk_available,
			th.disk_total_bytes,
			th.disk_free_bytes,

			th.network_adapter_available,
			th.lan_available,
			th.internet_available,
			th.dns_available,
			th.api_available,
			th.backend_latency_ms,
			th.network_issue,

			th.database_available,
			th.printer_available,

			th.sound_available,
			th.sound_issue,

			th.drivers_available,
			th.driver_issue,

			th.security_available,
			th.antivirus_enabled,
			th.security_issue,

			th.os_name,
			th.os_version,
			th.os_architecture,

			th.client_version,
			th.client_status,

			th.offline_queue_count,
			th.sync_pending_count,
			th.sync_failed_count,
			th.sync_status,
			th.last_sync_at,

			th.uptime_seconds,

			th.last_health_check,
			th.last_seen_at,
			th.created_at,
			th.updated_at
		FROM terminal_health th
		INNER JOIN terminals t
			ON t.id = th.terminal_id
		WHERE th.terminal_id = $1
		  AND t.tenant_id = $2
		`,
		terminalID,
		tenantID,
	).Scan(
		&health.TerminalID,
		&health.Status,
		&health.HealthMessage,
		&issuesJSON,

		&health.CPUUsagePercent,

		&health.MemoryAvailable,
		&health.MemoryTotal,
		&health.MemoryUsed,
		&health.MemoryFree,

		&health.DiskAvailable,
		&health.DiskTotal,
		&health.DiskFree,

		&health.NetworkAdapterAvailable,
		&health.LANAvailable,
		&health.InternetAvailable,
		&health.DNSAvailable,
		&health.APIAvailable,
		&health.BackendLatencyMS,
		&health.NetworkIssue,

		&health.DatabaseAvailable,
		&health.PrinterAvailable,

		&health.SoundAvailable,
		&health.SoundIssue,

		&health.DriversAvailable,
		&health.DriverIssue,

		&health.SecurityAvailable,
		&health.AntivirusEnabled,
		&health.SecurityIssue,

		&health.OSName,
		&health.OSVersion,
		&health.OSArchitecture,

		&health.ClientVersion,
		&health.ClientStatus,

		&health.OfflineQueueCount,
		&health.SyncPendingCount,
		&health.SyncFailedCount,
		&health.SyncStatus,
		&health.LastSyncAt,

		&health.UptimeSeconds,

		&health.LastHealthCheck,
		&health.LastSeenAt,
		&health.CreatedAt,
		&health.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("terminal health not found")
		}

		return nil, err
	}

	if err := json.Unmarshal(
		issuesJSON,
		&health.Issues,
	); err != nil {
		return nil, err
	}

	return &health, nil
}

// ============================================================
// BRANCH TERMINAL HEALTH
// ============================================================

func (r *Repository) ListBranchHealth(
	ctx context.Context,
	tenantID string,
	branchID string,
) ([]TerminalHealth, int64, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			th.terminal_id,
			th.status,
			th.health_message,
			th.issues,

			th.cpu_usage_percent,

			th.memory_available,
			th.memory_total_bytes,
			th.memory_used_bytes,
			th.memory_free_bytes,

			th.disk_available,
			th.disk_total_bytes,
			th.disk_free_bytes,

			th.network_adapter_available,
			th.lan_available,
			th.internet_available,
			th.dns_available,
			th.api_available,
			th.backend_latency_ms,
			th.network_issue,

			th.database_available,
			th.printer_available,

			th.sound_available,
			th.sound_issue,

			th.drivers_available,
			th.driver_issue,

			th.security_available,
			th.antivirus_enabled,
			th.security_issue,

			th.os_name,
			th.os_version,
			th.os_architecture,

			th.client_version,
			th.client_status,

			th.offline_queue_count,
			th.sync_pending_count,
			th.sync_failed_count,
			th.sync_status,
			th.last_sync_at,

			th.uptime_seconds,

			th.last_health_check,
			th.last_seen_at,
			th.created_at,
			th.updated_at
		FROM terminal_health th
		INNER JOIN terminals t
			ON t.id = th.terminal_id
		WHERE t.tenant_id = $1
		  AND t.branch_id = $2
		ORDER BY t.terminal_code ASC
		`,
		tenantID,
		branchID,
	)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	terminals := make([]TerminalHealth, 0)

	for rows.Next() {
		var health TerminalHealth
		var issuesJSON []byte

		if err := rows.Scan(
			&health.TerminalID,
			&health.Status,
			&health.HealthMessage,
			&issuesJSON,

			&health.CPUUsagePercent,

			&health.MemoryAvailable,
			&health.MemoryTotal,
			&health.MemoryUsed,
			&health.MemoryFree,

			&health.DiskAvailable,
			&health.DiskTotal,
			&health.DiskFree,

			&health.NetworkAdapterAvailable,
			&health.LANAvailable,
			&health.InternetAvailable,
			&health.DNSAvailable,
			&health.APIAvailable,
			&health.BackendLatencyMS,
			&health.NetworkIssue,

			&health.DatabaseAvailable,
			&health.PrinterAvailable,

			&health.SoundAvailable,
			&health.SoundIssue,

			&health.DriversAvailable,
			&health.DriverIssue,

			&health.SecurityAvailable,
			&health.AntivirusEnabled,
			&health.SecurityIssue,

			&health.OSName,
			&health.OSVersion,
			&health.OSArchitecture,

			&health.ClientVersion,
			&health.ClientStatus,

			&health.OfflineQueueCount,
			&health.SyncPendingCount,
			&health.SyncFailedCount,
			&health.SyncStatus,
			&health.LastSyncAt,

			&health.UptimeSeconds,

			&health.LastHealthCheck,
			&health.LastSeenAt,
			&health.CreatedAt,
			&health.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(
			issuesJSON,
			&health.Issues,
		); err != nil {
			return nil, 0, err
		}

		terminals = append(terminals, health)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64

	err = r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM terminals
		WHERE tenant_id = $1
		  AND branch_id = $2
		`,
		tenantID,
		branchID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	return terminals, total, nil
}

// ============================================================
// ATTENDANT HEALTH SUMMARY
// ============================================================

// GetAttendantHealthSummary returns the health summary for the
// branch currently assigned to the authenticated attendant.
//
// The branch is resolved from attendant_branch_assignments.
// The caller cannot supply a branch ID, preventing an attendant
// from viewing another branch's terminal health.
func (r *Repository) GetAttendantHealthSummary(
	ctx context.Context,
	tenantID string,
	attendantID string,
) (*TerminalHealthSummary, error) {

	var summary TerminalHealthSummary

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			COUNT(*) AS total_terminals,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND th.status = 'healthy'
			) AS healthy_terminals,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND th.status = 'attention'
			) AS attention_terminals,

			COUNT(*) FILTER (
				WHERE th.terminal_id IS NULL
				   OR th.last_seen_at < NOW() - INTERVAL '2 minutes'
			) AS offline_terminals,

			COUNT(*) FILTER (
				WHERE th.client_status = 'running'
				  AND th.last_seen_at >= NOW() - INTERVAL '2 minutes'
			) AS running_clients,

			COUNT(*) FILTER (
				WHERE th.client_status = 'stopped'
			) AS stopped_clients,

			COUNT(*) FILTER (
				WHERE th.client_status = 'error'
			) AS error_clients,

			COUNT(*) FILTER (
				WHERE th.client_status = 'unknown'
			) AS unknown_clients,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND NOT th.internet_available
			) AS internet_problems,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND NOT th.lan_available
			) AS lan_problems,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND th.cpu_usage_percent >= 95
			) AS high_cpu_usage,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND th.disk_total_bytes > 0
				  AND (
					th.disk_total_bytes - th.disk_free_bytes
				  )::NUMERIC / th.disk_total_bytes >= 0.95
			) AS low_disk_space,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND th.memory_total_bytes > 0
				  AND (
					th.memory_used_bytes::NUMERIC /
					th.memory_total_bytes
				  ) >= 0.95
			) AS high_memory_usage,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND (
					NOT th.security_available
					OR NOT th.antivirus_enabled
				  )
			) AS security_problems,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND NOT th.drivers_available
			) AS driver_problems,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND NOT th.sound_available
			) AS sound_problems,

			COUNT(*) FILTER (
				WHERE th.last_seen_at >= NOW() - INTERVAL '2 minutes'
				  AND NOT th.printer_available
			) AS printer_problems,

			COALESCE(
				SUM(th.sync_pending_count),
				0
			) AS sync_pending,

			COALESCE(
				SUM(th.sync_failed_count),
				0
			) AS sync_failed,

			MAX(th.last_seen_at) AS last_seen_at

		FROM attendant_branch_assignments aba

		INNER JOIN terminals t
			ON t.branch_id = aba.branch_id
		   AND t.tenant_id = aba.tenant_id

		LEFT JOIN terminal_health th
			ON th.terminal_id = t.id

		INNER JOIN users u
			ON u.id = aba.user_id
		   AND u.tenant_id = aba.tenant_id

		WHERE aba.tenant_id = $1
		  AND aba.user_id = $2
		  AND aba.unassigned_at IS NULL
		  AND u.role = 'attendant'
		  AND u.status = 'active'
		`,
		tenantID,
		attendantID,
	).Scan(
		&summary.TotalTerminals,

		&summary.HealthyTerminals,
		&summary.AttentionTerminals,
		&summary.OfflineTerminals,

		&summary.RunningClients,
		&summary.StoppedClients,
		&summary.ErrorClients,
		&summary.UnknownClients,

		&summary.InternetProblems,
		&summary.LANProblems,

		&summary.HighCPUUsage,
		&summary.LowDiskSpace,
		&summary.HighMemoryUsage,

		&summary.SecurityProblems,
		&summary.DriverProblems,
		&summary.SoundProblems,
		&summary.PrinterProblems,

		&summary.SyncPending,
		&summary.SyncFailed,

		&summary.LastSeenAt,
	)

	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// ============================================================
// ATTENDANT TERMINALS
// ============================================================

// ListAttendantTerminals returns the terminals belonging to the
// branch currently assigned to the authenticated attendant.
//
// The terminal UUID is retained for API requests, while the
// terminal_code is returned for human-friendly display.
func (r *Repository) ListAttendantTerminals(
	ctx context.Context,
	tenantID string,
	attendantID string,
) ([]AttendantTerminal, error) {

	const query = `
		SELECT DISTINCT
			t.id,
			t.terminal_code
		FROM attendant_branch_assignments aba
		INNER JOIN terminals t
			ON t.tenant_id = aba.tenant_id
			AND t.branch_id = aba.branch_id
		INNER JOIN users u
			ON u.id = aba.user_id
			AND u.tenant_id = aba.tenant_id
		WHERE aba.tenant_id = $1
		  AND aba.user_id = $2
		  AND aba.unassigned_at IS NULL
		  AND u.role = 'attendant'
		  AND u.status = 'active'
		ORDER BY t.terminal_code, t.id
	`

	rows, err := r.db.Query(
		ctx,
		query,
		tenantID,
		attendantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	terminals := make([]AttendantTerminal, 0)

	for rows.Next() {
		var terminal AttendantTerminal

		if err := rows.Scan(
			&terminal.TerminalID,
			&terminal.TerminalCode,
		); err != nil {
			return nil, err
		}

		terminals = append(terminals, terminal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return terminals, nil
}

// ============================================================
// ATTENDANT SINGLE TERMINAL HEALTH
// ============================================================

func (r *Repository) GetAttendantTerminalHealth(
	ctx context.Context,
	tenantID string,
	attendantID string,
	terminalID string,
) (*TerminalHealth, error) {

	const query = `
		SELECT
			th.terminal_id,
			th.status,

			th.issues,

			th.cpu_usage_percent,

			th.memory_available,
			th.memory_total_bytes,
			th.memory_used_bytes,
			th.memory_free_bytes,

			th.disk_available,
			th.disk_total_bytes,
			th.disk_free_bytes,

			th.network_adapter_available,
			th.lan_available,
			th.internet_available,
			th.dns_available,
			th.api_available,
			th.backend_latency_ms,
			th.network_issue,

			th.database_available,
			th.printer_available,

			th.sound_available,
			th.sound_issue,

			th.drivers_available,
			th.driver_issue,

			th.security_available,
			th.antivirus_enabled,
			th.security_issue,

			th.os_name,
			th.os_version,
			th.os_architecture,

			th.client_version,
			th.client_status,

			th.offline_queue_count,
			th.sync_pending_count,
			th.sync_failed_count,
			th.sync_status,
			th.last_sync_at,

			th.uptime_seconds,

			th.last_health_check,
			th.last_seen_at,
			th.created_at,
			th.updated_at

		FROM terminal_health th

		INNER JOIN terminals t
			ON t.id = th.terminal_id

		INNER JOIN attendant_branch_assignments aba
			ON aba.tenant_id = t.tenant_id
			AND aba.branch_id = t.branch_id

		INNER JOIN users u
			ON u.id = aba.user_id
			AND u.tenant_id = aba.tenant_id

		WHERE th.terminal_id = $1
		  AND t.tenant_id = $2
		  AND aba.user_id = $3
		  AND aba.unassigned_at IS NULL
		  AND u.role = 'attendant'
		  AND u.status = 'active'

		LIMIT 1
	`

	var health TerminalHealth
	var issuesJSON []byte

	err := r.db.QueryRow(
		ctx,
		query,
		terminalID,
		tenantID,
		attendantID,
	).Scan(
		&health.TerminalID,
		&health.Status,

		&issuesJSON,

		&health.CPUUsagePercent,

		&health.MemoryAvailable,
		&health.MemoryTotal,
		&health.MemoryUsed,
		&health.MemoryFree,

		&health.DiskAvailable,
		&health.DiskTotal,
		&health.DiskFree,

		&health.NetworkAdapterAvailable,
		&health.LANAvailable,
		&health.InternetAvailable,
		&health.DNSAvailable,
		&health.APIAvailable,
		&health.BackendLatencyMS,
		&health.NetworkIssue,

		&health.DatabaseAvailable,
		&health.PrinterAvailable,

		&health.SoundAvailable,
		&health.SoundIssue,

		&health.DriversAvailable,
		&health.DriverIssue,

		&health.SecurityAvailable,
		&health.AntivirusEnabled,
		&health.SecurityIssue,

		&health.OSName,
		&health.OSVersion,
		&health.OSArchitecture,

		&health.ClientVersion,
		&health.ClientStatus,

		&health.OfflineQueueCount,
		&health.SyncPendingCount,
		&health.SyncFailedCount,
		&health.SyncStatus,
		&health.LastSyncAt,

		&health.UptimeSeconds,

		&health.LastHealthCheck,
		&health.LastSeenAt,
		&health.CreatedAt,
		&health.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, err
	}

	if len(issuesJSON) == 0 {
		health.Issues = []string{}
	} else if err := json.Unmarshal(
		issuesJSON,
		&health.Issues,
	); err != nil {
		return nil, err
	}

	health.Status = HealthState(health)

	return &health, nil
}
