
"use client";

import { useEffect, useMemo, useState } from "react";
import {
  CalendarClock,
  Eye,
  Lock,
  Monitor,
  Pause,
  Play,
  Power,
  RefreshCw,
  Search,
  Unlock,
  UserRound,
  XCircle,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import { useAuthStore } from "@/store/auth.store";
import { useSessionStore } from "@/store/session.store";
import { useTerminalStore } from "@/store/terminal.store";

import type {
  Session,
  SessionStatus,
} from "@/types/session";

export default function SessionsPage() {
  const {
    isAuthenticated,
    isOffline,
  } = useAuthStore();

  const {
    sessions,
    total,
    loading: sessionLoading,
    saving: sessionSaving,
    error: sessionError,
    fetchSessions,
    getSession,
    pauseSession,
    resumeSession,
    cancelSession,
    endSession,
    clearError: clearSessionError,
  } = useSessionStore();

  const {
    saving: terminalSaving,
    error: terminalError,
    changeAttendantLockState,
    restartAttendantTerminal,
    shutdownAttendantTerminal,
    clearError: clearTerminalError,
  } = useTerminalStore();

  const [search, setSearch] = useState("");

  const [statusFilter, setStatusFilter] =
    useState<SessionStatus | "all">("all");

  const [selectedSession, setSelectedSession] =
    useState<Session | null>(null);

  const [showDetails, setShowDetails] =
    useState(false);

  const [cancelReason, setCancelReason] =
    useState("");

  const [showCancelForm, setShowCancelForm] =
    useState(false);

  const loading =
    sessionLoading || terminalSaving;

  const saving =
    sessionSaving || terminalSaving;

  const error =
    sessionError || terminalError;

  /*
   * =========================================================
   * LOAD SESSIONS
   * =========================================================
   *
   * Sessions are ONLINE ONLY.
   */
  useEffect(() => {
    if (
      !isAuthenticated ||
      isOffline
    ) {
      return;
    }

    void fetchSessions();
  }, [
    isAuthenticated,
    isOffline,
    fetchSessions,
  ]);

  /*
   * =========================================================
   * FILTER SESSIONS
   * =========================================================
   *
   * Search uses safe display fields instead of UUIDs.
   */
  const filteredSessions = useMemo(() => {
    const value =
      search.trim().toLowerCase();

    return sessions.filter(
      (session) => {
        const matchesSearch =
          !value ||
          session.customer_name
            .toLowerCase()
            .includes(value) ||
          (session.parent_name
            ?.toLowerCase()
            .includes(value) ?? false) ||
          (session.terminal_name
            ?.toLowerCase()
            .includes(value) ?? false);

        const matchesStatus =
          statusFilter === "all" ||
          session.status === statusFilter;

        return (
          matchesSearch &&
          matchesStatus
        );
      }
    );
  }, [
    sessions,
    search,
    statusFilter,
  ]);

  /*
   * =========================================================
   * SESSION COUNTS
   * =========================================================
   */

  const activeCount =
    sessions.filter(
      (session) =>
        session.status === "active"
    ).length;

  const pausedCount =
    sessions.filter(
      (session) =>
        session.status === "paused"
    ).length;

  const completedCount =
    sessions.filter(
      (session) =>
        session.status === "completed"
    ).length;

  /*
   * =========================================================
   * OPEN SESSION DETAILS
   * =========================================================
   */

  const openDetails = async (
    session: Session
  ) => {
    if (isOffline) {
      return;
    }

    try {
      const result =
        await getSession(
          session.id
        );

      setSelectedSession(result);
      setShowDetails(true);
      setShowCancelForm(false);
      setCancelReason("");
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * REFRESH SELECTED SESSION
   * =========================================================
   */

  const refreshSelectedSession =
    async () => {
      if (
        isOffline ||
        !selectedSession
      ) {
        return;
      }

      try {
        const updated =
          await getSession(
            selectedSession.id
          );

        setSelectedSession(
          updated
        );
      } catch {
        // Error is handled by the store.
      }
    };

  /*
   * =========================================================
   * TERMINAL LOCK / UNLOCK
   * =========================================================
   */

  const handleLockState = async (
    session: Session,
    state:
      | "locked"
      | "unlocked"
  ) => {
    try {
      await changeAttendantLockState(
        session.terminal_id,
        { state }
      );
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * RESTART TERMINAL
   * =========================================================
   */

  const handleRestart = async (
    session: Session
  ) => {
    const confirmed =
      window.confirm(
        `Restart ${session.terminal_name ?? "this computer"}?`
      );

    if (!confirmed) {
      return;
    }

    try {
      await restartAttendantTerminal(
        session.terminal_id
      );
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * SHUTDOWN TERMINAL
   * =========================================================
   */

  const handleShutdown = async (
    session: Session
  ) => {
    const confirmed =
      window.confirm(
        `Shut down ${session.terminal_name ?? "this computer"}?`
      );

    if (!confirmed) {
      return;
    }

    try {
      await shutdownAttendantTerminal(
        session.terminal_id
      );
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * PAUSE SESSION
   * =========================================================
   */

  const handlePause = async (
    session: Session
  ) => {
    if (isOffline) {
      return;
    }

    try {
      const updated =
        await pauseSession(
          session.id
        );

      setSelectedSession(
        updated
      );
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * RESUME SESSION
   * =========================================================
   */

  const handleResume = async (
    session: Session
  ) => {
    if (isOffline) {
      return;
    }

    try {
      const updated =
        await resumeSession(
          session.id
        );

      setSelectedSession(
        updated
      );
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * END SESSION
   * =========================================================
   */

  const handleEnd = async (
    session: Session
  ) => {
    if (isOffline) {
      return;
    }

    const confirmed =
      window.confirm(
        "End this session?"
      );

    if (!confirmed) {
      return;
    }

    try {
      const completed =
        await endSession(
          session.id
        );

      setSelectedSession(
        completed
      );
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * CANCEL SESSION
   * =========================================================
   */

  const handleCancel = async (
    session: Session
  ) => {
    if (isOffline) {
      return;
    }

    const reason =
      cancelReason.trim();

    if (!reason) {
      return;
    }

    try {
      await cancelSession(
        session.id,
        { reason }
      );

      setShowCancelForm(false);
      setCancelReason("");

      await refreshSelectedSession();
    } catch {
      // Error is handled by the store.
    }
  };

  /*
   * =========================================================
   * REFRESH SESSIONS
   * =========================================================
   */

  const handleRefresh = () => {
    if (
      !isAuthenticated ||
      isOffline
    ) {
      return;
    }

    void fetchSessions();
  };

  /*
   * =========================================================
   * FORMAT DATE
   * =========================================================
   */

  const formatDate = (
    value?: string | null
  ) => {
    if (!value) {
      return "—";
    }

    const date =
      new Date(value);

    if (
      Number.isNaN(
        date.getTime()
      )
    ) {
      return "—";
    }

    return date.toLocaleString();
  };

  /*
   * =========================================================
   * FORMAT MONEY
   * =========================================================
   */

  const formatMoney = (
    value?: string | null
  ) => {
    if (
      value === undefined ||
      value === null ||
      value === ""
    ) {
      return "—";
    }

    return `KSh ${value}`;
  };

  /*
   * =========================================================
   * STATUS STYLE
   * =========================================================
   */

  const statusClass = (
    status: SessionStatus
  ) => {
    switch (status) {
      case "active":
        return "bg-green-50 text-green-700";

      case "paused":
        return "bg-yellow-50 text-yellow-700";

      case "completed":
        return "bg-blue-50 text-blue-700";

      case "cancelled":
        return "bg-red-50 text-red-700";

      default:
        return "bg-muted text-muted-foreground";
    }
  };

  /*
   * =========================================================
   * TERMINAL OPERATIONS
   * =========================================================
   */

  const canOperateTerminal = (
    session: Session
  ) =>
    session.status === "active" ||
    session.status === "paused";

  return (
    <DashboardShell role="attendant">
      <div className="space-y-6">
        <PageHeader
          title="Sessions"
          description="View and operate customer sessions across your assigned branch."
        />

        {isOffline && (
          <Card>
            <CardContent className="p-4">
              <p className="text-sm text-muted-foreground">
                Sessions are available only while the
                attendant is online. Reconnect to the
                internet to view and manage sessions.
              </p>
            </CardContent>
          </Card>
        )}

        {error && (
          <Card>
            <CardContent className="flex items-center justify-between gap-4 p-4">
              <p className="text-sm text-destructive">
                {error}
              </p>

              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  clearSessionError();
                  clearTerminalError();
                }}
              >
                Dismiss
              </Button>
            </CardContent>
          </Card>
        )}

        {/* =====================================================
            SUMMARY CARDS
        ====================================================== */}

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardContent className="p-5">
              <div className="flex items-center gap-3">
                <div className="rounded-lg bg-blue-50 p-2 text-blue-600">
                  <CalendarClock className="h-5 w-5" />
                </div>

                <div>
                  <p className="text-sm text-muted-foreground">
                    Total sessions
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
                Active
              </p>

              <p className="mt-1 text-2xl font-semibold">
                {activeCount}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5">
              <p className="text-sm text-muted-foreground">
                Paused
              </p>

              <p className="mt-1 text-2xl font-semibold">
                {pausedCount}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5">
              <p className="text-sm text-muted-foreground">
                Completed
              </p>

              <p className="mt-1 text-2xl font-semibold">
                {completedCount}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* =====================================================
            SEARCH + FILTERS
        ====================================================== */}

        <Card>
          <CardContent className="p-5">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div className="relative w-full lg:max-w-xl">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />

                <Input
                  value={search}
                  onChange={(event) =>
                    setSearch(
                      event.target.value
                    )
                  }
                  placeholder="Search by customer, parent or computer..."
                  className="pl-9"
                />
              </div>

              <div className="flex flex-wrap gap-2">
                {(
                  [
                    "all",
                    "active",
                    "paused",
                    "completed",
                    "cancelled",
                  ] as const
                ).map((status) => (
                  <Button
                    key={status}
                    variant={
                      statusFilter === status
                        ? "default"
                        : "outline"
                    }
                    size="sm"
                    onClick={() =>
                      setStatusFilter(
                        status
                      )
                    }
                  >
                    {status === "all"
                      ? "All"
                      : status
                          .charAt(0)
                          .toUpperCase() +
                        status.slice(1)}
                  </Button>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* =====================================================
            SESSION TABLE
        ====================================================== */}

        <Card>
          <CardContent className="p-0">
            {sessionLoading &&
            sessions.length === 0 ? (
              <div className="p-8 text-center text-sm text-muted-foreground">
                Loading sessions...
              </div>
            ) : filteredSessions.length ===
              0 ? (
              <div className="p-10 text-center">
                <CalendarClock className="mx-auto mb-3 h-10 w-10 text-muted-foreground" />

                <p className="font-medium">
                  {isOffline
                    ? "Sessions unavailable offline"
                    : "No sessions found"}
                </p>

                <p className="mt-1 text-sm text-muted-foreground">
                  {isOffline
                    ? "Reconnect to the internet to view sessions from the backend."
                    : "Sessions will appear here when customers use terminals."}
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b bg-muted/40">
                      <th className="px-5 py-3 text-left font-medium">
                        Session
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Customer
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Computer
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Type
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Status
                      </th>

                      <th className="px-5 py-3 text-left font-medium">
                        Started
                      </th>

                      <th className="px-5 py-3 text-right font-medium">
                        Action
                      </th>
                    </tr>
                  </thead>

                  <tbody>
                    {filteredSessions.map(
                      (session) => (
                        <tr
                          key={session.id}
                          className="border-b last:border-0"
                        >
                          <td className="px-5 py-4">
                            <div className="flex items-center gap-3">
                              <div className="rounded-lg bg-muted p-2">
                                <CalendarClock className="h-4 w-4" />
                              </div>

                              <div>
                                <p className="font-medium">
                                  {session.id.slice(
                                    0,
                                    8
                                  )}
                                </p>

                                <p className="text-xs text-muted-foreground">
                                  Session
                                </p>
                              </div>
                            </div>
                          </td>

                          <td className="px-5 py-4">
                            <div className="flex items-start gap-2">
                              <UserRound className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />

                              <div>
                                <p className="font-medium">
                                  {session.customer_name}
                                </p>

                                {session.parent_name && (
                                  <p className="text-xs text-muted-foreground">
                                    Parent:{" "}
                                    {session.parent_name}
                                  </p>
                                )}

                                <p className="text-xs capitalize text-muted-foreground">
                                  {session.customer_type.replace(
                                    "_",
                                    " "
                                  )}
                                </p>
                              </div>
                            </div>
                          </td>

                          <td className="px-5 py-4">
                            <div className="flex items-center gap-2">
                              <Monitor className="h-4 w-4 text-muted-foreground" />

                              <span className="font-medium">
                                {session.terminal_name ??
                                  "Unnamed computer"}
                              </span>
                            </div>
                          </td>

                          <td className="px-5 py-4 capitalize">
                            {session.session_type.replace(
                              "_",
                              " "
                            )}
                          </td>

                          <td className="px-5 py-4">
                            <span
                              className={`rounded-full px-2.5 py-1 text-xs font-medium capitalize ${statusClass(
                                session.status
                              )}`}
                            >
                              {session.status}
                            </span>
                          </td>

                          <td className="px-5 py-4 text-muted-foreground">
                            {formatDate(
                              session.started_at
                            )}
                          </td>

                          <td className="px-5 py-4 text-right">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                void openDetails(
                                  session
                                )
                              }
                              disabled={
                                loading ||
                                isOffline
                              }
                            >
                              <Eye className="mr-2 h-4 w-4" />
                              Manage
                            </Button>
                          </td>
                        </tr>
                      )
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>

        {/* =====================================================
            REFRESH
        ====================================================== */}

        <div className="flex justify-end">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={
              loading ||
              !isAuthenticated ||
              isOffline
            }
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${
                sessionLoading
                  ? "animate-spin"
                  : ""
              }`}
            />

            Refresh sessions
          </Button>
        </div>

        {/* =====================================================
            SESSION DETAILS MODAL
        ====================================================== */}

        {showDetails &&
          selectedSession && (
            <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
              <Card className="max-h-[90vh] w-full max-w-2xl overflow-y-auto">
                <CardContent className="space-y-6 p-6">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <h2 className="text-lg font-semibold">
                        Manage session
                      </h2>

                      <p className="mt-1 text-sm text-muted-foreground">
                        View session details and operate the assigned computer.
                      </p>
                    </div>

                    <span
                      className={`rounded-full px-2.5 py-1 text-xs font-medium capitalize ${statusClass(
                        selectedSession.status
                      )}`}
                    >
                      {selectedSession.status}
                    </span>
                  </div>

                  {/* =================================================
                      SESSION INFORMATION
                  ================================================== */}

                  <div className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
                    <div>
                      <p className="text-xs text-muted-foreground">
                        Session ID
                      </p>

                      <p className="mt-1 break-all text-sm font-medium">
                        {selectedSession.id}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Session type
                      </p>

                      <p className="mt-1 text-sm font-medium capitalize">
                        {selectedSession.session_type.replace(
                          "_",
                          " "
                        )}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Customer
                      </p>

                      <p className="mt-1 text-sm font-medium">
                        {selectedSession.customer_name}
                      </p>

                      {selectedSession.parent_name && (
                        <p className="mt-1 text-xs text-muted-foreground">
                          Parent:{" "}
                          {selectedSession.parent_name}
                        </p>
                      )}

                      <p className="mt-1 text-xs capitalize text-muted-foreground">
                        {selectedSession.customer_type.replace(
                          "_",
                          " "
                        )}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Computer
                      </p>

                      <p className="mt-1 text-sm font-medium">
                        {selectedSession.terminal_name ??
                          "Unnamed computer"}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Started
                      </p>

                      <p className="mt-1 text-sm">
                        {formatDate(
                          selectedSession.started_at
                        )}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Ended
                      </p>

                      <p className="mt-1 text-sm">
                        {formatDate(
                          selectedSession.ended_at
                        )}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Rate per minute
                      </p>

                      <p className="mt-1 text-sm">
                        {formatMoney(
                          selectedSession.rate_per_minute
                        )}
                      </p>
                    </div>

                    <div>
                      <p className="text-xs text-muted-foreground">
                        Final amount
                      </p>

                      <p className="mt-1 text-sm font-semibold">
                        {formatMoney(
                          selectedSession.final_amount
                        )}
                      </p>
                    </div>
                  </div>

                  {/* =================================================
                      COMPUTER CONTROLS
                  ================================================== */}

                  {canOperateTerminal(
                    selectedSession
                  ) && (
                    <div className="space-y-3">
                      <div>
                        <h3 className="text-sm font-semibold">
                          Computer controls
                        </h3>

                        <p className="mt-1 text-xs text-muted-foreground">
                          These controls operate the assigned computer.
                        </p>
                      </div>

                      <div className="grid gap-2 sm:grid-cols-2">
                        <Button
                          variant="outline"
                          onClick={() =>
                            void handleLockState(
                              selectedSession,
                              "locked"
                            )
                          }
                          disabled={saving}
                        >
                          <Lock className="mr-2 h-4 w-4" />
                          Lock computer
                        </Button>

                        <Button
                          variant="outline"
                          onClick={() =>
                            void handleLockState(
                              selectedSession,
                              "unlocked"
                            )
                          }
                          disabled={saving}
                        >
                          <Unlock className="mr-2 h-4 w-4" />
                          Unlock computer
                        </Button>

                        <Button
                          variant="outline"
                          onClick={() =>
                            void handleRestart(
                              selectedSession
                            )
                          }
                          disabled={saving}
                        >
                          <RefreshCw className="mr-2 h-4 w-4" />
                          Restart computer
                        </Button>

                        <Button
                          variant="outline"
                          onClick={() =>
                            void handleShutdown(
                              selectedSession
                            )
                          }
                          disabled={saving}
                        >
                          <Power className="mr-2 h-4 w-4" />
                          Shutdown computer
                        </Button>
                      </div>
                    </div>
                  )}

                  {/* =================================================
                      ACTIVE SESSION CONTROLS
                  ================================================== */}

                  {selectedSession.status ===
                    "active" && (
                    <div className="space-y-3">
                      <div>
                        <h3 className="text-sm font-semibold">
                          Session controls
                        </h3>

                        <p className="mt-1 text-xs text-muted-foreground">
                          Pause, end, or cancel the current session.
                        </p>
                      </div>

                      <div className="grid gap-2 sm:grid-cols-3">
                        <Button
                          variant="outline"
                          onClick={() =>
                            void handlePause(
                              selectedSession
                            )
                          }
                          disabled={saving}
                        >
                          <Pause className="mr-2 h-4 w-4" />
                          Pause
                        </Button>

                        <Button
                          onClick={() =>
                            void handleEnd(
                              selectedSession
                            )
                          }
                          disabled={saving}
                        >
                          <XCircle className="mr-2 h-4 w-4" />
                          End session
                        </Button>

                        <Button
                          variant="destructive"
                          onClick={() =>
                            setShowCancelForm(
                              true
                            )
                          }
                          disabled={saving}
                        >
                          <XCircle className="mr-2 h-4 w-4" />
                          Cancel session
                        </Button>
                      </div>
                    </div>
                  )}

                  {/* =================================================
                      PAUSED SESSION CONTROLS
                  ================================================== */}

                  {selectedSession.status ===
                    "paused" && (
                    <div className="space-y-3">
                      <div>
                        <h3 className="text-sm font-semibold">
                          Session controls
                        </h3>

                        <p className="mt-1 text-xs text-muted-foreground">
                          The session is paused and billing is stopped.
                        </p>
                      </div>

                      <Button
                        onClick={() =>
                          void handleResume(
                            selectedSession
                          )
                        }
                        disabled={saving}
                      >
                        <Play className="mr-2 h-4 w-4" />
                        Resume session
                      </Button>
                    </div>
                  )}

                  {/* =================================================
                      CANCEL FORM
                  ================================================== */}

                  {showCancelForm &&
                    selectedSession.status ===
                      "active" && (
                      <div className="space-y-3 rounded-lg border border-red-200 bg-red-50/50 p-4">
                        <div>
                          <h3 className="text-sm font-semibold text-red-800">
                            Cancel session
                          </h3>

                          <p className="mt-1 text-xs text-red-700">
                            Enter the reason for cancelling this session.
                          </p>
                        </div>

                        <Input
                          value={cancelReason}
                          onChange={(event) =>
                            setCancelReason(
                              event.target.value
                            )
                          }
                          placeholder="Cancellation reason..."
                          maxLength={500}
                        />

                        <div className="flex flex-wrap justify-end gap-2">
                          <Button
                            variant="outline"
                            onClick={() => {
                              setShowCancelForm(
                                false
                              );
                              setCancelReason(
                                ""
                              );
                            }}
                            disabled={saving}
                          >
                            Keep session
                          </Button>

                          <Button
                            variant="destructive"
                            onClick={() =>
                              void handleCancel(
                                selectedSession
                              )
                            }
                            disabled={
                              saving ||
                              cancelReason.trim()
                                .length === 0
                            }
                          >
                            Confirm cancellation
                          </Button>
                        </div>
                      </div>
                    )}

                  {/* =================================================
                      CLOSE
                  ================================================== */}

                  <div className="flex justify-end border-t pt-4">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setShowDetails(false);
                        setSelectedSession(
                          null
                        );
                        setShowCancelForm(
                          false
                        );
                        setCancelReason("");
                      }}
                      disabled={saving}
                    >
                      Close
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
      </div>
    </DashboardShell>
  );
}
