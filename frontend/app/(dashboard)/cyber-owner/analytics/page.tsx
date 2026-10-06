"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Activity,
  ArrowDownRight,
  ArrowUpRight,
  BarChart3,
  Building2,
  CalendarDays,
  DollarSign,
  Loader2,
  Monitor,
  RefreshCw,
  ShoppingCart,
  TrendingUp,
  Wallet,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";

import { useReportStore } from "@/store/report.store";
import { useAuthStore } from "@/store/auth.store";
import { branchService } from "@/services/branch.service";
import type { Branch } from "@/types/branch";

function formatCurrency(
  value: string | number | undefined | null,
) {
  const num =
    typeof value === "string"
      ? Number.parseFloat(value)
      : Number(value ?? 0);

  if (!Number.isFinite(num)) {
    return "KES 0.00";
  }

  return new Intl.NumberFormat("en-KE", {
    style: "currency",
    currency: "KES",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(num);
}

function getLocalDateString(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

function getTodayPeriod() {
  const now = new Date();

  const start = new Date(now);
  start.setHours(0, 0, 0, 0);

  const end = new Date(start);
  end.setDate(end.getDate() + 1);

  return {
    start: start.toISOString(),
    end: end.toISOString(),
  };
}

function getSelectedEndDate(value?: string) {
  if (!value) {
    return "";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value.split("T")[0];
  }

  date.setDate(date.getDate() - 1);

  return getLocalDateString(date);
}

function formatTrendDate(value: string) {
  const date = new Date(`${value}T00:00:00`);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleDateString("en-KE", {
    month: "short",
    day: "numeric",
  });
}

function formatMonth(value: string) {
  const date = new Date(`${value}-01T00:00:00`);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleDateString("en-KE", {
    month: "short",
    year: "numeric",
  });
}

function TrendBar({
  value,
  maximum,
}: {
  value: number;
  maximum: number;
}) {
  const percentage =
    maximum > 0
      ? Math.max(4, (value / maximum) * 100)
      : 4;

  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100">
      <div
        className="h-full rounded-full bg-[#0757B8] transition-all"
        style={{
          width: `${Math.min(100, percentage)}%`,
        }}
      />
    </div>
  );
}

export default function OwnerAnalyticsPage() {
  const {
    data,
    loading,
    error,
    filters,
    setFilters,
    fetchReport,
  } = useReportStore();

  const token = useAuthStore((state) => state.token);

  const [branches, setBranches] =
    useState<Branch[]>([]);

  const [branchesLoading, setBranchesLoading] =
    useState(false);

  useEffect(() => {
    if (!token) {
      return;
    }

    const authToken = token;

    async function loadBranches() {
      setBranchesLoading(true);

      try {
        const response =
          await branchService.list(authToken);

        setBranches(response.branches);

        const currentFilters =
          useReportStore.getState().filters;

        if (
          !currentFilters.period_start ||
          !currentFilters.period_end
        ) {
          const period = getTodayPeriod();

          setFilters({
            period_start: period.start,
            period_end: period.end,
          });
        }
      } catch (err) {
        console.error(
          "Failed to load branches for analytics:",
          err,
        );
      } finally {
        setBranchesLoading(false);
      }
    }

    loadBranches();
  }, [token, setFilters]);

  useEffect(() => {
    if (
      !token ||
      !filters.period_start ||
      !filters.period_end
    ) {
      return;
    }

    fetchReport();
  }, [
    token,
    filters.period_start,
    filters.period_end,
    filters.branch_id,
    fetchReport,
  ]);

  const summary = data?.summary;

  const branchesData = data?.branches ?? [];

  const dailyTotals = data?.daily_totals ?? [];

  const monthlyTotals = data?.monthly_totals ?? [];

  const bestBranch = useMemo(() => {
    if (branchesData.length === 0) {
      return null;
    }

    return [...branchesData].sort(
      (a, b) =>
        Number.parseFloat(b.net_amount || "0") -
        Number.parseFloat(a.net_amount || "0"),
    )[0];
  }, [branchesData]);

  const highestRevenueBranch = useMemo(() => {
    if (branchesData.length === 0) {
      return null;
    }

    return [...branchesData].sort(
      (a, b) =>
        Number.parseFloat(b.total_revenue || "0") -
        Number.parseFloat(a.total_revenue || "0"),
    )[0];
  }, [branchesData]);

  const highestServiceLikeActivity = useMemo(() => {
    if (branchesData.length === 0) {
      return null;
    }

    return [...branchesData].sort(
      (a, b) =>
        b.sales_count +
        b.session_count -
        (a.sales_count + a.session_count),
    )[0];
  }, [branchesData]);

  const dailyMaximum = useMemo(() => {
    return Math.max(
      0,
      ...dailyTotals.map((item) =>
        Number.parseFloat(item.total_revenue || "0"),
      ),
    );
  }, [dailyTotals]);

  const monthlyMaximum = useMemo(() => {
    return Math.max(
      0,
      ...monthlyTotals.map((item) =>
        Number.parseFloat(item.total_revenue || "0"),
      ),
    );
  }, [monthlyTotals]);

  const selectedStartDate =
    filters.period_start
      ? filters.period_start.split("T")[0]
      : "";

  const selectedEndDate =
    getSelectedEndDate(filters.period_end);

  const selectedBranchName = useMemo(() => {
    if (!filters.branch_id) {
      return "All Branches";
    }

    const branch = branches.find(
      (item) => item.id === filters.branch_id,
    );

    return branch?.name || "Selected Branch";
  }, [branches, filters.branch_id]);

  const handleApplyFilters = (
    event: React.FormEvent<HTMLFormElement>,
  ) => {
    event.preventDefault();

    if (
      !filters.period_start ||
      !filters.period_end
    ) {
      return;
    }

    fetchReport();
  };

  const handleStartDateChange = (
    value: string,
  ) => {
    if (!value) {
      setFilters({
        period_start: undefined,
      });

      return;
    }

    const date = new Date(
      `${value}T00:00:00`,
    );

    setFilters({
      period_start: date.toISOString(),
    });
  };

  const handleEndDateChange = (
    value: string,
  ) => {
    if (!value) {
      setFilters({
        period_end: undefined,
      });

      return;
    }

    const date = new Date(
      `${value}T00:00:00`,
    );

    date.setDate(date.getDate() + 1);

    setFilters({
      period_end: date.toISOString(),
    });
  };

  return (
    <DashboardShell role="owner">
      <div className="space-y-6">
        {/* HEADER */}
        <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <div className="flex items-center gap-2">
              <BarChart3 className="h-6 w-6 text-[#0757B8]" />

              <h1 className="text-2xl font-bold tracking-tight text-slate-950">
                Analytics
              </h1>
            </div>

            <p className="mt-1 text-sm text-slate-500">
              Understand revenue, profitability,
              branch performance, and business
              activity.
            </p>
          </div>

          <Button
            variant="outline"
            onClick={() => fetchReport()}
            disabled={
              loading ||
              !filters.period_start ||
              !filters.period_end
            }
            className="border-slate-200 bg-white text-slate-700 shadow-sm hover:bg-slate-50"
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

        {/* ERROR */}
        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            <p className="font-semibold">
              Unable to load analytics
            </p>

            <p className="mt-1">
              {error}
            </p>
          </div>
        )}

        {/* FILTERS */}
        <Card className="shadow-sm">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base font-medium">
              <CalendarDays className="h-4 w-4 text-[#0757B8]" />
              Analytics Period
            </CardTitle>
          </CardHeader>

          <CardContent>
            <form
              onSubmit={handleApplyFilters}
              className="grid items-end gap-4 sm:grid-cols-2 lg:grid-cols-4"
            >
              {/* BRANCH */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="analytics-branch"
                  className="text-xs text-slate-600"
                >
                  Branch
                </Label>

                <select
                  id="analytics-branch"
                  value={
                    filters.branch_id || ""
                  }
                  onChange={(event) =>
                    setFilters({
                      branch_id:
                        event.target.value ||
                        undefined,
                    })
                  }
                  disabled={
                    branchesLoading
                  }
                  className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none focus:border-[#0757B8] focus:ring-2 focus:ring-[#0757B8]/20 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="">
                    All Branches
                  </option>

                  {branches.map(
                    (branch) => (
                      <option
                        key={branch.id}
                        value={branch.id}
                      >
                        {branch.name}
                      </option>
                    ),
                  )}
                </select>
              </div>

              {/* START */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="analytics-start"
                  className="text-xs text-slate-600"
                >
                  Start Date
                </Label>

                <Input
                  id="analytics-start"
                  type="date"
                  value={selectedStartDate}
                  onChange={(event) =>
                    handleStartDateChange(
                      event.target.value,
                    )
                  }
                  className="h-10 border-slate-200 shadow-sm"
                />
              </div>

              {/* END */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="analytics-end"
                  className="text-xs text-slate-600"
                >
                  End Date
                </Label>

                <Input
                  id="analytics-end"
                  type="date"
                  value={selectedEndDate}
                  onChange={(event) =>
                    handleEndDateChange(
                      event.target.value,
                    )
                  }
                  className="h-10 border-slate-200 shadow-sm"
                />
              </div>

              <Button
                type="submit"
                disabled={
                  loading ||
                  !filters.period_start ||
                  !filters.period_end
                }
                className="h-10 bg-[#0757B8] shadow-sm hover:bg-[#064A9D]"
              >
                {loading && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}

                Apply Filters
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* REPORT CONTEXT */}
        {summary && (
          <div className="rounded-xl border border-blue-100 bg-blue-50/50 px-4 py-3 text-sm text-slate-600">
            Showing analytics for{" "}
            <span className="font-semibold text-slate-900">
              {selectedBranchName}
            </span>{" "}
            from{" "}
            <span className="font-semibold text-slate-900">
              {selectedStartDate}
            </span>{" "}
            to{" "}
            <span className="font-semibold text-slate-900">
              {selectedEndDate}
            </span>
          </div>
        )}

        {/* OVERVIEW CARDS */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-green-50 text-green-600">
                <DollarSign className="h-5 w-5" />
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Total Revenue
                </p>

                <p className="text-2xl font-bold text-green-600">
                  {formatCurrency(
                    summary?.total_revenue,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  Sessions + sales
                </p>
              </div>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
                <Wallet className="h-5 w-5" />
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Net Revenue
                </p>

                <p className="text-2xl font-bold text-slate-950">
                  {formatCurrency(
                    summary?.net_amount,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  Revenue − expenses
                </p>
              </div>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600">
                <ArrowUpRight className="h-5 w-5" />
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Collections
                </p>

                <p className="text-2xl font-bold text-emerald-600">
                  {formatCurrency(
                    summary?.total_collections,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  {summary?.payments_count || 0} payments
                </p>
              </div>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-red-50 text-red-600">
                <ArrowDownRight className="h-5 w-5" />
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Expenses
                </p>

                <p className="text-2xl font-bold text-red-600">
                  {formatCurrency(
                    summary?.total_expenses,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  {summary?.expenses_count || 0} records
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* BUSINESS ACTIVITY */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card className="shadow-sm">
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <p className="text-sm text-slate-500">
                  Session Revenue
                </p>

                <Monitor className="h-5 w-5 text-[#0757B8]" />
              </div>

              <p className="mt-2 text-xl font-bold text-slate-950">
                {formatCurrency(
                  summary?.session_revenue,
                )}
              </p>

              <p className="mt-1 text-xs text-slate-400">
                {summary?.session_count || 0} completed sessions
              </p>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <p className="text-sm text-slate-500">
                  Sales Revenue
                </p>

                <ShoppingCart className="h-5 w-5 text-indigo-600" />
              </div>

              <p className="mt-2 text-xl font-bold text-slate-950">
                {formatCurrency(
                  summary?.sales_revenue,
                )}
              </p>

              <p className="mt-1 text-xs text-slate-400">
                {summary?.sales_count || 0} completed sales
              </p>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <p className="text-sm text-slate-500">
                  Total Activity
                </p>

                <Activity className="h-5 w-5 text-violet-600" />
              </div>

              <p className="mt-2 text-xl font-bold text-slate-950">
                {(summary?.session_count || 0) +
                  (summary?.sales_count || 0)}
              </p>

              <p className="mt-1 text-xs text-slate-400">
                Sessions + sales
              </p>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <p className="text-sm text-slate-500">
                  Revenue / Activity
                </p>

                <TrendingUp className="h-5 w-5 text-green-600" />
              </div>

              <p className="mt-2 text-xl font-bold text-slate-950">
                {formatCurrency(
                  summary &&
                    summary.session_count +
                      summary.sales_count >
                      0
                    ? Number.parseFloat(
                        summary.total_revenue || "0",
                      ) /
                        (summary.session_count +
                          summary.sales_count)
                    : 0,
                )}
              </p>

              <p className="mt-1 text-xs text-slate-400">
                Average revenue per activity
              </p>
            </CardContent>
          </Card>
        </div>

        {/* BEST PERFORMERS */}
        <div className="grid gap-4 lg:grid-cols-3">
          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <TrendingUp className="h-4 w-4 text-green-600" />
                Best Branch
              </CardTitle>
            </CardHeader>

            <CardContent>
              {bestBranch ? (
                <>
                  <p className="text-xl font-bold text-slate-950">
                    {bestBranch.branch_name}
                  </p>

                  <p className="mt-1 text-sm text-slate-500">
                    Highest net revenue
                  </p>

                  <p className="mt-4 text-2xl font-bold text-green-600">
                    {formatCurrency(
                      bestBranch.net_amount,
                    )}
                  </p>

                  <Badge className="mt-3 border-0 bg-green-100 text-green-700 hover:bg-green-100">
                    Top performer
                  </Badge>
                </>
              ) : (
                <p className="text-sm text-slate-500">
                  No branch data for this period.
                </p>
              )}
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <DollarSign className="h-4 w-4 text-[#0757B8]" />
                Highest Revenue
              </CardTitle>
            </CardHeader>

            <CardContent>
              {highestRevenueBranch ? (
                <>
                  <p className="text-xl font-bold text-slate-950">
                    {highestRevenueBranch.branch_name}
                  </p>

                  <p className="mt-1 text-sm text-slate-500">
                    Highest gross revenue
                  </p>

                  <p className="mt-4 text-2xl font-bold text-[#0757B8]">
                    {formatCurrency(
                      highestRevenueBranch.total_revenue,
                    )}
                  </p>
                </>
              ) : (
                <p className="text-sm text-slate-500">
                  No branch data for this period.
                </p>
              )}
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Activity className="h-4 w-4 text-violet-600" />
                Highest Activity
              </CardTitle>
            </CardHeader>

            <CardContent>
              {highestServiceLikeActivity ? (
                <>
                  <p className="text-xl font-bold text-slate-950">
                    {
                      highestServiceLikeActivity.branch_name
                    }
                  </p>

                  <p className="mt-1 text-sm text-slate-500">
                    Most sessions and sales
                  </p>

                  <p className="mt-4 text-2xl font-bold text-violet-600">
                    {highestServiceLikeActivity.session_count +
                      highestServiceLikeActivity.sales_count}
                  </p>

                  <p className="mt-1 text-xs text-slate-400">
                    Total recorded activities
                  </p>
                </>
              ) : (
                <p className="text-sm text-slate-500">
                  No activity for this period.
                </p>
              )}
            </CardContent>
          </Card>
        </div>

        {/* BRANCH PERFORMANCE */}
        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg font-bold">
              <Building2 className="h-5 w-5 text-[#0757B8]" />
              Branch Performance
            </CardTitle>
          </CardHeader>

          <CardContent>
            {loading ? (
              <div className="flex items-center justify-center py-12 text-sm text-slate-500">
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Loading analytics...
              </div>
            ) : branchesData.length === 0 ? (
              <div className="py-12 text-center">
                <Building2 className="mx-auto h-8 w-8 text-slate-300" />

                <p className="mt-3 font-semibold text-slate-900">
                  No branch activity
                </p>

                <p className="mt-1 text-sm text-slate-500">
                  No branch activity was recorded
                  for this period.
                </p>
              </div>
            ) : (
              <div className="space-y-5">
                {branchesData.map((branch) => {
                  const net =
                    Number.parseFloat(
                      branch.net_amount || "0",
                    );

                  const revenue =
                    Number.parseFloat(
                      branch.total_revenue || "0",
                    );

                  const collections =
                    Number.parseFloat(
                      branch.total_collections || "0",
                    );

                  const expenses =
                    Number.parseFloat(
                      branch.total_expenses || "0",
                    );

                  const isBest =
                    bestBranch?.branch_id ===
                    branch.branch_id;

                  return (
                    <div
                      key={branch.branch_id}
                      className="rounded-xl border border-slate-100 bg-slate-50/50 p-4"
                    >
                      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-2">
                            <h3 className="font-semibold text-slate-950">
                              {branch.branch_name}
                            </h3>

                            {isBest && (
                              <Badge className="border-0 bg-green-100 text-green-700 hover:bg-green-100">
                                Best net revenue
                              </Badge>
                            )}
                          </div>

                          <p className="mt-1 text-xs text-slate-500">
                            {branch.session_count} sessions •{" "}
                            {branch.sales_count} sales
                          </p>
                        </div>

                        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                          <div>
                            <p className="text-xs text-slate-500">
                              Revenue
                            </p>

                            <p className="font-semibold text-slate-900">
                              {formatCurrency(revenue)}
                            </p>
                          </div>

                          <div>
                            <p className="text-xs text-slate-500">
                              Collections
                            </p>

                            <p className="font-semibold text-green-600">
                              {formatCurrency(collections)}
                            </p>
                          </div>

                          <div>
                            <p className="text-xs text-slate-500">
                              Expenses
                            </p>

                            <p className="font-semibold text-red-600">
                              {formatCurrency(expenses)}
                            </p>
                          </div>

                          <div>
                            <p className="text-xs text-slate-500">
                              Net
                            </p>

                            <p className="font-bold text-slate-950">
                              {formatCurrency(net)}
                            </p>
                          </div>
                        </div>
                      </div>

                      <div className="mt-4">
                        <TrendBar
                          value={Math.max(net, 0)}
                          maximum={Math.max(
                            0,
                            ...branchesData.map(
                              (item) =>
                                Number.parseFloat(
                                  item.net_amount || "0",
                                ),
                            ),
                          )}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* DAILY TREND */}
        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg font-bold">
              <TrendingUp className="h-5 w-5 text-[#0757B8]" />
              Daily Revenue Trend
            </CardTitle>
          </CardHeader>

          <CardContent>
            {dailyTotals.length === 0 ? (
              <div className="py-10 text-center text-sm text-slate-500">
                No daily trend data is available
                for the selected period.
              </div>
            ) : (
              <div className="space-y-4">
                {dailyTotals.map((item) => {
                  const revenue =
                    Number.parseFloat(
                      item.total_revenue || "0",
                    );

                  return (
                    <div
                      key={item.date}
                      className="grid gap-3 sm:grid-cols-[90px_1fr_120px] sm:items-center"
                    >
                      <p className="text-sm font-medium text-slate-600">
                        {formatTrendDate(
                          item.date,
                        )}
                      </p>

                      <TrendBar
                        value={revenue}
                        maximum={dailyMaximum}
                      />

                      <div className="text-left sm:text-right">
                        <p className="font-semibold text-slate-900">
                          {formatCurrency(
                            revenue,
                          )}
                        </p>

                        <p className="text-xs text-slate-400">
                          {item.session_count} sessions •{" "}
                          {item.sales_count} sales
                        </p>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* MONTHLY TREND */}
        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg font-bold">
              <BarChart3 className="h-5 w-5 text-[#0757B8]" />
              Monthly Revenue Trend
            </CardTitle>
          </CardHeader>

          <CardContent>
            {monthlyTotals.length === 0 ? (
              <div className="py-10 text-center text-sm text-slate-500">
                No monthly trend data is available
                for the selected period.
              </div>
            ) : (
              <div className="space-y-4">
                {monthlyTotals.map((item) => {
                  const revenue =
                    Number.parseFloat(
                      item.total_revenue || "0",
                    );

                  return (
                    <div
                      key={item.month}
                      className="grid gap-3 sm:grid-cols-[120px_1fr_140px] sm:items-center"
                    >
                      <p className="text-sm font-medium text-slate-600">
                        {formatMonth(
                          item.month,
                        )}
                      </p>

                      <TrendBar
                        value={revenue}
                        maximum={monthlyMaximum}
                      />

                      <div className="text-left sm:text-right">
                        <p className="font-semibold text-slate-900">
                          {formatCurrency(
                            revenue,
                          )}
                        </p>

                        <p className="text-xs text-slate-400">
                          {item.session_count} sessions •{" "}
                          {item.sales_count} sales
                        </p>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* NOTE ABOUT SERVICE ANALYTICS */}
        <Card className="border-blue-100 bg-blue-50/40 shadow-sm">
          <CardContent className="flex gap-4 p-5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-100 text-[#0757B8]">
              <ShoppingCart className="h-5 w-5" />
            </div>

            <div>
              <p className="font-semibold text-slate-900">
                Service Performance
              </p>

              <p className="mt-1 text-sm leading-6 text-slate-600">
                Detailed service-level analytics
                such as quantity sold, revenue per
                service, and top services should come
                from the sales-item data. We will
                connect those figures to this dashboard
                once the backend exposes the service
                aggregation, rather than estimating
                them from overall sales totals.
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}