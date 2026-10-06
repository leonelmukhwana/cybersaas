
"use client";

import { useEffect, useState } from "react";
import {
  ArrowDownRight,
  ArrowUpRight,
  Building2,
  CreditCard,
  DollarSign,
  FileSpreadsheet,
  FileText,
  Filter,
  Loader2,
  RefreshCw,
  Wallet,
  Monitor,
  ShoppingCart,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

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

function formatDate(value?: string) {
  if (!value) {
    return "";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleDateString("en-KE", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
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

function getFileDate(value?: string) {
  if (!value) {
    return getLocalDateString(new Date());
  }

  return value.split("T")[0];
}

function getSelectedDate(value?: string) {
  if (!value) {
    return "";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value.split("T")[0];
  }

  return getLocalDateString(date);
}

export default function CyberOwnerReportPage() {
  const {
    data,
    loading,
    error,
    filters,
    setFilters,
    fetchReport,
  } = useReportStore();

  const token = useAuthStore((state) => state.token);

  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchesLoading, setBranchesLoading] = useState(false);

  const [downloadLoading, setDownloadLoading] = useState<
    "pdf" | "excel" | null
  >(null);

  /*
   * Load the owner's branches and establish today's
   * reporting period.
   *
   * The branch selector remains available so the owner
   * can select one specific branch or All Branches.
   */
  useEffect(() => {
    if (!token) {
      return;
    }

    const authToken = token;

    async function initializeReport() {
      setBranchesLoading(true);

      try {
        const branchResponse =
          await branchService.list(authToken);

        setBranches(branchResponse.branches);

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
          "Failed to initialize owner report:",
          err,
        );
      } finally {
        setBranchesLoading(false);
      }
    }

    initializeReport();
  }, [token, setFilters]);

  /*
   * Fetch the report whenever the selected branch
   * or reporting period changes.
   */
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

  const handleFilterSubmit = (
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

  const summary = data?.summary;

  const paymentMethods =
    data?.payment_methods ?? [];

  const branchBreakdown =
    data?.branches ?? [];

  const getSelectedBranchName = () => {
    if (!filters.branch_id) {
      return "All Branches";
    }

    const branch = branches.find(
      (item) => item.id === filters.branch_id,
    );

    return branch?.name || "Selected Branch";
  };

  /*
   * Get the date selected by the owner for display.
   *
   * The API uses an exclusive end boundary internally,
   * so period_end is one day after the selected end date.
   */
  const selectedStartDate =
    getSelectedDate(filters.period_start);

  const selectedEndDate = (() => {
    if (!filters.period_end) {
      return "";
    }

    const date = new Date(filters.period_end);

    if (Number.isNaN(date.getTime())) {
      return filters.period_end.split("T")[0];
    }

    date.setDate(date.getDate() - 1);

    return getLocalDateString(date);
  })();

  const handleDownloadExcel = async () => {
    if (!data) {
      return;
    }

    setDownloadLoading("excel");

    try {
      const XLSX = await import("xlsx");

      const branchName =
        getSelectedBranchName();

      const summaryRows = [
        {
          Report:
            "Cyber Owner Financial Report",
          "Period Start":
            selectedStartDate,
          "Period End":
            selectedEndDate,
          Branch: branchName,
        },
        {},
        {
          Metric: "Session Revenue",
          Amount: Number(
            summary?.session_revenue || 0,
          ),
          Count:
            summary?.session_count || 0,
        },
        {
          Metric: "Sales Revenue",
          Amount: Number(
            summary?.sales_revenue || 0,
          ),
          Count:
            summary?.sales_count || 0,
        },
        {
          Metric: "Total Revenue",
          Amount: Number(
            summary?.total_revenue || 0,
          ),
          Count: "",
        },
        {
          Metric: "Total Collections",
          Amount: Number(
            summary?.total_collections || 0,
          ),
          Count:
            summary?.payments_count || 0,
        },
        {
          Metric: "Total Expenses",
          Amount: Number(
            summary?.total_expenses || 0,
          ),
          Count:
            summary?.expenses_count || 0,
        },
        {
          Metric: "Net Revenue",
          Amount: Number(
            summary?.net_amount || 0,
          ),
          Count: "",
        },
      ];

      const paymentRows =
        paymentMethods.map((payment) => ({
          Method: payment.method,
          Amount: Number(
            payment.amount || 0,
          ),
          Transactions: payment.count,
        }));

      const branchRows =
        branchBreakdown.map((branch) => ({
          "Branch Name":
            branch.branch_name,
          "Session Count":
            branch.session_count,
          "Session Revenue": Number(
            branch.session_revenue || 0,
          ),
          "Sales Count":
            branch.sales_count,
          "Sales Revenue": Number(
            branch.sales_revenue || 0,
          ),
          "Total Revenue": Number(
            branch.total_revenue || 0,
          ),
          Collections: Number(
            branch.total_collections || 0,
          ),
          Expenses: Number(
            branch.total_expenses || 0,
          ),
          "Net Revenue": Number(
            branch.net_amount || 0,
          ),
        }));

      const workbook =
        XLSX.utils.book_new();

      const summarySheet =
        XLSX.utils.json_to_sheet(
          summaryRows,
        );

      const paymentSheet =
        XLSX.utils.json_to_sheet(
          paymentRows,
        );

      const branchSheet =
        XLSX.utils.json_to_sheet(
          branchRows,
        );

      XLSX.utils.book_append_sheet(
        workbook,
        summarySheet,
        "Summary",
      );

      XLSX.utils.book_append_sheet(
        workbook,
        paymentSheet,
        "Payment Methods",
      );

      XLSX.utils.book_append_sheet(
        workbook,
        branchSheet,
        "Branches",
      );

      const filename =
        `cybersaas-owner-report-${getFileDate(
          filters.period_start,
        )}.xlsx`;

      XLSX.writeFile(
        workbook,
        filename,
      );
    } catch (err) {
      console.error(
        "Failed to generate Excel report:",
        err,
      );
    } finally {
      setDownloadLoading(null);
    }
  };

  const handleDownloadPDF = async () => {
    if (!data) {
      return;
    }

    setDownloadLoading("pdf");

    try {
      const { jsPDF } =
        await import("jspdf");

      const pdf = new jsPDF();

      const branchName =
        getSelectedBranchName();

      let y = 20;

      pdf.setFontSize(18);

      pdf.text(
        "CyberSaaS - Cyber Owner Financial Report",
        14,
        y,
      );

      y += 10;

      pdf.setFontSize(10);

      pdf.text(
        `Branch: ${branchName}`,
        14,
        y,
      );

      y += 6;

      pdf.text(
        `Period: ${formatDate(
          filters.period_start,
        )} - ${formatDate(
          filters.period_end
            ? new Date(
                new Date(
                  filters.period_end,
                ).getTime() -
                  24 * 60 * 60 * 1000,
              ).toISOString()
            : undefined,
        )}`,
        14,
        y,
      );

      y += 12;

      pdf.setFontSize(13);

      pdf.text(
        "Financial Summary",
        14,
        y,
      );

      y += 8;

      pdf.setFontSize(10);

      const summaryLines = [
        `Session Revenue: ${formatCurrency(
          summary?.session_revenue,
        )}`,
        `Sales Revenue: ${formatCurrency(
          summary?.sales_revenue,
        )}`,
        `Total Revenue: ${formatCurrency(
          summary?.total_revenue,
        )}`,
        `Total Collections: ${formatCurrency(
          summary?.total_collections,
        )}`,
        `Total Expenses: ${formatCurrency(
          summary?.total_expenses,
        )}`,
        `Net Revenue: ${formatCurrency(
          summary?.net_amount,
        )}`,
        `Session Count: ${
          summary?.session_count || 0
        }`,
        `Sales Count: ${
          summary?.sales_count || 0
        }`,
        `Payments Count: ${
          summary?.payments_count || 0
        }`,
        `Expenses Count: ${
          summary?.expenses_count || 0
        }`,
      ];

      for (const line of summaryLines) {
        pdf.text(line, 14, y);
        y += 6;
      }

      y += 6;

      pdf.setFontSize(13);

      pdf.text(
        "Payment Methods",
        14,
        y,
      );

      y += 8;

      pdf.setFontSize(10);

      pdf.text("Method", 14, y);
      pdf.text("Amount", 80, y);
      pdf.text(
        "Transactions",
        140,
        y,
      );

      y += 6;

      for (const payment of paymentMethods) {
        pdf.text(
          String(payment.method),
          14,
          y,
        );

        pdf.text(
          formatCurrency(payment.amount),
          80,
          y,
        );

        pdf.text(
          String(payment.count),
          140,
          y,
        );

        y += 6;

        if (y > 275) {
          pdf.addPage();
          y = 20;
        }
      }

      y += 8;

      pdf.setFontSize(13);

      pdf.text(
        "Branch Performance",
        14,
        y,
      );

      y += 8;

      pdf.setFontSize(8);

      pdf.text("Branch", 14, y);
      pdf.text("Sessions", 60, y);
      pdf.text("Sales", 90, y);
      pdf.text("Revenue", 120, y);
      pdf.text(
        "Collections",
        155,
        y,
      );

      y += 6;

      for (const branch of branchBreakdown) {
        pdf.text(
          String(
            branch.branch_name,
          ).substring(0, 22),
          14,
          y,
        );

        pdf.text(
          String(
            branch.session_count,
          ),
          60,
          y,
        );

        pdf.text(
          String(
            branch.sales_count,
          ),
          90,
          y,
        );

        pdf.text(
          formatCurrency(
            branch.total_revenue,
          ),
          120,
          y,
        );

        pdf.text(
          formatCurrency(
            branch.total_collections,
          ),
          165,
          y,
        );

        y += 6;

        if (y > 275) {
          pdf.addPage();
          y = 20;
        }
      }

      const filename =
        `cybersaas-owner-report-${getFileDate(
          filters.period_start,
        )}.pdf`;

      pdf.save(filename);
    } catch (err) {
      console.error(
        "Failed to generate PDF report:",
        err,
      );
    } finally {
      setDownloadLoading(null);
    }
  };

  return (
    <DashboardShell role="owner">
      <div className="space-y-6">
        {/* HEADER */}
        <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950">
              Financial Reports
            </h1>

            <p className="mt-1 text-sm text-slate-500">
              Select a branch and reporting date
              to view its financial and
              operational performance.
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
              Unable to load report
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
              <Filter className="h-4 w-4 text-[#0757B8]" />
              Report Filters
            </CardTitle>
          </CardHeader>

          <CardContent>
            <form
              onSubmit={handleFilterSubmit}
              className="grid items-end gap-4 sm:grid-cols-2 lg:grid-cols-5"
            >
              {/* BRANCH */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="branch-select"
                  className="text-xs text-slate-600"
                >
                  Branch
                </Label>

                <select
                  id="branch-select"
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

              {/* START DATE */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="period-start"
                  className="text-xs text-slate-600"
                >
                  Start Date
                </Label>

                <Input
                  id="period-start"
                  type="date"
                  value={
                    filters.period_start
                      ? filters.period_start.split(
                          "T",
                        )[0]
                      : ""
                  }
                  onChange={(event) => {
                    if (
                      !event.target.value
                    ) {
                      setFilters({
                        period_start:
                          undefined,
                      });

                      return;
                    }

                    const date =
                      new Date(
                        `${event.target.value}T00:00:00`,
                      );

                    setFilters({
                      period_start:
                        date.toISOString(),
                    });
                  }}
                  className="h-10 border-slate-200 shadow-sm"
                />
              </div>

              {/* END DATE */}
              <div className="space-y-1.5">
                <Label
                  htmlFor="period-end"
                  className="text-xs text-slate-600"
                >
                  End Date
                </Label>

                <Input
                  id="period-end"
                  type="date"
                  value={
                    selectedEndDate
                  }
                  onChange={(event) => {
                    if (
                      !event.target.value
                    ) {
                      setFilters({
                        period_end:
                          undefined,
                      });

                      return;
                    }

                    const selected =
                      new Date(
                        `${event.target.value}T00:00:00`,
                      );

                    /*
                     * Backend period_end is exclusive.
                     * Add one day so the selected end
                     * date is fully included.
                     */
                    selected.setDate(
                      selected.getDate() + 1,
                    );

                    setFilters({
                      period_end:
                        selected.toISOString(),
                    });
                  }}
                  className="h-10 border-slate-200 shadow-sm"
                />
              </div>

              {/* APPLY */}
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

              {/* DOWNLOADS */}
              <div className="flex h-10 gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={
                    handleDownloadPDF
                  }
                  disabled={
                    !data ||
                    downloadLoading !== null
                  }
                  className="flex-1 border-slate-200 bg-white text-slate-700 shadow-sm hover:bg-slate-50"
                >
                  {downloadLoading ===
                  "pdf" ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <FileText className="mr-2 h-4 w-4 text-red-600" />
                  )}

                  PDF
                </Button>

                <Button
                  type="button"
                  variant="outline"
                  onClick={
                    handleDownloadExcel
                  }
                  disabled={
                    !data ||
                    downloadLoading !== null
                  }
                  className="flex-1 border-slate-200 bg-white text-slate-700 shadow-sm hover:bg-slate-50"
                >
                  {downloadLoading ===
                  "excel" ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <FileSpreadsheet className="mr-2 h-4 w-4 text-green-600" />
                  )}

                  Excel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* SELECTED REPORT */}
        {summary && (
          <div className="rounded-xl border border-blue-100 bg-blue-50/50 px-4 py-3 text-sm text-slate-600">
            Showing report for{" "}
            <span className="font-semibold text-slate-900">
              {getSelectedBranchName()}
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

        {/* MAIN REVENUE METRICS */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {/* SESSION REVENUE */}
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
                <Monitor className="h-5 w-5" />
              </div>

              <div className="min-w-0">
                <p className="text-sm text-slate-500">
                  Session Revenue
                </p>

                <p className="text-2xl font-bold text-slate-950">
                  {formatCurrency(
                    summary?.session_revenue,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  {summary?.session_count ||
                    0} completed sessions
                </p>
              </div>
            </CardContent>
          </Card>

          {/* SALES REVENUE */}
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600">
                <ShoppingCart className="h-5 w-5" />
              </div>

              <div className="min-w-0">
                <p className="text-sm text-slate-500">
                  Sales Revenue
                </p>

                <p className="text-2xl font-bold text-slate-950">
                  {formatCurrency(
                    summary?.sales_revenue,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  {summary?.sales_count ||
                    0} completed sales
                </p>
              </div>
            </CardContent>
          </Card>

          {/* TOTAL REVENUE */}
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-green-50 text-green-600">
                <DollarSign className="h-5 w-5" />
              </div>

              <div className="min-w-0">
                <p className="text-sm text-slate-500">
                  Total Revenue
                </p>

                <p className="text-2xl font-bold text-green-600">
                  {formatCurrency(
                    summary?.total_revenue,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  Sessions + Sales
                </p>
              </div>
            </CardContent>
          </Card>

          {/* NET REVENUE */}
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-700">
                <Wallet className="h-5 w-5" />
              </div>

              <div className="min-w-0">
                <p className="text-sm text-slate-500">
                  Net Revenue
                </p>

                <p className="text-2xl font-bold text-slate-950">
                  {formatCurrency(
                    summary?.net_amount,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  Revenue - Expenses
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* COLLECTIONS / EXPENSES */}
        <div className="grid gap-4 sm:grid-cols-2">
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-green-50 text-green-600">
                <ArrowUpRight className="h-5 w-5" />
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Confirmed Collections
                </p>

                <p className="text-2xl font-bold text-green-600">
                  {formatCurrency(
                    summary?.total_collections,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  {summary?.payments_count ||
                    0} confirmed payments
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
                  Total Expenses
                </p>

                <p className="text-2xl font-bold text-red-600">
                  {formatCurrency(
                    summary?.total_expenses,
                  )}
                </p>

                <p className="mt-0.5 text-xs text-slate-400">
                  {summary?.expenses_count ||
                    0} expense records
                </p>
              </div>
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
                Loading report data...
              </div>
            ) : branchBreakdown.length === 0 ? (
              <div className="py-12 text-center">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-slate-100">
                  <Building2 className="h-6 w-6 text-slate-400" />
                </div>

                <h3 className="mt-4 font-semibold text-slate-900">
                  No Branch Data
                </h3>

                <p className="mt-1 text-sm text-slate-500">
                  No branch activity was found
                  for the selected period.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>
                        Branch
                      </TableHead>

                      <TableHead>
                        Sessions
                      </TableHead>

                      <TableHead>
                        Session Revenue
                      </TableHead>

                      <TableHead>
                        Sales
                      </TableHead>

                      <TableHead>
                        Sales Revenue
                      </TableHead>

                      <TableHead>
                        Total Revenue
                      </TableHead>

                      <TableHead>
                        Collections
                      </TableHead>

                      <TableHead>
                        Expenses
                      </TableHead>

                      <TableHead className="text-right">
                        Net Revenue
                      </TableHead>
                    </TableRow>
                  </TableHeader>

                  <TableBody>
                    {branchBreakdown.map(
                      (branch) => (
                        <TableRow
                          key={
                            branch.branch_id
                          }
                        >
                          <TableCell className="font-medium text-slate-900">
                            {
                              branch.branch_name
                            }
                          </TableCell>

                          <TableCell className="text-slate-600">
                            {
                              branch.session_count
                            }
                          </TableCell>

                          <TableCell className="font-medium text-[#0757B8]">
                            {formatCurrency(
                              branch.session_revenue,
                            )}
                          </TableCell>

                          <TableCell className="text-slate-600">
                            {
                              branch.sales_count
                            }
                          </TableCell>

                          <TableCell className="font-medium text-indigo-600">
                            {formatCurrency(
                              branch.sales_revenue,
                            )}
                          </TableCell>

                          <TableCell className="font-bold text-green-600">
                            {formatCurrency(
                              branch.total_revenue,
                            )}
                          </TableCell>

                          <TableCell className="font-medium text-green-600">
                            {formatCurrency(
                              branch.total_collections,
                            )}
                          </TableCell>

                          <TableCell className="font-medium text-red-600">
                            {formatCurrency(
                              branch.total_expenses,
                            )}
                          </TableCell>

                          <TableCell className="text-right font-bold text-slate-900">
                            {formatCurrency(
                              branch.net_amount,
                            )}
                          </TableCell>
                        </TableRow>
                      ),
                    )}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>

        {/* PAYMENT METHODS */}
        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg font-bold">
              <CreditCard className="h-5 w-5 text-[#0757B8]" />
              Payment Methods
            </CardTitle>
          </CardHeader>

          <CardContent>
            {paymentMethods.length ===
            0 ? (
              <div className="py-8 text-center text-sm text-slate-500">
                No confirmed payment
                transactions recorded for
                this period.
              </div>
            ) : (
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {paymentMethods.map(
                  (payment, index) => (
                    <div
                      key={`${payment.method}-${index}`}
                      className="flex items-center justify-between rounded-lg border border-slate-100 bg-slate-50/50 p-4"
                    >
                      <div>
                        <Badge className="border-0 bg-blue-100 text-[#0757B8] hover:bg-blue-100 capitalize">
                          {
                            payment.method
                          }
                        </Badge>

                        <p className="mt-2 text-xs text-slate-500">
                          {
                            payment.count
                          }{" "}
                          Transactions
                        </p>
                      </div>

                      <div className="text-right">
                        <p className="text-base font-bold text-slate-900">
                          {formatCurrency(
                            payment.amount,
                          )}
                        </p>
                      </div>
                    </div>
                  ),
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}
