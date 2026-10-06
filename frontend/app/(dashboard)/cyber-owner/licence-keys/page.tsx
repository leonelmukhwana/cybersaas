"use client";

import {
  Check,
  Clipboard,
  Eye,
  EyeOff,
  KeyRound,
  Loader2,
  Plus,
  RefreshCw,
  ShieldCheck,
  Trash2,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import { branchService } from "@/services/branch.service";
import { terminalService } from "@/services/terminal.service";
import { useAuthStore } from "@/store/auth.store";

import type { Branch } from "@/types/branch";
import type {
  GenerateLicenceKeyResponse,
  LicenceKey,
} from "@/types/terminal";

export default function LicenceKeysPage() {
  const token = useAuthStore((state) => state.token);

  const [branches, setBranches] = useState<Branch[]>([]);
  const [licenceKeys, setLicenceKeys] = useState<LicenceKey[]>([]);

  const [selectedBranchId, setSelectedBranchId] = useState("");
  const [expiresInHours, setExpiresInHours] = useState("24");

  const [generatedKey, setGeneratedKey] =
    useState<GenerateLicenceKeyResponse | null>(null);

  const [showKey, setShowKey] = useState(true);
  const [copied, setCopied] = useState(false);

  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [revokingId, setRevokingId] = useState<string | null>(null);

  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const activeBranches = useMemo(
    () =>
      branches.filter(
        (branch) => branch.status.toLowerCase() === "active",
      ),
    [branches],
  );

  async function loadData() {
    if (typeof token !== "string" || token.length === 0) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      const authToken = token;

      const [branchResponse, licenceResponse] =
        await Promise.all([
          branchService.list(authToken),
          terminalService.listLicenceKeys(100, 0),
        ]);

      setBranches(branchResponse.branches ?? []);
      setLicenceKeys(licenceResponse.licence_keys ?? []);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to load licence keys.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadData();
  }, [token]);

  async function handleGenerate() {
    if (!selectedBranchId) {
      setError("Please select a branch.");
      return;
    }

    if (typeof token !== "string" || token.length === 0) {
      setError("Your session has expired. Please log in again.");
      return;
    }

    try {
      setGenerating(true);
      setError("");
      setSuccess("");
      setCopied(false);
      setShowKey(true);

      const hours = Number(expiresInHours);

      const response =
        await terminalService.generateLicenceKey({
          branch_id: selectedBranchId,
          ...(Number.isFinite(hours) && hours > 0
            ? { expires_in_hours: hours }
            : {}),
        });

      /*
       * IMPORTANT:
       * Store the COMPLETE API response here.
       * The licence_key property is the actual key returned
       * by the backend.
       */
      setGeneratedKey(response);

      setSuccess("Licence key generated successfully.");

      await loadData();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to generate licence key.",
      );
    } finally {
      setGenerating(false);
    }
  }

  async function handleCopyKey() {
    if (!generatedKey?.licence_key) {
      setError("There is no licence key to copy.");
      return;
    }

    try {
      await navigator.clipboard.writeText(
        generatedKey.licence_key,
      );

      setCopied(true);

      window.setTimeout(() => {
        setCopied(false);
      }, 2000);
    } catch {
      setError(
        "Unable to copy the licence key. Please copy it manually.",
      );
    }
  }

  async function handleRevoke(id: string) {
    const confirmed = window.confirm(
      "Are you sure you want to revoke this licence key? It will no longer be usable for terminal registration.",
    );

    if (!confirmed) return;

    try {
      setRevokingId(id);
      setError("");
      setSuccess("");

      await terminalService.revokeLicenceKey(id);

      setLicenceKeys((current) =>
        current.map((key) =>
          key.id === id
            ? {
                ...key,
                revoked_at: new Date().toISOString(),
              }
            : key,
        ),
      );

      setSuccess("Licence key revoked successfully.");
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to revoke licence key.",
      );
    } finally {
      setRevokingId(null);
    }
  }

  function formatDate(value?: string | null) {
    if (!value) return "—";

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
      return value;
    }

    return date.toLocaleString();
  }

  function getKeyStatus(key: LicenceKey) {
    if (key.revoked_at) {
      return {
        label: "Revoked",
        className:
          "bg-red-50 text-red-700 ring-red-200",
      };
    }

    if (key.used_at) {
      return {
        label: "Used",
        className:
          "bg-blue-50 text-blue-700 ring-blue-200",
      };
    }

    if (key.expires_at) {
      const expiresAt = new Date(key.expires_at);

      if (
        !Number.isNaN(expiresAt.getTime()) &&
        expiresAt.getTime() <= Date.now()
      ) {
        return {
          label: "Expired",
          className:
            "bg-amber-50 text-amber-700 ring-amber-200",
        };
      }
    }

    return {
      label: "Available",
      className:
        "bg-green-50 text-green-700 ring-green-200",
    };
  }

  return (
    <DashboardShell role="owner">
      <PageHeader
        title="Licence Keys"
        description="Generate and manage terminal registration keys for your branches."
      />

      {error && (
        <div className="mb-6 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-6 rounded-xl border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
          {success}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Generate */}
        <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-6 flex items-start gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-50 text-blue-600">
              <KeyRound className="h-5 w-5" />
            </div>

            <div>
              <h2 className="text-lg font-semibold text-slate-900">
                Generate Licence Key
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Generate a registration key for a terminal in one of
                your branches.
              </p>
            </div>
          </div>

          <div className="space-y-5">
            <div>
              <label
                htmlFor="branch"
                className="mb-2 block text-sm font-medium text-slate-700"
              >
                Branch
              </label>

              <select
                id="branch"
                value={selectedBranchId}
                onChange={(event) =>
                  setSelectedBranchId(event.target.value)
                }
                disabled={
                  loading ||
                  generating ||
                  activeBranches.length === 0
                }
                className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
              >
                <option value="">
                  Select a branch
                </option>

                {activeBranches.map((branch) => (
                  <option
                    key={branch.id}
                    value={branch.id}
                  >
                    {branch.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label
                htmlFor="expiry"
                className="mb-2 block text-sm font-medium text-slate-700"
              >
                Key validity
              </label>

              <select
                id="expiry"
                value={expiresInHours}
                onChange={(event) =>
                  setExpiresInHours(event.target.value)
                }
                disabled={generating}
                className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
              >
                <option value="1">1 hour</option>
                <option value="6">6 hours</option>
                <option value="12">12 hours</option>
                <option value="24">24 hours</option>
                <option value="48">48 hours</option>
                <option value="72">72 hours</option>
                <option value="168">7 days</option>
              </select>
            </div>

            <button
              type="button"
              onClick={handleGenerate}
              disabled={
                generating ||
                loading ||
                !selectedBranchId ||
                activeBranches.length === 0
              }
              className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {generating ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Generating...
                </>
              ) : (
                <>
                  <Plus className="h-4 w-4" />
                  Generate Licence Key
                </>
              )}
            </button>
          </div>
        </section>

        {/* Generated key */}
        <section className="rounded-2xl border-2 border-blue-200 bg-white p-6 shadow-sm">
          <div className="mb-5 flex items-start gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-600 text-white">
              <ShieldCheck className="h-5 w-5" />
            </div>

            <div>
              <h2 className="text-lg font-semibold text-slate-900">
                Your Licence Key
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Use this key when registering the terminal.
              </p>
            </div>
          </div>

          {generatedKey?.licence_key ? (
            <div className="space-y-4">
              <div className="rounded-xl border border-blue-200 bg-blue-50 p-4">
                <div className="mb-2 flex items-center justify-between">
                  <span className="text-xs font-semibold uppercase tracking-wider text-blue-700">
                    Licence Key
                  </span>

                  <span className="rounded-full bg-green-100 px-2.5 py-1 text-xs font-medium text-green-700">
                    Ready
                  </span>
                </div>

                <div className="flex items-center gap-2">
                  <div className="min-w-0 flex-1 overflow-hidden rounded-lg border border-slate-200 bg-white px-4 py-4">
                    <p
                      className={`break-all font-mono text-base font-bold tracking-wide text-slate-900 ${
                        !showKey
                          ? "select-none"
                          : ""
                      }`}
                    >
                      {showKey
                        ? generatedKey.licence_key
                        : "••••••••••••••••••••••••"}
                    </p>
                  </div>

                  <button
                    type="button"
                    onClick={() =>
                      setShowKey((current) => !current)
                    }
                    title={
                      showKey
                        ? "Hide licence key"
                        : "View licence key"
                    }
                    aria-label={
                      showKey
                        ? "Hide licence key"
                        : "View licence key"
                    }
                    className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg border border-slate-300 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-blue-600"
                  >
                    {showKey ? (
                      <EyeOff className="h-5 w-5" />
                    ) : (
                      <Eye className="h-5 w-5" />
                    )}
                  </button>

                  <button
                    type="button"
                    onClick={handleCopyKey}
                    title="Copy licence key"
                    aria-label="Copy licence key"
                    className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-blue-600 text-white transition hover:bg-blue-700"
                  >
                    {copied ? (
                      <Check className="h-5 w-5" />
                    ) : (
                      <Clipboard className="h-5 w-5" />
                    )}
                  </button>
                </div>

                <p className="mt-3 text-xs text-blue-700">
                  {copied
                    ? "Licence key copied to clipboard."
                    : "Click the eye to view/hide the key or the copy icon to copy it."}
                </p>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-lg bg-slate-50 p-3">
                  <p className="text-xs text-slate-500">
                    Branch
                  </p>
                  <p className="mt-1 text-sm font-medium text-slate-900">
                    {generatedKey.branch_name ??
                      generatedKey.branch_id}
                  </p>
                </div>

                <div className="rounded-lg bg-slate-50 p-3">
                  <p className="text-xs text-slate-500">
                    Expires
                  </p>
                  <p className="mt-1 text-sm font-medium text-slate-900">
                    {formatDate(
                      generatedKey.expires_at,
                    )}
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex min-h-[230px] items-center justify-center rounded-xl border-2 border-dashed border-slate-300 bg-slate-50 px-6 text-center">
              <div>
                <KeyRound className="mx-auto mb-3 h-10 w-10 text-slate-400" />

                <p className="text-sm font-semibold text-slate-700">
                  No licence key generated yet
                </p>

                <p className="mt-1 text-xs text-slate-500">
                  Generate a key and it will appear here immediately.
                </p>
              </div>
            </div>
          )}
        </section>
      </div>

      {/* History */}
      <section className="mt-6 rounded-2xl border border-slate-200 bg-white shadow-sm">
        <div className="flex flex-col gap-4 border-b border-slate-200 px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-slate-900">
              Licence Key History
            </h2>

            <p className="mt-1 text-sm text-slate-500">
              Keys generated for your branches.
            </p>
          </div>

          <button
            type="button"
            onClick={loadData}
            disabled={loading}
            className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-2.5 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
          >
            <RefreshCw
              className={`h-4 w-4 ${
                loading ? "animate-spin" : ""
              }`}
            />
            Refresh
          </button>
        </div>

        {loading ? (
          <div className="flex min-h-[180px] items-center justify-center">
            <Loader2 className="h-6 w-6 animate-spin text-blue-600" />
          </div>
        ) : licenceKeys.length === 0 ? (
          <div className="px-6 py-12 text-center">
            <KeyRound className="mx-auto mb-3 h-8 w-8 text-slate-400" />

            <p className="text-sm font-medium text-slate-700">
              No licence keys yet
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[800px]">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-left">
                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Branch
                  </th>

                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Status
                  </th>

                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Created
                  </th>

                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Expires
                  </th>

                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Used
                  </th>

                  <th className="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Action
                  </th>
                </tr>
              </thead>

              <tbody>
                {licenceKeys.map((key) => {
                  const status = getKeyStatus(key);

                  return (
                    <tr
                      key={key.id}
                      className="border-b border-slate-100 last:border-0"
                    >
                      <td className="px-6 py-4">
                        <span className="font-medium text-slate-900">
                          {key.branch_name ??
                            key.branch_id}
                        </span>
                      </td>

                      <td className="px-6 py-4">
                        <span
                          className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ring-1 ring-inset ${status.className}`}
                        >
                          {status.label}
                        </span>
                      </td>

                      <td className="px-6 py-4 text-sm text-slate-600">
                        {formatDate(key.created_at)}
                      </td>

                      <td className="px-6 py-4 text-sm text-slate-600">
                        {formatDate(key.expires_at)}
                      </td>

                      <td className="px-6 py-4 text-sm text-slate-600">
                        {key.used_at
                          ? formatDate(key.used_at)
                          : "Not used"}
                      </td>

                      <td className="px-6 py-4 text-right">
                        {!key.revoked_at &&
                          !key.used_at && (
                            <button
                              type="button"
                              onClick={() =>
                                handleRevoke(key.id)
                              }
                              disabled={
                                revokingId === key.id
                              }
                              className="inline-flex items-center gap-2 rounded-lg border border-red-200 bg-white px-3 py-2 text-xs font-medium text-red-600 hover:bg-red-50 disabled:opacity-50"
                            >
                              {revokingId === key.id ? (
                                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                              ) : (
                                <Trash2 className="h-3.5 w-3.5" />
                              )}
                              Revoke
                            </button>
                          )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </DashboardShell>
  );
}