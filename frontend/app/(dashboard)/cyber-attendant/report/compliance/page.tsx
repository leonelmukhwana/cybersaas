"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  CalendarDays,
  Download,
  Eye,
  FileSpreadsheet,
  FileText,
  RefreshCw,
  ShieldCheck,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import { Button } from "@/components/ui/button";

import { reportService } from "@/services/reportService";
import { useAuthStore } from "@/store/auth.store";
import type { Attendant } from "@/types/attendant";
import { attendantService } from "@/services/attendant.service";

import {
  downloadCompliancePDF,
  downloadComplianceExcel,
} from "@/lib/compliance-report-export";

interface ComplianceReport {
  id: string;
  tenant_id: string;
  branch_id?: string | null;
  generated_by?: string | null;
  report_type: string;
  period_start: string;
  period_end: string;
  generated_at: string;
  retention_until: string;
  snapshot: Record<string, unknown>;
  snapshot_hash?: string | null;
  created_at: string;
}

export default function ComplianceReportsPage() {
  const { token } = useAuthStore();

  const [attendant, setAttendant] =
    useState<Attendant | null>(null);

  const [reports, setReports] =
    useState<ComplianceReport[]>([]);

  const [selectedReport, setSelectedReport] =
    useState<ComplianceReport | null>(null);

  const [periodStart, setPeriodStart] =
    useState("");

  const [periodEnd, setPeriodEnd] =
    useState("");

  const [loadingAttendant, setLoadingAttendant] =
    useState(true);

  const [loadingReports, setLoadingReports] =
    useState(true);

  const [generating, setGenerating] =
    useState(false);

  const [error, setError] = useState("");

  /*
   * Load authenticated attendant.
   *
   * The attendant's branch comes from /api/attendants/me.
   * We NEVER load /api/branches here.
   */
  const loadAttendant = useCallback(async () => {
    if (!token) {
      setLoadingAttendant(false);
      return;
    }

    try {
      setLoadingAttendant(true);
      setError("");

      const result =
        await attendantService.me(token);

      setAttendant(result);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load attendant information."
      );
    } finally {
      setLoadingAttendant(false);
    }
  }, [token]);

  useEffect(() => {
    loadAttendant();
  }, [loadAttendant]);

  /*
   * The branch is determined by the authenticated attendant.
   */
  const branchId =
    attendant?.branch_id ?? "";

  const branchName =
    attendant?.branch_name ?? "";

  /*
   * Load previously generated reports.
   */
  const loadReports = useCallback(async () => {
    if (!token || !branchId) {
      setLoadingReports(false);
      return;
    }

    try {
      setLoadingReports(true);
      setError("");

      const response =
        await reportService.listComplianceReports(
          token,
          branchId
        );

      setReports(response.reports ?? []);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load compliance reports."
      );
    } finally {
      setLoadingReports(false);
    }
  }, [token, branchId]);

  useEffect(() => {
    if (!loadingAttendant) {
      loadReports();
    }
  }, [
    loadingAttendant,
    loadReports,
  ]);

  /*
   * Generate a CAK compliance report.
   */
  const handleGenerateReport = async () => {
    if (!token) {
      setError("You are not authenticated.");
      return;
    }

    if (!branchId) {
      setError(
        "Your attendant account does not have an active branch assignment."
      );
      return;
    }

    if (!periodStart || !periodEnd) {
      setError(
        "Please select both a start date and an end date."
      );
      return;
    }

    if (periodEnd < periodStart) {
      setError(
        "The end date cannot be before the start date."
      );
      return;
    }

    try {
      setGenerating(true);
      setError("");

      const response =
        await reportService.generateComplianceReport(
          token,
          {
            branch_id: branchId,
            report_type: "cak_compliance",
            period_start: periodStart,
            period_end: periodEnd,
          }
        );

      const generatedReport =
        response.snapshot;

      setSelectedReport(
        generatedReport
      );

      setReports((current) => [
        generatedReport,
        ...current.filter(
          (report) =>
            report.id !==
            generatedReport.id
        ),
      ]);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to generate compliance report."
      );
    } finally {
      setGenerating(false);
    }
  };

  /*
   * Open an existing report.
   */
  const handleViewReport = async (
    report: ComplianceReport
  ) => {
    if (!token) {
      setError("You are not authenticated.");
      return;
    }

    try {
      setError("");

      const response =
        await reportService.getComplianceReport(
          token,
          report.id
        );

      setSelectedReport(
        response.snapshot
      );
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load compliance report."
      );
    }
  };

  /*
   * Download a report as PDF or Excel.
   */
  const handleDownloadReport = (
    report: ComplianceReport
  ) => {
    try {
      setError("");

      downloadCompliancePDF(report);

      /*
       * Generate the Excel file as well.
       *
       * Both files are downloaded from the same
       * report snapshot.
       */
      downloadComplianceExcel(report);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to download compliance report."
      );
    }
  };

  /*
   * Download only PDF.
   */
  const handleDownloadPDF = (
    report: ComplianceReport
  ) => {
    try {
      setError("");
      downloadCompliancePDF(report);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to download PDF report."
      );
    }
  };

  /*
   * Download only Excel.
   */
  const handleDownloadExcel = (
    report: ComplianceReport
  ) => {
    try {
      setError("");
      downloadComplianceExcel(report);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to download Excel report."
      );
    }
  };

  /*
   * Format date.
   */
  const formatDate = (
    value?: string | null
  ) => {
    if (!value) {
      return "—";
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
      return value;
    }

    return date.toLocaleDateString(
      "en-KE",
      {
        year: "numeric",
        month: "short",
        day: "numeric",
      }
    );
  };

  /*
   * Format date/time.
   */
  const formatDateTime = (
    value?: string | null
  ) => {
    if (!value) {
      return "—";
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
      return value;
    }

    return date.toLocaleString(
      "en-KE",
      {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      }
    );
  };

  /*
   * Read statistics from the report snapshot.
   *
   * Supports:
   *
   * snapshot.records.*
   *
   * and the older:
   *
   * snapshot.*
   */
  const reportStats = useMemo(() => {
    if (!selectedReport?.snapshot) {
      return {
        customers: 0,
        terminals: 0,
        sessions: 0,
        sales: 0,
        payments: 0,
        receipts: 0,
        expenses: 0,
        auditLogs: 0,
      };
    }

    const snapshot =
      selectedReport.snapshot;

    const records =
      snapshot.records;

    const source =
      records &&
      typeof records === "object" &&
      !Array.isArray(records)
        ? (records as Record<
            string,
            unknown
          >)
        : snapshot;

    const getLength = (
      ...keys: string[]
    ) => {
      for (const key of keys) {
        const value =
          source[key];

        if (Array.isArray(value)) {
          return value.length;
        }
      }

      return 0;
    };

    return {
      customers: getLength(
        "customers"
      ),

      terminals: getLength(
        "terminals"
      ),

      sessions: getLength(
        "sessions"
      ),

      sales: getLength(
        "sales"
      ),

      payments: getLength(
        "payments"
      ),

      receipts: getLength(
        "receipts"
      ),

      expenses: getLength(
        "expenses"
      ),

      auditLogs: getLength(
        "audit_logs",
        "audit_events"
      ),
    };
  }, [selectedReport]);

  return (
    <DashboardShell role="attendant">
      <div className="space-y-6">
        <PageHeader
          title="CAK Compliance Reports"
          description={
            branchName
              ? `Compliance reports for ${branchName}.`
              : "View and generate CAK compliance reports."
          }
        />

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {/* Assigned branch */}
        <section className="rounded-xl border bg-white shadow-sm">
          <div className="flex items-center gap-3 px-6 py-5">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
              <ShieldCheck className="h-5 w-5" />
            </div>

            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                Branch
              </p>

              {loadingAttendant ? (
                <div className="mt-1 flex items-center gap-2 text-sm text-slate-500">
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  Loading branch...
                </div>
              ) : (
                <>
                  <h2 className="text-lg font-semibold text-slate-900">
                    {branchName ||
                      "Branch assignment unavailable"}
                  </h2>

                  <p className="text-sm text-slate-500">
                    Compliance records are scoped to this branch.
                  </p>
                </>
              )}
            </div>
          </div>
        </section>

        {/* Generate report */}
        <section className="rounded-xl border bg-white shadow-sm">
          <div className="border-b px-6 py-5">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
                <FileText className="h-5 w-5" />
              </div>

              <div>
                <h2 className="text-lg font-semibold text-slate-900">
                  Generate Compliance Report
                </h2>

                <p className="text-sm text-slate-500">
                  Select any historical period for this branch.
                </p>
              </div>
            </div>
          </div>

          <div className="grid gap-5 p-6 md:grid-cols-2">
            {/* From */}
            <div className="space-y-2">
              <label
                htmlFor="period-start"
                className="text-sm font-medium text-slate-700"
              >
                From
              </label>

              <div className="relative">
                <CalendarDays className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />

                <input
                  id="period-start"
                  type="date"
                  value={periodStart}
                  onChange={(event) =>
                    setPeriodStart(
                      event.target.value
                    )
                  }
                  disabled={generating}
                  className="h-10 w-full rounded-md border border-slate-300 bg-white pl-9 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:cursor-not-allowed disabled:bg-slate-50"
                />
              </div>
            </div>

            {/* To */}
            <div className="space-y-2">
              <label
                htmlFor="period-end"
                className="text-sm font-medium text-slate-700"
              >
                To
              </label>

              <div className="relative">
                <CalendarDays className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />

                <input
                  id="period-end"
                  type="date"
                  value={periodEnd}
                  onChange={(event) =>
                    setPeriodEnd(
                      event.target.value
                    )
                  }
                  disabled={generating}
                  className="h-10 w-full rounded-md border border-slate-300 bg-white pl-9 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:cursor-not-allowed disabled:bg-slate-50"
                />
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-3 border-t bg-slate-50 px-6 py-4 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-xs text-slate-500">
              Reports can cover any historical period,
              including multiple years.
            </p>

            <Button
              type="button"
              onClick={
                handleGenerateReport
              }
              disabled={
                generating ||
                loadingAttendant ||
                !branchId ||
                !periodStart ||
                !periodEnd
              }
              className="h-10 bg-blue-600 px-4 text-white hover:bg-blue-700"
            >
              {generating ? (
                <>
                  <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                  Generating...
                </>
              ) : (
                <>
                  <FileText className="mr-2 h-4 w-4" />
                  Generate Report
                </>
              )}
            </Button>
          </div>
        </section>

        {/* Existing reports */}
        <section className="rounded-xl border bg-white shadow-sm">
          <div className="flex flex-col gap-3 border-b px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-lg font-semibold text-slate-900">
                Generated Reports
              </h2>

              <p className="text-sm text-slate-500">
                Previously generated reports for{" "}
                {branchName ||
                  "this branch"}.
              </p>
            </div>

            <Button
              type="button"
              variant="outline"
              onClick={loadReports}
              disabled={
                loadingReports ||
                loadingAttendant ||
                !branchId
              }
              className="h-10 border-blue-200 text-blue-600 hover:bg-blue-50 hover:text-blue-700"
            >
              <RefreshCw
                className={`mr-2 h-4 w-4 ${
                  loadingReports
                    ? "animate-spin"
                    : ""
                }`}
              />
              Refresh
            </Button>
          </div>

          {loadingReports ? (
            <div className="flex items-center justify-center px-6 py-12">
              <div className="flex items-center gap-2 text-sm text-slate-500">
                <RefreshCw className="h-4 w-4 animate-spin" />
                Loading reports...
              </div>
            </div>
          ) : reports.length === 0 ? (
            <div className="flex flex-col items-center justify-center px-6 py-14 text-center">
              <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-blue-600">
                <FileText className="h-6 w-6" />
              </div>

              <h3 className="text-sm font-semibold text-slate-900">
                No generated reports
              </h3>

              <p className="mt-1 max-w-md text-sm text-slate-500">
                Generate a report above using the
                period you need for the CAK inspection.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[900px] text-sm">
                <thead>
                  <tr className="border-b bg-slate-50 text-left">
                    <th className="px-6 py-3 font-medium text-slate-600">
                      Period
                    </th>

                    <th className="px-6 py-3 font-medium text-slate-600">
                      Generated
                    </th>

                    <th className="px-6 py-3 font-medium text-slate-600">
                      Retention Until
                    </th>

                    <th className="px-6 py-3 text-right font-medium text-slate-600">
                      Actions
                    </th>
                  </tr>
                </thead>

                <tbody>
                  {reports.map((report) => (
                    <tr
                      key={report.id}
                      className="border-b last:border-0 hover:bg-slate-50"
                    >
                      <td className="px-6 py-4 text-slate-600">
                        <div>
                          {formatDate(
                            report.period_start
                          )}
                        </div>

                        <div className="text-xs text-slate-400">
                          to{" "}
                          {formatDate(
                            report.period_end
                          )}
                        </div>
                      </td>

                      <td className="px-6 py-4 text-slate-600">
                        {formatDateTime(
                          report.generated_at
                        )}
                      </td>

                      <td className="px-6 py-4 text-slate-600">
                        {formatDate(
                          report.retention_until
                        )}
                      </td>

                      <td className="px-6 py-4">
                        <div className="flex justify-end gap-2">
                          <Button
                            type="button"
                            variant="outline"
                            onClick={() =>
                              handleViewReport(
                                report
                              )
                            }
                            className="h-9 border-blue-200 px-3 text-blue-600 hover:bg-blue-50 hover:text-blue-700"
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            View
                          </Button>

                          <Button
                            type="button"
                            variant="outline"
                            onClick={() =>
                              handleDownloadPDF(
                                report
                              )
                            }
                            className="h-9 border-blue-200 px-3 text-blue-600 hover:bg-blue-50 hover:text-blue-700"
                          >
                            <FileText className="mr-2 h-4 w-4" />
                            PDF
                          </Button>

                          <Button
                            type="button"
                            variant="outline"
                            onClick={() =>
                              handleDownloadExcel(
                                report
                              )
                            }
                            className="h-9 border-blue-200 px-3 text-blue-600 hover:bg-blue-50 hover:text-blue-700"
                          >
                            <FileSpreadsheet className="mr-2 h-4 w-4" />
                            Excel
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {/* Selected report */}
        {selectedReport && (
          <section className="rounded-xl border bg-white shadow-sm">
            <div className="flex flex-col gap-4 border-b px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <ShieldCheck className="h-5 w-5 text-blue-600" />

                  <h2 className="text-lg font-semibold text-slate-900">
                    Report Preview
                  </h2>
                </div>

                <p className="mt-1 text-sm text-slate-500">
                  {branchName} •{" "}
                  {formatDate(
                    selectedReport.period_start
                  )}{" "}
                  –{" "}
                  {formatDate(
                    selectedReport.period_end
                  )}
                </p>
              </div>

              <div className="flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() =>
                    handleDownloadPDF(
                      selectedReport
                    )
                  }
                  className="h-10 border-blue-200 text-blue-600 hover:bg-blue-50 hover:text-blue-700"
                >
                  <FileText className="mr-2 h-4 w-4" />
                  Download PDF
                </Button>

                <Button
                  type="button"
                  variant="outline"
                  onClick={() =>
                    handleDownloadExcel(
                      selectedReport
                    )
                  }
                  className="h-10 border-blue-200 text-blue-600 hover:bg-blue-50 hover:text-blue-700"
                >
                  <FileSpreadsheet className="mr-2 h-4 w-4" />
                  Download Excel
                </Button>
              </div>
            </div>

            {/* Statistics */}
            <div className="grid gap-4 p-6 sm:grid-cols-2 lg:grid-cols-4">
              <div className="rounded-lg border bg-slate-50 p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Customers
                </p>

                <p className="mt-2 text-2xl font-semibold text-slate-900">
                  {reportStats.customers}
                </p>
              </div>

              <div className="rounded-lg border bg-slate-50 p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Terminals
                </p>

                <p className="mt-2 text-2xl font-semibold text-slate-900">
                  {reportStats.terminals}
                </p>
              </div>

              <div className="rounded-lg border bg-slate-50 p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Sessions
                </p>

                <p className="mt-2 text-2xl font-semibold text-slate-900">
                  {reportStats.sessions}
                </p>
              </div>

              <div className="rounded-lg border bg-slate-50 p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Receipts
                </p>

                <p className="mt-2 text-2xl font-semibold text-slate-900">
                  {reportStats.receipts}
                </p>
              </div>
            </div>

            {/* Additional statistics */}
            <div className="grid gap-4 px-6 pb-6 sm:grid-cols-2 lg:grid-cols-4">
              <div className="rounded-lg border p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Sales
                </p>

                <p className="mt-2 text-xl font-semibold text-slate-900">
                  {reportStats.sales}
                </p>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Payments
                </p>

                <p className="mt-2 text-xl font-semibold text-slate-900">
                  {reportStats.payments}
                </p>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Expenses
                </p>

                <p className="mt-2 text-xl font-semibold text-slate-900">
                  {reportStats.expenses}
                </p>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Audit Logs
                </p>

                <p className="mt-2 text-xl font-semibold text-slate-900">
                  {reportStats.auditLogs}
                </p>
              </div>
            </div>

            {/* Report information */}
            <div className="border-t px-6 py-5">
              <h3 className="mb-4 text-sm font-semibold text-slate-900">
                Report Information
              </h3>

              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                <div>
                  <p className="text-xs text-slate-500">
                    Report ID
                  </p>

                  <p className="mt-1 break-all text-sm font-medium text-slate-900">
                    {selectedReport.id}
                  </p>
                </div>

                <div>
                  <p className="text-xs text-slate-500">
                    Branch
                  </p>

                  <p className="mt-1 text-sm font-medium text-slate-900">
                    {branchName || "—"}
                  </p>
                </div>

                <div>
                  <p className="text-xs text-slate-500">
                    Report Type
                  </p>

                  <p className="mt-1 text-sm font-medium text-slate-900">
                    CAK Compliance
                  </p>
                </div>

                <div>
                  <p className="text-xs text-slate-500">
                    Generated
                  </p>

                  <p className="mt-1 text-sm font-medium text-slate-900">
                    {formatDateTime(
                      selectedReport.generated_at
                    )}
                  </p>
                </div>

                <div>
                  <p className="text-xs text-slate-500">
                    Retention Until
                  </p>

                  <p className="mt-1 text-sm font-medium text-slate-900">
                    {formatDate(
                      selectedReport.retention_until
                    )}
                  </p>
                </div>

                <div>
                  <p className="text-xs text-slate-500">
                    Generated By
                  </p>

                  <p className="mt-1 break-all text-sm font-medium text-slate-900">
                    {selectedReport.generated_by ??
                      "—"}
                  </p>
                </div>
              </div>
            </div>

            {/* Compliance scope */}
            <div className="border-t bg-slate-50 px-6 py-5">
              <h3 className="text-sm font-semibold text-slate-900">
                Compliance Scope
              </h3>

              <ul className="mt-3 space-y-2 text-sm text-slate-600">
                <li>
                  • Customer identification records
                </li>

                <li>
                  • Terminal/computer identification
                </li>

                <li>
                  • Session start and end information
                </li>

                <li>
                  • Sales and payment records
                </li>

                <li>
                  • Receipt records and reprints
                </li>

                <li>
                  • Branch expenses
                </li>

                <li>
                  • Relevant branch audit records
                </li>

                <li>
                  • Historical records within the selected period
                </li>

                <li>
                  • Browsing history is not included
                </li>
              </ul>
            </div>

            {/* Raw snapshot */}
            <details className="border-t">
              <summary className="cursor-pointer px-6 py-4 text-sm font-medium text-slate-700 hover:bg-slate-50">
                View Raw Report Snapshot
              </summary>

              <div className="border-t bg-slate-950 p-6">
                <pre className="max-h-[500px] overflow-auto whitespace-pre-wrap break-words text-xs text-slate-200">
                  {JSON.stringify(
                    selectedReport.snapshot,
                    null,
                    2
                  )}
                </pre>
              </div>
            </details>
          </section>
        )}
      </div>
    </DashboardShell>
  );
}