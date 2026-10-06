"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Banknote,
  Clock,
  CreditCard,
  Monitor,
  Receipt,
  ShoppingCart,
  Users,
  Wifi,
  WifiOff,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import StatCard from "@/components/dashboard/StatCard";

import { useAuthStore } from "@/store/auth.store";

import { sessionService } from "@/services/session.service";
import { terminalService } from "@/services/terminal.service";


import type { Session } from "@/types/session";
import type { Terminal } from "@/types/terminal";
import type {
  AttendantReportResponse,
} from "@/types/attendant-report";
import { attendantReportService } from "@/services/attendant-report.service";

function getKenyaDateString(date = new Date()) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: "Africa/Nairobi",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(date);
}

function getKenyaDayRange() {
  const date = getKenyaDateString();

  return {
    start: `${date}T00:00:00+03:00`,
    end: `${date}T23:59:59.999+03:00`,
  };
}

function formatMoney(value: string | number) {
  const amount =
    typeof value === "number"
      ? value
      : Number.parseFloat(value || "0");

  if (!Number.isFinite(amount)) {
    return "KSh 0.00";
  }

  return new Intl.NumberFormat("en-KE", {
    style: "currency",
    currency: "KES",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

export default function AttendantDashboard() {
  const user = useAuthStore((state) => state.user);

  const isOfflineSession = useAuthStore(
    (state) => state.isOffline
  );

  const offlineBranchName = useAuthStore(
    (state) => state.offlineBranchName
  );

  const offlineBranchId = useAuthStore(
    (state) => state.offlineBranchId
  );

  const [isOnline, setIsOnline] = useState(true);

  const [terminals, setTerminals] = useState<Terminal[]>([]);
  const [sessions, setSessions] = useState<Session[]>([]);

  const [report, setReport] =
    useState<AttendantReportResponse | null>(null);

  const [loading, setLoading] = useState(true);
  const [reportLoading, setReportLoading] =
    useState(true);

  const [error, setError] = useState("");
  const [reportError, setReportError] =
    useState("");

  useEffect(() => {
    setIsOnline(navigator.onLine);

    function handleOnline() {
      setIsOnline(true);
    }

    function handleOffline() {
      setIsOnline(false);
    }

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  const loadReport = useCallback(async () => {
    if (!navigator.onLine || isOfflineSession) {
      setReportLoading(false);
      return;
    }

    setReportLoading(true);
    setReportError("");

    try {
      const { start, end } = getKenyaDayRange();

      const response =
        await attendantReportService.getSummary(
          start,
          end
        );

      setReport(response);
    } catch (err) {
      console.error(
        "Failed to load attendant report:",
        err
      );

      setReportError(
        err instanceof Error
          ? err.message
          : "Failed to load today's earnings."
      );
    } finally {
      setReportLoading(false);
    }
  }, [isOfflineSession]);

  const loadDashboard = useCallback(async () => {
    if (!navigator.onLine || isOfflineSession) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError("");

    try {
      const terminalResponse =
        await terminalService.listAttendantTerminals();

      const terminalItems =
        terminalResponse.terminals ?? [];

      setTerminals(terminalItems);

      const branchId =
        offlineBranchId ||
        terminalItems.find(
          (terminal) => terminal.branch_id
        )?.branch_id ||
        "";

      if (!branchId) {
        setSessions([]);
        setLoading(false);
        return;
      }

      const sessionResponse =
        await sessionService.list(branchId);

      setSessions(sessionResponse.sessions ?? []);
    } catch (err) {
      console.error(
        "Failed to load attendant dashboard:",
        err
      );

      setError(
        err instanceof Error
          ? err.message
          : "Failed to load dashboard data."
      );
    } finally {
      setLoading(false);
    }
  }, [isOfflineSession, offlineBranchId]);

  const loadAll = useCallback(async () => {
    await Promise.all([
      loadDashboard(),
      loadReport(),
    ]);
  }, [loadDashboard, loadReport]);

  useEffect(() => {
    void loadAll();
  }, [loadAll]);

  useEffect(() => {
    function handleOnline() {
      void loadAll();
    }

    window.addEventListener("online", handleOnline);

    return () => {
      window.removeEventListener(
        "online",
        handleOnline
      );
    };
  }, [loadAll]);

  useEffect(() => {
    if (!isOnline || isOfflineSession) {
      return;
    }

    const interval = window.setInterval(() => {
      void loadAll();
    }, 30_000);

    return () => {
      window.clearInterval(interval);
    };
  }, [isOnline, isOfflineSession, loadAll]);

  const today = getKenyaDateString();

  const todaySessions = sessions.filter((session) => {
    return getKenyaDateString(
      new Date(session.started_at)
    ) === today;
  });

  const activeSessions = sessions.filter((session) => {
    return (
      session.status === "active" ||
      session.status === "paused"
    );
  });

  const uniqueCustomerIds = new Set(
    todaySessions.map(
      (session) => session.customer_id
    )
  );

  const customerCount = uniqueCustomerIds.size;
  const activeSessionCount =
    activeSessions.length;
  const terminalCount = terminals.length;

  const branchName =
    offlineBranchName ||
    terminals.find(
      (terminal) => terminal.branch_id
    )?.branch_name ||
    "Assigned Branch";

  const connectionStatus =
    isOnline && !isOfflineSession
      ? "Online"
      : isOnline && isOfflineSession
        ? "Online • Local session"
        : "Offline";

  const summary = report?.summary;

  const paymentMethods =
    report?.payment_methods ?? [];

  const dailyTotals =
    report?.daily_totals ?? [];

  const salesRevenue =
    summary?.sales_revenue ?? "0";

  const sessionRevenue =
    summary?.session_revenue ?? "0";

  const totalRevenue =
    summary?.total_revenue ?? "0";

  const confirmedCollections =
    summary?.confirmed_collections ?? "0";

  const salesCount =
    summary?.sales_count ?? 0;

  const sessionCount =
    summary?.session_count ?? 0;

  const paymentsCount =
    summary?.payments_count ?? 0;

  const hasReportData =
    Boolean(report) &&
    !reportLoading;

  const topPaymentMethod = useMemo(() => {
    if (paymentMethods.length === 0) {
      return null;
    }

    return [...paymentMethods].sort(
      (a, b) =>
        Number.parseFloat(b.amount || "0") -
        Number.parseFloat(a.amount || "0")
    )[0];
  }, [paymentMethods]);

  const formatTime = (value: string) => {
    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
      return "—";
    }

    return date.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getSessionStatusLabel = (
    status: Session["status"]
  ) => {
    switch (status) {
      case "active":
        return "Active";
      case "paused":
        return "Paused";
      case "completed":
        return "Completed";
      case "cancelled":
        return "Cancelled";
      default:
        return status;
    }
  };

  return (
    <DashboardShell role="attendant">
      <PageHeader
        title="Dashboard"
        description="Manage your cyber café activities."
      />

      <div className="mb-6 flex flex-col gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-sm text-slate-500">
            Welcome back
          </p>

          <h2 className="mt-1 text-base font-semibold text-slate-950">
            {user?.full_name || "Attendant"}
          </h2>

          <p className="mt-1 text-sm text-slate-500">
            {branchName}
          </p>
        </div>

        <div
          className={`inline-flex w-fit items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold ${
            isOnline
              ? "bg-emerald-50 text-emerald-700"
              : "bg-amber-50 text-amber-700"
          }`}
        >
          {isOnline ? (
            <Wifi className="h-3.5 w-3.5" />
          ) : (
            <WifiOff className="h-3.5 w-3.5" />
          )}

          {connectionStatus}
        </div>
      </div>

      {error &&
        isOnline &&
        !isOfflineSession && (
          <div className="mb-6 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Customers"
          value={
            loading
              ? "…"
              : String(customerCount)
          }
          description="Customers today"
          icon={
            <Users className="h-5 w-5" />
          }
        />

        <StatCard
          title="Active Sessions"
          value={
            loading
              ? "…"
              : String(activeSessionCount)
          }
          description="Currently active"
          icon={
            <Clock className="h-5 w-5" />
          }
        />

        <StatCard
          title="Terminals"
          value={
            loading
              ? "…"
              : String(terminalCount)
          }
          description="Assigned computers"
          icon={
            <Monitor className="h-5 w-5" />
          }
        />

        <StatCard
          title="Today's Revenue"
          value={
            reportLoading
              ? "…"
              : formatMoney(totalRevenue)
          }
          description="Your completed revenue"
          icon={
            <Banknote className="h-5 w-5" />
          }
        />
      </div>

      <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Sales Revenue"
          value={
            reportLoading
              ? "…"
              : formatMoney(salesRevenue)
          }
          description={`${salesCount} completed sales`}
          icon={
            <ShoppingCart className="h-5 w-5" />
          }
        />

        <StatCard
          title="Session Revenue"
          value={
            reportLoading
              ? "…"
              : formatMoney(sessionRevenue)
          }
          description={`${sessionCount} completed sessions`}
          icon={
            <Clock className="h-5 w-5" />
          }
        />

        <StatCard
          title="Collections"
          value={
            reportLoading
              ? "…"
              : formatMoney(confirmedCollections)
          }
          description={`${paymentsCount} confirmed payments`}
          icon={
            <CreditCard className="h-5 w-5" />
          }
        />

        <StatCard
          title="Receipts"
          value={
            reportLoading
              ? "…"
              : String(paymentsCount)
          }
          description="Confirmed payments today"
          icon={
            <Receipt className="h-5 w-5" />
          }
        />
      </div>

      {reportError &&
        isOnline &&
        !isOfflineSession && (
          <div className="mt-6 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {reportError}
          </div>
        )}

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="text-base font-semibold text-slate-950">
                Today&apos;s Sales
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Sales completed by you today.
              </p>
            </div>

            <ShoppingCart className="h-5 w-5 text-slate-400" />
          </div>

          <div className="mt-6 rounded-lg bg-slate-50 p-5">
            <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
              Sales revenue
            </p>

            <p className="mt-2 text-3xl font-bold tracking-tight text-slate-950">
              {reportLoading
                ? "…"
                : formatMoney(salesRevenue)}
            </p>

            <p className="mt-2 text-sm text-slate-500">
              {salesCount} completed sale
              {salesCount === 1 ? "" : "s"}
            </p>
          </div>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="text-base font-semibold text-slate-950">
                Payment Methods
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Confirmed collections from your sales.
              </p>
            </div>

            <CreditCard className="h-5 w-5 text-slate-400" />
          </div>

          {reportLoading ? (
            <div className="mt-6 flex min-h-32 items-center justify-center rounded-lg border border-dashed border-slate-200">
              <p className="text-sm text-slate-400">
                Loading payment data...
              </p>
            </div>
          ) : paymentMethods.length === 0 ? (
            <div className="mt-6 flex min-h-32 items-center justify-center rounded-lg border border-dashed border-slate-200">
              <p className="text-sm text-slate-400">
                No confirmed payments today.
              </p>
            </div>
          ) : (
            <div className="mt-6 space-y-3">
              {paymentMethods.map((payment) => (
                <div
                  key={payment.method}
                  className="flex items-center justify-between rounded-lg border border-slate-100 px-4 py-3"
                >
                  <div>
                    <p className="font-medium capitalize text-slate-950">
                      {payment.method}
                    </p>

                    <p className="mt-0.5 text-xs text-slate-500">
                      {payment.count} payment
                      {payment.count === 1
                        ? ""
                        : "s"}
                    </p>
                  </div>

                  <p className="font-semibold text-slate-950">
                    {formatMoney(payment.amount)}
                  </p>
                </div>
              ))}

              {topPaymentMethod && (
                <p className="pt-1 text-xs text-slate-500">
                  Top payment method:{" "}
                  <span className="font-semibold text-slate-700">
                    {topPaymentMethod.method}
                  </span>
                </p>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="mt-6 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-slate-950">
              Today&apos;s Activity
            </h2>

            <p className="mt-1 text-sm text-slate-500">
              Recent sessions from your assigned branch.
            </p>
          </div>

          {!isOnline && (
            <div className="hidden items-center gap-2 text-xs font-medium text-amber-600 sm:flex">
              <WifiOff className="h-4 w-4" />
              Working offline
            </div>
          )}
        </div>

        {loading ? (
          <div className="mt-8 flex min-h-40 items-center justify-center rounded-lg border border-dashed border-slate-200">
            <p className="text-sm text-slate-400">
              Loading activity...
            </p>
          </div>
        ) : todaySessions.length === 0 ? (
          <div className="mt-8 flex min-h-40 items-center justify-center rounded-lg border border-dashed border-slate-200">
            <p className="text-sm text-slate-400">
              No activity yet today.
            </p>
          </div>
        ) : (
          <div className="mt-6 overflow-x-auto">
            <div className="min-w-[650px]">
              <div className="grid grid-cols-[1.4fr_1fr_1fr_1fr] gap-4 border-b border-slate-200 px-4 pb-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                <span>Customer</span>
                <span>Terminal</span>
                <span>Started</span>
                <span>Status</span>
              </div>

              <div className="divide-y divide-slate-100">
                {todaySessions
                  .slice()
                  .sort(
                    (a, b) =>
                      new Date(
                        b.started_at
                      ).getTime() -
                      new Date(
                        a.started_at
                      ).getTime()
                  )
                  .slice(0, 10)
                  .map((session) => (
                    <div
                      key={session.id}
                      className="grid grid-cols-[1.4fr_1fr_1fr_1fr] gap-4 px-4 py-4 text-sm"
                    >
                      <div>
                        <p className="font-medium text-slate-950">
                          {session.customer_name ||
                            "Customer"}
                        </p>

                        <p className="mt-0.5 text-xs text-slate-500">
                          {session.customer_type}
                        </p>
                      </div>

                      <div className="text-slate-600">
                        {session.terminal_name ||
                          "Terminal"}
                      </div>

                      <div className="text-slate-600">
                        {formatTime(
                          session.started_at
                        )}
                      </div>

                      <div>
                        <span
                          className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${
                            session.status ===
                            "active"
                              ? "bg-emerald-50 text-emerald-700"
                              : session.status ===
                                  "paused"
                                ? "bg-amber-50 text-amber-700"
                                : session.status ===
                                    "completed"
                                  ? "bg-slate-100 text-slate-700"
                                  : "bg-red-50 text-red-700"
                          }`}
                        >
                          {getSessionStatusLabel(
                            session.status
                          )}
                        </span>
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          </div>
        )}

        {!isOnline && (
          <div className="mt-4 flex items-center gap-2 rounded-lg bg-amber-50 px-4 py-3 text-xs text-amber-700">
            <WifiOff className="h-4 w-4 shrink-0" />

            <span>
              You are offline. The dashboard is
              showing the last available server data.
              Terminal operations continue through
              the offline terminal system.
            </span>
          </div>
        )}
      </div>

      {hasReportData &&
        dailyTotals.length > 0 && (
          <div className="mt-6 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <div>
              <h2 className="text-base font-semibold text-slate-950">
                Daily Revenue
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Revenue generated by you for the selected
                report period.
              </p>
            </div>

            <div className="mt-6 overflow-x-auto">
              <table className="w-full min-w-[650px] text-sm">
                <thead>
                  <tr className="border-b border-slate-200 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                    <th className="px-3 py-3">
                      Date
                    </th>
                    <th className="px-3 py-3">
                      Sessions
                    </th>
                    <th className="px-3 py-3">
                      Sales
                    </th>
                    <th className="px-3 py-3">
                      Session Revenue
                    </th>
                    <th className="px-3 py-3">
                      Sales Revenue
                    </th>
                    <th className="px-3 py-3">
                      Total
                    </th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-slate-100">
                  {dailyTotals.map((day) => (
                    <tr key={day.date}>
                      <td className="px-3 py-3 font-medium text-slate-950">
                        {day.date}
                      </td>

                      <td className="px-3 py-3 text-slate-600">
                        {day.session_count}
                      </td>

                      <td className="px-3 py-3 text-slate-600">
                        {day.sales_count}
                      </td>

                      <td className="px-3 py-3 text-slate-600">
                        {formatMoney(
                          day.session_revenue
                        )}
                      </td>

                      <td className="px-3 py-3 text-slate-600">
                        {formatMoney(
                          day.sales_revenue
                        )}
                      </td>

                      <td className="px-3 py-3 font-semibold text-slate-950">
                        {formatMoney(
                          day.total_revenue
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
    </DashboardShell>
  );
}