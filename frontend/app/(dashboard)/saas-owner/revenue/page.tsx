
"use client";

import { useCallback, useEffect, useState } from "react";
import { RefreshCw, Search } from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import {
  platformService,
  type RevenueListItem,
} from "@/services/platform.service";

import { useAuthStore } from "@/store/auth.store";

const PAGE_SIZE = 20;

export default function RevenuePage() {
  const token = useAuthStore((state) => state.token);

  const [items, setItems] = useState<RevenueListItem[]>([]);
  const [search, setSearch] = useState("");
  const [activeSearch, setActiveSearch] = useState("");

  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadRevenue = useCallback(async () => {
    if (!token) {
      setError("You are not authenticated.");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      const response = await platformService.revenue(
        token,
        activeSearch,
        PAGE_SIZE,
        offset
      );

      setItems(response.items);
      setTotal(response.total);
    } catch (err) {
      console.error("PLATFORM REVENUE ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Failed to load platform revenue."
      );
    } finally {
      setLoading(false);
    }
  }, [token, activeSearch, offset]);

  useEffect(() => {
    loadRevenue();
  }, [loadRevenue]);

  function handleSearch() {
    setOffset(0);
    setActiveSearch(search.trim());
  }

  function handleClear() {
    setSearch("");
    setActiveSearch("");
    setOffset(0);
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

  function formatDate(date: string) {
    return new Date(date).toLocaleDateString("en-KE", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  function statusClass(status: string) {
    switch (status.toLowerCase()) {
      case "completed":
        return "bg-green-100 text-green-700";

      case "pending":
        return "bg-yellow-100 text-yellow-700";

      case "failed":
      case "cancelled":
        return "bg-red-100 text-red-700";

      default:
        return "bg-gray-100 text-gray-700";
    }
  }

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <DashboardShell role="platform_admin">
      <PageHeader
        title="Revenue"
        description="View CyberSaaS subscription payments and platform revenue."
      />

      <div className="space-y-6">
        <div className="rounded-xl border bg-white p-4 shadow-sm">
          <div className="flex flex-col gap-3 md:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />

              <input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") {
                    handleSearch();
                  }
                }}
                placeholder="Search owner, email, payment method or receipt..."
                className="h-10 w-full rounded-lg border border-gray-300 pl-10 pr-4 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <button
              type="button"
              onClick={handleSearch}
              className="h-10 rounded-lg bg-blue-600 px-5 text-sm font-medium text-white hover:bg-blue-700"
            >
              Search
            </button>

            {activeSearch && (
              <button
                type="button"
                onClick={handleClear}
                className="h-10 rounded-lg border border-gray-300 px-5 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Clear
              </button>
            )}

            <button
              type="button"
              onClick={loadRevenue}
              disabled={loading}
              className="inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-gray-300 px-4 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              <RefreshCw
                className={`h-4 w-4 ${
                  loading ? "animate-spin" : ""
                }`}
              />
              Refresh
            </button>
          </div>
        </div>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="overflow-hidden rounded-xl border bg-white shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[900px] text-sm">
              <thead className="border-b bg-gray-50">
                <tr>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Date
                  </th>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Cyber Owner
                  </th>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Email
                  </th>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Amount
                  </th>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    Method
                  </th>
                  <th className="px-5 py-4 text-left font-semibold text-blue-900">
                    M-Pesa Receipt
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
                      Loading revenue...
                    </td>
                  </tr>
                ) : items.length === 0 ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-5 py-10 text-center text-gray-500"
                    >
                      No revenue records found.
                    </td>
                  </tr>
                ) : (
                  items.map((item) => (
                    <tr
                      key={item.id}
                      className="hover:bg-gray-50"
                    >
                      <td className="px-5 py-4 text-gray-600">
                        {formatDate(item.created_at)}
                      </td>

                      <td className="px-5 py-4 font-medium text-gray-900">
                        {item.owner_name}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {item.owner_email || "—"}
                      </td>

                      <td className="px-5 py-4 font-semibold text-gray-900">
                        {formatAmount(item.amount)}
                      </td>

                      <td className="px-5 py-4 capitalize text-gray-600">
                        {item.payment_method.replace("_", " ")}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {item.mpesa_receipt_number || "—"}
                      </td>

                      <td className="px-5 py-4">
                        <span
                          className={`inline-flex rounded-full px-3 py-1 text-xs font-medium ${statusClass(
                            item.status
                          )}`}
                        >
                          {item.status.replace("_", " ")}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div className="flex flex-col gap-3 border-t px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-sm text-gray-500">
              {total === 0
                ? "Showing 0 payments"
                : `Showing ${offset + 1}–${Math.min(
                    offset + PAGE_SIZE,
                    total
                  )} of ${total} payments`}
            </p>

            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() =>
                  setOffset(Math.max(0, offset - PAGE_SIZE))
                }
                disabled={offset === 0 || loading}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                Previous
              </button>

              <span className="px-3 text-sm text-gray-600">
                Page {currentPage} of {totalPages}
              </span>

              <button
                type="button"
                onClick={() => setOffset(offset + PAGE_SIZE)}
                disabled={currentPage >= totalPages || loading}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
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
