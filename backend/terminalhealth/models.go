package terminalhealth

import "time"

// TerminalHealth represents the latest health state reported
// by a registered CyberSaaS terminal.
type TerminalHealth struct {
	TerminalID string `json:"terminal_id"`

	// Overall health.
	Status        string   `json:"status"`
	HealthMessage string   `json:"health_message"`
	Issues        []string `json:"issues"`

	// CPU.
	CPUUsagePercent float64 `json:"cpu_usage_percent"`

	// Memory.
	MemoryAvailable bool  `json:"memory_available"`
	MemoryTotal     int64 `json:"memory_total_bytes"`
	MemoryUsed      int64 `json:"memory_used_bytes"`
	MemoryFree      int64 `json:"memory_free_bytes"`

	// Disk.
	DiskAvailable bool  `json:"disk_available"`
	DiskTotal     int64 `json:"disk_total_bytes"`
	DiskFree      int64 `json:"disk_free_bytes"`

	// Network.
	NetworkAdapterAvailable bool   `json:"network_adapter_available"`
	LANAvailable            bool   `json:"lan_available"`
	InternetAvailable       bool   `json:"internet_available"`
	DNSAvailable            bool   `json:"dns_available"`
	APIAvailable            bool   `json:"api_available"`
	BackendLatencyMS        int    `json:"backend_latency_ms"`
	NetworkIssue            string `json:"network_issue"`

	// Database.
	DatabaseAvailable bool `json:"database_available"`

	// Printer.
	PrinterAvailable bool `json:"printer_available"`

	// Sound.
	SoundAvailable bool   `json:"sound_available"`
	SoundIssue     string `json:"sound_issue"`

	// Drivers.
	DriversAvailable bool   `json:"drivers_available"`
	DriverIssue      string `json:"driver_issue"`

	// Windows Security / antivirus.
	SecurityAvailable bool   `json:"security_available"`
	AntivirusEnabled  bool   `json:"antivirus_enabled"`
	SecurityIssue     string `json:"security_issue"`

	// Operating system.
	OSName         string `json:"os_name"`
	OSVersion      string `json:"os_version"`
	OSArchitecture string `json:"os_architecture"`

	// CyberSaaS PC Client.
	ClientVersion string `json:"client_version"`
	ClientStatus  string `json:"client_status"`

	// Offline/synchronization.
	OfflineQueueCount int        `json:"offline_queue_count"`
	SyncPendingCount  int        `json:"sync_pending_count"`
	SyncFailedCount   int        `json:"sync_failed_count"`
	SyncStatus        string     `json:"sync_status"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`

	// Terminal runtime.
	UptimeSeconds int64 `json:"uptime_seconds"`

	// Timestamps.
	LastHealthCheck time.Time `json:"last_health_check"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ReportHealthRequest is sent by the C# terminal client.
//
// TerminalID is intentionally NOT accepted from the request body.
// The terminal identity comes from terminal authentication.
type ReportHealthRequest struct {
	Status        string   `json:"status"`
	HealthMessage string   `json:"health_message"`
	Issues        []string `json:"issues"`

	// CPU.
	CPUUsagePercent float64 `json:"cpu_usage_percent"`

	// Memory.
	MemoryAvailable bool  `json:"memory_available"`
	MemoryTotal     int64 `json:"memory_total_bytes"`
	MemoryUsed      int64 `json:"memory_used_bytes"`
	MemoryFree      int64 `json:"memory_free_bytes"`

	// Disk.
	DiskAvailable bool  `json:"disk_available"`
	DiskTotal     int64 `json:"disk_total_bytes"`
	DiskFree      int64 `json:"disk_free_bytes"`

	// Network.
	NetworkAdapterAvailable bool   `json:"network_adapter_available"`
	LANAvailable            bool   `json:"lan_available"`
	InternetAvailable       bool   `json:"internet_available"`
	DNSAvailable            bool   `json:"dns_available"`
	APIAvailable            bool   `json:"api_available"`
	BackendLatencyMS        int    `json:"backend_latency_ms"`
	NetworkIssue            string `json:"network_issue"`

	// Database.
	DatabaseAvailable bool `json:"database_available"`

	// Printer.
	PrinterAvailable bool `json:"printer_available"`

	// Sound.
	SoundAvailable bool   `json:"sound_available"`
	SoundIssue     string `json:"sound_issue"`

	// Drivers.
	DriversAvailable bool   `json:"drivers_available"`
	DriverIssue      string `json:"driver_issue"`

	// Windows Security / antivirus.
	SecurityAvailable bool   `json:"security_available"`
	AntivirusEnabled  bool   `json:"antivirus_enabled"`
	SecurityIssue     string `json:"security_issue"`

	// Operating system.
	OSName         string `json:"os_name"`
	OSVersion      string `json:"os_version"`
	OSArchitecture string `json:"os_architecture"`

	// PC Client.
	ClientVersion string `json:"client_version"`
	ClientStatus  string `json:"client_status"`

	// Synchronization.
	OfflineQueueCount int        `json:"offline_queue_count"`
	SyncPendingCount  int        `json:"sync_pending_count"`
	SyncFailedCount   int        `json:"sync_failed_count"`
	SyncStatus        string     `json:"sync_status"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`

	// Runtime.
	UptimeSeconds int64 `json:"uptime_seconds"`
}

// TerminalHealthResponse is returned after a health report
// has been successfully stored.
type TerminalHealthResponse struct {
	Health TerminalHealth `json:"health"`
}

// TerminalHealthListResponse contains terminal health information
// for terminals belonging to an authorized branch/tenant.
type TerminalHealthListResponse struct {
	Terminals []TerminalHealth `json:"terminals"`
	Total     int64            `json:"total"`
}

type TerminalHealthSummary struct {
	TotalTerminals int64 `json:"total_terminals"`

	HealthyTerminals   int64 `json:"healthy_terminals"`
	AttentionTerminals int64 `json:"attention_terminals"`
	OfflineTerminals   int64 `json:"offline_terminals"`

	RunningClients int64 `json:"running_clients"`
	StoppedClients int64 `json:"stopped_clients"`
	ErrorClients   int64 `json:"error_clients"`
	UnknownClients int64 `json:"unknown_clients"`

	InternetProblems int64 `json:"internet_problems"`
	LANProblems      int64 `json:"lan_problems"`

	HighCPUUsage    int64 `json:"high_cpu_usage"`
	LowDiskSpace    int64 `json:"low_disk_space"`
	HighMemoryUsage int64 `json:"high_memory_usage"`

	SecurityProblems int64 `json:"security_problems"`
	DriverProblems   int64 `json:"driver_problems"`
	SoundProblems    int64 `json:"sound_problems"`
	PrinterProblems  int64 `json:"printer_problems"`

	SyncPending int64 `json:"sync_pending"`
	SyncFailed  int64 `json:"sync_failed"`

	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

type AttendantTerminal struct {
	TerminalID   string `json:"terminal_id"`
	TerminalCode string `json:"terminal_code"`
}

type AttendantTerminalListResponse struct {
	Terminals []AttendantTerminal `json:"terminals"`
	Total     int64               `json:"total"`
}
