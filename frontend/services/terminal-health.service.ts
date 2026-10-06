import { api } from "@/lib/api";

export type TerminalHealth = {
  terminal_id: string;
  status: "healthy" | "attention" | "offline";
  health_message: string;
  issues: string[];

  cpu_usage_percent: number;

  memory_available: boolean;
  memory_total_bytes: number;
  memory_used_bytes: number;
  memory_free_bytes: number;

  disk_available: boolean;
  disk_total_bytes: number;
  disk_free_bytes: number;

  network_adapter_available: boolean;
  lan_available: boolean;
  internet_available: boolean;
  dns_available: boolean;
  api_available: boolean;
  database_available: boolean;

  printer_available: boolean;
  backend_latency_ms: number;
  network_issue: string;

  sound_available: boolean;
  sound_issue: string;

  drivers_available: boolean;
  driver_issue: string;

  security_available: boolean;
  antivirus_enabled: boolean;
  security_issue: string;

  os_name: string;
  os_version: string;
  os_architecture: string;

  client_version: string;
  client_status: "running" | "stopped" | "error" | "unknown";

  offline_queue_count: number;
  sync_pending_count: number;
  sync_failed_count: number;
  sync_status: "synced" | "pending" | "failed" | "offline";

  last_sync_at?: string | null;

  uptime_seconds: number;
  last_health_check: string;
  last_seen_at: string;

  created_at: string;
  updated_at: string;
};

export type AttendantTerminal = {
  terminal_id: string;
  terminal_code?: string;
  machine_name?: string;
  device_identifier?: string;
};

export type AttendantTerminalListResponse = {
  terminals: AttendantTerminal[];
  total: number;
};

export type TerminalHealthResponse = {
  health: TerminalHealth;
};

class TerminalHealthService {
  async listAttendantTerminals(): Promise<AttendantTerminalListResponse> {
    return api<AttendantTerminalListResponse>(
      "/api/attendant/health/terminals"
    );
  }

  async getAttendantTerminalHealth(
    terminalId: string
  ): Promise<TerminalHealthResponse> {
    return api<TerminalHealthResponse>(
      `/api/attendant/health/terminals/${terminalId}`
    );
  }
}

export const terminalHealthService =
  new TerminalHealthService();