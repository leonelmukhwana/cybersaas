
"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Lock,
  Monitor,
  MoreHorizontal,
  RefreshCw,
  Search,
  Unlock,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

import { useTerminalStore } from "@/store/terminal.store";

import type { Terminal } from "@/types/terminal";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

export default function TerminalsPage() {
  const {
    terminals,
    total,
    loading,
    saving,
    error,
    fetchAttendantTerminals,
    changeAttendantLockState,
    clearError,
  } = useTerminalStore();

  const [search, setSearch] = useState("");
  const [showActions, setShowActions] =
    useState<string | null>(null);

  useEffect(() => {
    void fetchAttendantTerminals();
  }, [fetchAttendantTerminals]);

  const filteredTerminals = useMemo(() => {
    const value = search
      .trim()
      .toLowerCase();

    if (!value) {
      return terminals;
    }

    return terminals.filter(
      (terminal) =>
        terminal.terminal_code
          .toLowerCase()
          .includes(value) ||
        (terminal.machine_name ?? "")
          .toLowerCase()
          .includes(value) ||
        (terminal.branch_name ?? "")
          .toLowerCase()
          .includes(value)
    );
  }, [terminals, search]);

  const onlineCount = terminals.filter(
    (terminal) =>
      getOnlineState(terminal) === "online"
  ).length;

  const offlineCount = terminals.filter(
    (terminal) =>
      getOnlineState(terminal) === "offline"
  ).length;

  const activeCount = terminals.filter(
    (terminal) =>
      terminal.status === "active"
  ).length;

  const handleLockChange = async (
    terminal: Terminal,
    state: "locked" | "unlocked"
  ) => {
    try {
      await changeAttendantLockState(
        terminal.id,
        { state }
      );

      setShowActions(null);
    } catch {
      // Store handles the error.
    }
  };

  const formatLastSeen = (
    value?: string | null
  ) => {
    if (!value) {
      return "Never";
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
      return "Unknown";
    }

    return date.toLocaleString();
  };

  return (
    <DashboardShell role="attendant">
      <div className="space-y-6">
        <PageHeader
          title="Terminals"
          description="View and control the computers assigned to your branch."
        />

        {/* Error */}

        {error && (
          <Card>
            <CardContent className="flex items-center justify-between gap-4 p-4">
              <p className="text-sm text-destructive">
                {error}
              </p>

              <Button
                variant="ghost"
                size="sm"
                onClick={clearError}
              >
                Dismiss
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Summary */}

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center gap-3">
                <div className="rounded-lg bg-blue-50 p-2 text-blue-600">
                  <Monitor className="h-5 w-5" />
                </div>

                <div>
                  <p className="text-sm text-muted-foreground">
                    My branch terminals
                  </p>

                  <p className="text-2xl font-semibold">
                    {total}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5">
              <p className="text-sm text-muted-foreground">
                Online
              </p>

              <p className="mt-1 text-2xl font-semibold">
                {onlineCount}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5">
              <p className="text-sm text-muted-foreground">
                Offline
              </p>

              <p className="mt-1 text-2xl font-semibold">
                {offlineCount}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5">
              <p className="text-sm text-muted-foreground">
                Active
              </p>

              <p className="mt-1 text-2xl font-semibold">
                {activeCount}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Search */}

        <Card>
          <CardContent className="p-5">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="relative w-full sm:max-w-xl">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />

                <Input
                  value={search}
                  onChange={(event) =>
                    setSearch(
                      event.target.value
                    )
                  }
                  placeholder="Search terminals, machine names or branches..."
                  className="pl-9"
                />
              </div>

              <Button
                variant="outline"
                onClick={() =>
                  void fetchAttendantTerminals()
                }
                disabled={loading}
              >
                <RefreshCw
                  className={`mr-2 h-4 w-4 ${
                    loading
                      ? "animate-spin"
                      : ""
                  }`}
                />

                Refresh
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Terminal table */}

        <Card>
          <CardContent className="p-0">
            {loading &&
            terminals.length === 0 ? (
              <div className="p-8 text-center text-sm text-muted-foreground">
                Loading terminals...
              </div>
            ) : filteredTerminals.length ===
              0 ? (
              <div className="p-10 text-center">
                <Monitor className="mx-auto mb-3 h-10 w-10 text-muted-foreground" />

                <p className="font-medium">
                  No terminals found
                </p>

                <p className="mt-1 text-sm text-muted-foreground">
                  Terminals assigned to your branch will appear here.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/40">
                      <th className="px-5 py-3 text-left font-medium">
                        Terminal
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Branch
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Status
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Connection
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Last seen
                      </th>

                      <th className="px-5 py-3 text-right font-medium">
                        Actions
                      </th>
                    </tr>
                  </thead>

                  <tbody>
                    {filteredTerminals.map(
                      (terminal) => {
                        const online =
                          getOnlineState(
                            terminal
                          ) === "online";

                        return (
                          <tr
                            key={terminal.id}
                            className="border-b last:border-0"
                          >
                            <td className="px-5 py-4">
                              <div className="flex items-center gap-3">
                                <div className="rounded-lg bg-muted p-2">
                                  <Monitor className="h-4 w-4" />
                                </div>

                                <div>
                                  <p className="font-medium">
                                    {
                                      terminal.terminal_code
                                    }
                                  </p>

                                  <p className="text-xs text-muted-foreground">
                                    {terminal.machine_name ??
                                      "Unnamed machine"}
                                  </p>
                                </div>
                              </div>
                            </td>

                            <td className="px-5 py-4">
                              {terminal.branch_name ??
                                terminal.branch_id}
                            </td>

                            <td className="px-5 py-4">
                              <span className="rounded-full bg-muted px-2.5 py-1 text-xs font-medium capitalize">
                                {terminal.status}
                              </span>
                            </td>

                            <td className="px-5 py-4">
                              <div className="flex items-center gap-2">
                                <span
                                  className={`h-2.5 w-2.5 rounded-full ${
                                    online
                                      ? "bg-green-500"
                                      : "bg-gray-400"
                                  }`}
                                />

                                <span>
                                  {online
                                    ? "Online"
                                    : "Offline"}
                                </span>
                              </div>
                            </td>

                            <td className="px-5 py-4 text-muted-foreground">
                              {formatLastSeen(
                                terminal.last_seen_at
                              )}
                            </td>

                            <td className="relative px-5 py-4 text-right">
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() =>
                                  setShowActions(
                                    showActions ===
                                      terminal.id
                                      ? null
                                      : terminal.id
                                  )
                                }
                              >
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>

                              {showActions ===
                                terminal.id && (
                                <div className="absolute right-5 top-12 z-20 w-52 rounded-lg border bg-background p-1 text-left shadow-lg">
                                  <button
                                    className="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted disabled:cursor-not-allowed disabled:opacity-50"
                                    onClick={() =>
                                      void handleLockChange(
                                        terminal,
                                        "locked"
                                      )
                                    }
                                    disabled={saving}
                                  >
                                    <Lock className="h-4 w-4" />
                                    Lock terminal
                                  </button>

                                  <button
                                    className="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted disabled:cursor-not-allowed disabled:opacity-50"
                                    onClick={() =>
                                      void handleLockChange(
                                        terminal,
                                        "unlocked"
                                      )
                                    }
                                    disabled={saving}
                                  >
                                    <Unlock className="h-4 w-4" />
                                    Unlock terminal
                                  </button>
                                </div>
                              )}
                            </td>
                          </tr>
                        );
                      }
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>

        <p className="text-sm text-muted-foreground">
          Terminal registration and configuration are managed by the Cyber Owner.
        </p>
      </div>
    </DashboardShell>
  );
}

function getOnlineState(
  terminal: Terminal
): "online" | "offline" {
  if (!terminal.last_seen_at) {
    return "offline";
  }

  const lastSeen =
    new Date(
      terminal.last_seen_at
    ).getTime();

  if (Number.isNaN(lastSeen)) {
    return "offline";
  }

  return Date.now() - lastSeen <= 90_000
    ? "online"
    : "offline";
}
