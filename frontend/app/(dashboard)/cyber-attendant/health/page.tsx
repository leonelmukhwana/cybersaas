
"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AlertTriangle,
  ArrowDownUp,
  CheckCircle2,
  Clock3,
  Cpu,
  Database,
  HardDrive,
  Monitor,
  Printer,
  RefreshCw,
  ShieldCheck,
  Wifi,
  WifiOff,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/store/auth.store";
import { terminalHealthService } from "@/services/terminal-health.service";

type TerminalHealth = {
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
  backend_latency_ms: number;
  network_issue: string;

  database_available: boolean;
  printer_available: boolean;

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

type AttendantTerminal = {
  terminal_id: string;
  terminal_code?: string;
  machine_name?: string;
  device_identifier?: string;
};

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB", "TB"];

  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }

  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`;
}

function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return "Unknown";
  }

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) {
    return `${days}d ${hours}h`;
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }

  return `${minutes}m`;
}

function formatDate(value?: string | null): string {
  if (!value) {
    return "Never";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return date.toLocaleString();
}

function statusLabel(status: TerminalHealth["status"]): string {
  switch (status) {
    case "healthy":
      return "Healthy";
    case "attention":
      return "Attention";
    case "offline":
      return "Offline";
  }
}

function statusClasses(status: TerminalHealth["status"]): string {
  switch (status) {
    case "healthy":
      return "bg-green-50 text-green-700";
    case "attention":
      return "bg-amber-50 text-amber-700";
    case "offline":
      return "bg-red-50 text-red-700";
  }
}

function getNetworkLabel(health: TerminalHealth): {
  label: string;
  className: string;
} {
  if (!health.network_adapter_available) {
    return {
      label: "Adapter unavailable",
      className: "text-red-700",
    };
  }

  if (!health.lan_available) {
    return {
      label: "LAN problem",
      className: "text-red-700",
    };
  }

  if (!health.internet_available) {
    return {
      label: "Internet unavailable",
      className: "text-amber-700",
    };
  }

  if (!health.dns_available) {
    return {
      label: "DNS problem",
      className: "text-amber-700",
    };
  }

  if (!health.api_available) {
    return {
      label: "Backend unavailable",
      className: "text-amber-700",
    };
  }

  return {
    label: "Connected",
    className: "text-green-700",
  };
}

function getSyncLabel(health: TerminalHealth): string {
  switch (health.sync_status) {
    case "synced":
      return "Synced";

    case "pending":
      return `${health.sync_pending_count} pending`;

    case "failed":
      return `${health.sync_failed_count} failed`;

    case "offline":
      return "Offline";

    default:
      return "Unknown";
  }
}

export default function HealthPage() {
  const token = useAuthStore((state) => state.token);

  const [terminals, setTerminals] = useState<AttendantTerminal[]>([]);
  const [selectedTerminalId, setSelectedTerminalId] = useState("");

  const [health, setHealth] = useState<TerminalHealth | null>(null);

  const [loadingTerminals, setLoadingTerminals] = useState(true);
  const [loadingHealth, setLoadingHealth] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const [error, setError] = useState<string | null>(null);

  const selectedTerminal = useMemo(
    () =>
      terminals.find(
        (terminal) => terminal.terminal_id === selectedTerminalId
      ),
    [terminals, selectedTerminalId]
  );

  const loadTerminalHealth = useCallback(
    async (terminalId: string, refresh = false) => {
      if (!terminalId) {
        setHealth(null);
        return;
      }

      try {
        setLoadingHealth(true);

        if (refresh) {
          setRefreshing(true);
        }

        setError(null);

        const response =
          await terminalHealthService.getAttendantTerminalHealth(
            terminalId
          );

        setHealth(response.health);
      } catch (err) {
        setHealth(null);

        setError(
          err instanceof Error
            ? err.message
            : "Failed to load terminal health."
        );
      } finally {
        setLoadingHealth(false);
        setRefreshing(false);
      }
    },
    []
  );

  const loadTerminals = useCallback(
    async (refresh = false) => {
      if (!token) {
        setLoadingTerminals(false);
        return;
      }

      try {
        setLoadingTerminals(true);

        if (refresh) {
          setRefreshing(true);
        }

        setError(null);

        const response =
          await terminalHealthService.listAttendantTerminals();

        const terminalList = response.terminals ?? [];

        setTerminals(terminalList);

        if (terminalList.length === 0) {
          setSelectedTerminalId("");
          setHealth(null);
          return;
        }

        const currentStillExists = terminalList.some(
          (terminal) =>
            terminal.terminal_id === selectedTerminalId
        );

        const terminalId = currentStillExists
          ? selectedTerminalId
          : terminalList[0].terminal_id;

        setSelectedTerminalId(terminalId);

        await loadTerminalHealth(terminalId, false);
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "Failed to load terminals."
        );
      } finally {
        setLoadingTerminals(false);
        setRefreshing(false);
      }
    },
    [token, selectedTerminalId, loadTerminalHealth]
  );

  useEffect(() => {
    loadTerminals();
  }, [loadTerminals]);

  const handleTerminalChange = async (
    event: React.ChangeEvent<HTMLSelectElement>
  ) => {
    const terminalId = event.target.value;

    setSelectedTerminalId(terminalId);

    await loadTerminalHealth(terminalId);
  };

  const handleRefresh = async () => {
    if (!selectedTerminalId) {
      await loadTerminals(true);
      return;
    }

    await loadTerminalHealth(selectedTerminalId, true);
  };

  const memoryPercent =
    health && health.memory_total_bytes > 0
      ? Math.round(
          (health.memory_used_bytes /
            health.memory_total_bytes) *
            100
        )
      : 0;

  const diskUsedBytes =
    health && health.disk_total_bytes >= health.disk_free_bytes
      ? health.disk_total_bytes - health.disk_free_bytes
      : 0;

  const diskUsedPercent =
    health && health.disk_total_bytes > 0
      ? Math.round(
          (diskUsedBytes / health.disk_total_bytes) * 100
        )
      : 0;

  const network = health
    ? getNetworkLabel(health)
    : null;

  return (
    <DashboardShell role="attendant">
      <PageHeader
        title="Terminal Health"
        description="Monitor the health, connectivity, security, resources, and synchronization of your branch terminals."
        action={
          <Button
            type="button"
            variant="outline"
            onClick={handleRefresh}
            disabled={
              refreshing ||
              loadingTerminals ||
              loadingHealth
            }
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${
                refreshing ? "animate-spin" : ""
              }`}
            />
            Refresh
          </Button>
        }
      />

      {error && (
        <div className="mt-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* TERMINAL SELECTOR */}
      <div className="mt-6 rounded-xl border border-slate-200 bg-white p-5">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="text-base font-semibold text-slate-950">
              Select Terminal
            </h2>

            <p className="mt-1 text-sm text-slate-500">
              Select one terminal to view its complete health
              information.
            </p>
          </div>

          <div className="w-full sm:max-w-md">
            <label
              htmlFor="terminal-select"
              className="mb-2 block text-sm font-medium text-slate-700"
            >
              Terminal
            </label>

            <select
              id="terminal-select"
              value={selectedTerminalId}
              onChange={handleTerminalChange}
              disabled={
                loadingTerminals ||
                terminals.length === 0
              }
              className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none focus:border-[#0757B8] focus:ring-2 focus:ring-[#0757B8]/20"
            >
              {terminals.length === 0 ? (
                <option value="">
                  No terminals available
                </option>
              ) : (
                terminals.map((terminal) => (
                  <option
                    key={terminal.terminal_id}
                    value={terminal.terminal_id}
                  >
                    {terminal.terminal_code ||
                      terminal.machine_name ||
                      terminal.terminal_id}
                    {terminal.machine_name
                      ? ` — ${terminal.machine_name}`
                      : ""}
                  </option>
                ))
              )}
            </select>
          </div>
        </div>
      </div>

      {loadingTerminals ? (
        <div className="mt-6 flex min-h-64 items-center justify-center rounded-xl border border-slate-200 bg-white">
          <div className="flex items-center gap-2 text-sm text-slate-500">
            <RefreshCw className="h-4 w-4 animate-spin" />
            Loading terminals...
          </div>
        </div>
      ) : terminals.length === 0 ? (
        <div className="mt-6 flex min-h-64 flex-col items-center justify-center rounded-xl border border-slate-200 bg-white px-6 text-center">
          <Monitor className="h-10 w-10 text-slate-300" />

          <p className="mt-3 text-sm font-medium text-slate-700">
            No terminals available
          </p>

          <p className="mt-1 max-w-md text-sm text-slate-500">
            No terminals are currently available to this
            attendant.
          </p>
        </div>
      ) : loadingHealth ? (
        <div className="mt-6 flex min-h-64 items-center justify-center rounded-xl border border-slate-200 bg-white">
          <div className="flex items-center gap-2 text-sm text-slate-500">
            <RefreshCw className="h-4 w-4 animate-spin" />
            Loading terminal health...
          </div>
        </div>
      ) : !health ? (
        <div className="mt-6 flex min-h-64 flex-col items-center justify-center rounded-xl border border-slate-200 bg-white px-6 text-center">
          <AlertTriangle className="h-10 w-10 text-amber-400" />

          <p className="mt-3 text-sm font-medium text-slate-700">
            Health information unavailable
          </p>

          <p className="mt-1 max-w-md text-sm text-slate-500">
            This terminal has not provided health information
            yet.
          </p>
        </div>
      ) : (
        <>
          {/* SELECTED TERMINAL HEADER */}
          <div className="mt-6 rounded-xl border border-slate-200 bg-white p-5">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-blue-50">
                  <Monitor className="h-5 w-5 text-[#0757B8]" />
                </div>

                <div>
                  <p className="text-lg font-semibold text-slate-950">
                    {selectedTerminal?.terminal_code ||
                      selectedTerminal?.machine_name ||
                      health.terminal_id}
                  </p>

                  <p className="text-sm text-slate-500">
                    {selectedTerminal?.machine_name ||
                      "Terminal"}
                  </p>
                </div>
              </div>

              <span
                className={`inline-flex w-fit rounded-full px-3 py-1.5 text-sm font-medium ${statusClasses(
                  health.status
                )}`}
              >
                {statusLabel(health.status)}
              </span>
            </div>

            {health.health_message && (
              <div className="mt-4 rounded-lg bg-slate-50 px-4 py-3 text-sm text-slate-600">
                {health.health_message}
              </div>
            )}

            {health.issues.length > 0 && (
              <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 p-4">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-amber-600" />

                  <p className="text-sm font-semibold text-amber-800">
                    Issues detected
                  </p>
                </div>

                <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-amber-800">
                  {health.issues.map((issue, index) => (
                    <li key={`${issue}-${index}`}>
                      {issue}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          {/* RESOURCE HEALTH */}
          <div className="mt-6 grid gap-4 md:grid-cols-3">
            <HealthMetricCard
              title="CPU Usage"
              value={`${health.cpu_usage_percent.toFixed(1)}%`}
              icon={<Cpu className="h-5 w-5" />}
              warning={
                health.cpu_usage_percent >= 95
              }
            />

            <HealthMetricCard
              title="Memory"
              value={`${memoryPercent}% used`}
              detail={`${formatBytes(
                health.memory_free_bytes
              )} free`}
              icon={<Database className="h-5 w-5" />}
              warning={memoryPercent >= 95}
            />

            <HealthMetricCard
              title="Disk"
              value={`${diskUsedPercent}% used`}
              detail={`${formatBytes(
                health.disk_free_bytes
              )} free`}
              icon={<HardDrive className="h-5 w-5" />}
              warning={diskUsedPercent >= 95}
            />
          </div>

          {/* NETWORK */}
          <HealthSection
            title="Network & Connectivity"
            icon={<Wifi className="h-5 w-5" />}
          >
            <HealthRow
              label="Network Adapter"
              value={
                health.network_adapter_available
                  ? "Available"
                  : "Unavailable"
              }
              ok={health.network_adapter_available}
            />

            <HealthRow
              label="LAN"
              value={
                health.lan_available
                  ? "Connected"
                  : "Unavailable"
              }
              ok={health.lan_available}
            />

            <HealthRow
              label="Internet"
              value={
                health.internet_available
                  ? "Available"
                  : "Unavailable"
              }
              ok={health.internet_available}
            />

            <HealthRow
              label="DNS"
              value={
                health.dns_available
                  ? "Available"
                  : "Problem"
              }
              ok={health.dns_available}
            />

            <HealthRow
              label="Backend API"
              value={
                health.api_available
                  ? `Available${
                      health.backend_latency_ms > 0
                        ? ` · ${health.backend_latency_ms} ms`
                        : ""
                    }`
                  : "Unavailable"
              }
              ok={health.api_available}
            />

            {health.network_issue && (
              <HealthRow
                label="Network Issue"
                value={health.network_issue}
                ok={false}
              />
            )}
          </HealthSection>

          {/* DEVICES */}
          <div className="mt-6 grid gap-6 lg:grid-cols-2">
            <HealthSection
              title="Devices"
              icon={<Printer className="h-5 w-5" />}
            >
              <HealthRow
                label="Printer"
                value={
                  health.printer_available
                    ? "Available"
                    : "Problem"
                }
                ok={health.printer_available}
              />

              <HealthRow
                label="Sound"
                value={
                  health.sound_available
                    ? "Available"
                    : "Problem"
                }
                ok={health.sound_available}
              />

              {health.sound_issue && (
                <HealthRow
                  label="Sound Issue"
                  value={health.sound_issue}
                  ok={false}
                />
              )}

              <HealthRow
                label="Drivers"
                value={
                  health.drivers_available
                    ? "Available"
                    : "Problem"
                }
                ok={health.drivers_available}
              />

              {health.driver_issue && (
                <HealthRow
                  label="Driver Issue"
                  value={health.driver_issue}
                  ok={false}
                />
              )}
            </HealthSection>

            {/* SECURITY */}
            <HealthSection
              title="Security"
              icon={<ShieldCheck className="h-5 w-5" />}
            >
              <HealthRow
                label="Security"
                value={
                  health.security_available
                    ? "Available"
                    : "Problem"
                }
                ok={health.security_available}
              />

              <HealthRow
                label="Antivirus"
                value={
                  health.antivirus_enabled
                    ? "Enabled"
                    : "Disabled"
                }
                ok={health.antivirus_enabled}
              />

              {health.security_issue && (
                <HealthRow
                  label="Security Issue"
                  value={health.security_issue}
                  ok={false}
                />
              )}
            </HealthSection>
          </div>

          {/* CLIENT & SYNC */}
          <div className="mt-6 grid gap-6 lg:grid-cols-2">
            <HealthSection
              title="Cyber Client"
              icon={<Monitor className="h-5 w-5" />}
            >
              <HealthRow
                label="Status"
                value={
                  health.client_status
                    .charAt(0)
                    .toUpperCase() +
                  health.client_status.slice(1)
                }
                ok={
                  health.client_status === "running"
                }
              />

              <HealthRow
                label="Client Version"
                value={
                  health.client_version ||
                  "Unknown"
                }
                ok
              />

              <HealthRow
                label="Operating System"
                value={
                  health.os_name || "Unknown"
                }
                ok
              />

              <HealthRow
                label="OS Version"
                value={
                  health.os_version || "Unknown"
                }
                ok
              />

              <HealthRow
                label="Architecture"
                value={
                  health.os_architecture ||
                  "Unknown"
                }
                ok
              />
            </HealthSection>

            <HealthSection
              title="Synchronization"
              icon={<ArrowDownUp className="h-5 w-5" />}
            >
              <HealthRow
                label="Sync Status"
                value={getSyncLabel(health)}
                ok={
                  health.sync_status === "synced"
                }
              />

              <HealthRow
                label="Offline Queue"
                value={String(
                  health.offline_queue_count
                )}
                ok={
                  health.offline_queue_count === 0
                }
              />

              <HealthRow
                label="Pending"
                value={String(
                  health.sync_pending_count
                )}
                ok={
                  health.sync_pending_count === 0
                }
              />

              <HealthRow
                label="Failed"
                value={String(
                  health.sync_failed_count
                )}
                ok={
                  health.sync_failed_count === 0
                }
              />

              <HealthRow
                label="Last Sync"
                value={formatDate(
                  health.last_sync_at
                )}
                ok
              />
            </HealthSection>
          </div>

          {/* SYSTEM */}
          <div className="mt-6 grid gap-4 md:grid-cols-3">
            <HealthMetricCard
              title="Uptime"
              value={formatUptime(
                health.uptime_seconds
              )}
              icon={<Clock3 className="h-5 w-5" />}
            />

            <HealthMetricCard
              title="Last Health Check"
              value={formatDate(
                health.last_health_check
              )}
              icon={<CheckCircle2 className="h-5 w-5" />}
            />

            <HealthMetricCard
              title="Last Seen"
              value={formatDate(
                health.last_seen_at
              )}
              icon={
                health.status === "offline" ? (
                  <WifiOff className="h-5 w-5" />
                ) : (
                  <Wifi className="h-5 w-5" />
                )
              }
              warning={health.status === "offline"}
            />
          </div>
        </>
      )}
    </DashboardShell>
  );
}

function HealthMetricCard({
  title,
  value,
  detail,
  icon,
  warning = false,
}: {
  title: string;
  value: string;
  detail?: string;
  icon: React.ReactNode;
  warning?: boolean;
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-slate-500">
            {title}
          </p>

          <p
            className={`mt-1 text-xl font-bold ${
              warning
                ? "text-amber-700"
                : "text-slate-950"
            }`}
          >
            {value}
          </p>

          {detail && (
            <p className="mt-1 text-xs text-slate-500">
              {detail}
            </p>
          )}
        </div>

        <div
          className={`flex h-10 w-10 items-center justify-center rounded-lg ${
            warning
              ? "bg-amber-50 text-amber-600"
              : "bg-slate-100 text-slate-600"
          }`}
        >
          {icon}
        </div>
      </div>
    </div>
  );
}

function HealthSection({
  title,
  icon,
  children,
}: {
  title: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <div className="flex items-center gap-2">
        <div className="text-slate-600">
          {icon}
        </div>

        <h2 className="font-semibold text-slate-950">
          {title}
        </h2>
      </div>

      <div className="mt-4 divide-y divide-slate-100">
        {children}
      </div>
    </div>
  );
}

function HealthRow({
  label,
  value,
  ok,
}: {
  label: string;
  value: string;
  ok: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-3">
      <span className="text-sm text-slate-600">
        {label}
      </span>

      <span
        className={`flex items-center gap-1.5 text-right text-sm font-medium ${
          ok
            ? "text-green-700"
            : "text-red-700"
        }`}
      >
        {ok ? (
          <CheckCircle2 className="h-4 w-4" />
        ) : (
          <AlertTriangle className="h-4 w-4" />
        )}

        {value}
      </span>
    </div>
  );
}
