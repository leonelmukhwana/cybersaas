
"use client";

import { useCallback, useEffect, useState } from "react";
import { Search, RefreshCw } from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import { platformService } from "@/services/platform.service";
import { useAuthStore } from "@/store/auth.store";

const PAGE_SIZE = 20;

export default function CyberOwnersPage() {
  const token = useAuthStore((state) => state.token);

  const [owners, setOwners] = useState<
    Awaited<ReturnType<typeof platformService.owners>>["items"]
  >([]);

  const [search, setSearch] = useState("");
  const [activeSearch, setActiveSearch] = useState("");

  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadOwners = useCallback(async () => {
    if (!token) {
      setError("You are not authenticated.");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      const response = await platformService.owners(
        token,
        activeSearch,
        PAGE_SIZE,
        offset
      );

      setOwners(response.items);
      setTotal(response.total);
    } catch (err) {
      console.error("CYBER OWNERS ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Failed to load Cyber Owners."
      );
    } finally {
      setLoading(false);
    }
  }, [token, activeSearch, offset]);

  useEffect(() => {
    loadOwners();
  }, [loadOwners]);

  function handleSearch() {
    setOffset(0);
    setActiveSearch(search.trim());
  }

  function handleClearSearch() {
    setSearch("");
    setActiveSearch("");
    setOffset(0);
  }

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  function goToPreviousPage() {
    if (offset === 0) return;

    setOffset(Math.max(0, offset - PAGE_SIZE));
  }

  function goToNextPage() {
    if (currentPage >= totalPages) return;

    setOffset(offset + PAGE_SIZE);
  }

  function formatAmount(amount: string) {
    const value = Number(amount);

    if (Number.isNaN(value)) {
      return `KES ${amount}`;
    }

    return `KES ${value.toLocaleString("en-KE", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })}`;
  }

  function formatDate(date?: string | null, lifetime?: boolean) {
    if (lifetime) {
      return "Lifetime";
    }

    if (!date) {
      return "—";
    }

    return new Date(date).toLocaleDateString("en-KE", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  function getStatusClass(status: string) {
    switch (status.toLowerCase()) {
      case "active":
        return "bg-green-100 text-green-700";

      case "past_due":
        return "bg-yellow-100 text-yellow-700";

      case "expired":
        return "bg-red-100 text-red-700";

      case "suspended":
        return "bg-red-100 text-red-700";

      default:
        return "bg-gray-100 text-gray-700";
    }
  }

  return (
    <DashboardShell role="platform_admin">
      <PageHeader
        title="Cyber Owners"
        description="View registered Cyber Owners and their current subscription information."
      />

      <div className="space-y-6">
        {/* Search */}
        <div className="rounded-xl border bg-white p-4 shadow-sm">
          <div className="flex flex-col gap-3 md:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />

              <input
                type="text"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") {
                    handleSearch();
                  }
                }}
                placeholder="Search by name, email or phone..."
                className="h-10 w-full rounded-lg border border-gray-300 pl-10 pr-4 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <button
              type="button"
              onClick={handleSearch}
              className="h-10 rounded-lg bg-blue-600 px-5 text-sm font-medium text-white transition hover:bg-blue-700"
            >
              Search
            </button>

            {activeSearch && (
              <button
                type="button"
                onClick={handleClearSearch}
                className="h-10 rounded-lg border border-gray-300 px-5 text-sm font-medium text-gray-700 transition hover:bg-gray-50"
              >
                Clear
              </button>
            )}

            <button
              type="button"
              onClick={loadOwners}
              disabled={loading}
              className="inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-gray-300 px-4 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <RefreshCw
                className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
              />
              Refresh
            </button>
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </div>
        )}

        {/* Table */}
        <div className="overflow-hidden rounded-xl border bg-white shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[950px] text-sm">
              <thead className="border-b bg-gray-50">
                <tr>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Name
                  </th>

                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Email
                  </th>

                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Phone
                  </th>

                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Package
                  </th>

                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Amount
                  </th>

                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Expiry
                  </th>

                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Status
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y">
                {loading ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-5 py-10 text-center text-gray-500"
                    >
                      Loading Cyber Owners...
                    </td>
                  </tr>
                ) : owners.length === 0 ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-5 py-10 text-center text-gray-500"
                    >
                      {activeSearch
                        ? "No Cyber Owners found."
                        : "No Cyber Owners registered yet."}
                    </td>
                  </tr>
                ) : (
                  owners.map((owner) => (
                    <tr
                      key={owner.id}
                      className="transition hover:bg-gray-50"
                    >
                      <td className="px-5 py-4 font-medium text-gray-900">
                        {owner.full_name}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {owner.email || "—"}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {owner.phone || "—"}
                      </td>

                      <td className="px-5 py-4 text-gray-700">
                        {owner.package_name || "No package"}
                      </td>

                      <td className="px-5 py-4 font-medium text-gray-900">
                        {owner.subscription_id
                          ? formatAmount(owner.amount)
                          : "—"}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {formatDate(
                          owner.subscription_expires_at,
                          owner.is_lifetime
                        )}
                      </td>

                      <td className="px-5 py-4">
                        <span
                          className={`inline-flex rounded-full px-3 py-1 text-xs font-medium ${getStatusClass(
                            owner.subscription_status || owner.status
                          )}`}
                        >
                          {(
                            owner.subscription_status || owner.status
                          ).replace("_", " ")}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="flex flex-col gap-3 border-t px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-sm text-gray-500">
              {total === 0
                ? "Showing 0 owners"
                : `Showing ${offset + 1}–${Math.min(
                    offset + PAGE_SIZE,
                    total
                  )} of ${total} owners`}
            </p>

            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={goToPreviousPage}
                disabled={offset === 0 || loading}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                Previous
              </button>

              <span className="px-3 text-sm text-gray-600">
                Page {currentPage} of {totalPages}
              </span>

              <button
                type="button"
                onClick={goToNextPage}
                disabled={currentPage >= totalPages || loading}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      </div>
    </DashboardShell>
  );
}
