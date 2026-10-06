"use client";

import { useCallback, useEffect, useState } from "react";
import { Search, ChevronLeft, ChevronRight } from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import {
  platformService,
  type SubscriptionListItem,
} from "@/services/platform.service";

import { useAuthStore } from "@/store/auth.store";

const PAGE_SIZE = 20;

function formatAmount(amount: string) {
  return `KES ${Number(amount || 0).toLocaleString()}`;
}

function formatDate(value?: string | null) {
  if (!value) return "—";

  return new Date(value).toLocaleDateString("en-KE", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export default function SubscriptionsPage() {
  const token = useAuthStore((state) => state.token);

  const [subscriptions, setSubscriptions] = useState<
    SubscriptionListItem[]
  >([]);

  const [total, setTotal] = useState(0);

  const [search, setSearch] = useState("");
  const [activeSearch, setActiveSearch] = useState("");

  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const totalPages = Math.max(
    1,
    Math.ceil(total / PAGE_SIZE)
  );

  const loadSubscriptions = useCallback(async () => {
    if (typeof token !== "string" || !token) {
      setError("You are not authenticated.");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      const response =
        await platformService.subscriptions(
          token,
          activeSearch,
          PAGE_SIZE,
          (page - 1) * PAGE_SIZE
        );

      setSubscriptions(response.items);
      setTotal(response.total);
    } catch (err) {
      console.error("SUBSCRIPTIONS ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Failed to load subscriptions."
      );
    } finally {
      setLoading(false);
    }
  }, [token, activeSearch, page]);

  useEffect(() => {
    loadSubscriptions();
  }, [loadSubscriptions]);

  function handleSearch(event: React.FormEvent) {
    event.preventDefault();

    setPage(1);
    setActiveSearch(search);
  }

  return (
    <DashboardShell role="platform_admin">
      <PageHeader
        title="Current Subscriptions"
        description="View active CyberSaaS subscriptions and their current billing periods."
      />

      <div className="space-y-6">
        <form
          onSubmit={handleSearch}
          className="flex gap-3"
        >
          <div className="relative max-w-md flex-1">
            <Search
              size={18}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
            />

            <input
              value={search}
              onChange={(event) =>
                setSearch(event.target.value)
              }
              placeholder="Search owner or package..."
              className="h-11 w-full rounded-lg border border-gray-300 bg-white pl-10 pr-4 text-sm outline-none focus:border-blue-700 focus:ring-2 focus:ring-blue-100"
            />
          </div>

          <button
            type="submit"
            className="rounded-lg bg-blue-700 px-5 text-sm font-semibold text-white hover:bg-blue-800"
          >
            Search
          </button>
        </form>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[950px] text-sm">
              <thead className="border-b bg-gray-50">
                <tr className="text-left text-blue-900">
                  <th className="px-5 py-4 font-semibold">
                    Owner
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Email
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Package
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Amount
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Start
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Expiry
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Status
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y">
                {loading ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-5 py-12 text-center text-gray-500"
                    >
                      Loading subscriptions...
                    </td>
                  </tr>
                ) : subscriptions.length === 0 ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-5 py-12 text-center text-gray-500"
                    >
                      No current subscriptions found.
                    </td>
                  </tr>
                ) : (
                  subscriptions.map((subscription) => (
                    <tr
                      key={subscription.id}
                      className="hover:bg-blue-50/40"
                    >
                      <td className="px-5 py-4 font-medium text-blue-900">
                        {subscription.owner_name}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {subscription.owner_email || "—"}
                      </td>

                      <td className="px-5 py-4 font-medium text-gray-800">
                        {subscription.package_name}
                      </td>

                      <td className="px-5 py-4 font-medium text-gray-800">
                        {formatAmount(subscription.amount)}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {formatDate(
                          subscription.current_period_start
                        )}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {subscription.is_lifetime
                          ? "Lifetime"
                          : formatDate(
                              subscription.current_period_end
                            )}
                      </td>

                      <td className="px-5 py-4">
                        <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-semibold text-green-700">
                          Active
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between border-t px-5 py-4">
            <p className="text-sm text-gray-500">
              {total === 0
                ? "0 results"
                : `${(page - 1) * PAGE_SIZE + 1}–${Math.min(
                    page * PAGE_SIZE,
                    total
                  )} of ${total}`}
            </p>

            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() =>
                  setPage((value) => value - 1)
                }
                className="rounded-lg border p-2 text-gray-600 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                <ChevronLeft size={18} />
              </button>

              <span className="min-w-[90px] text-center text-sm font-medium text-blue-900">
                Page {page} of {totalPages}
              </span>

              <button
                disabled={page >= totalPages}
                onClick={() =>
                  setPage((value) => value + 1)
                }
                className="rounded-lg border p-2 text-gray-600 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                <ChevronRight size={18} />
              </button>
            </div>
          </div>
        </div>
      </div>
    </DashboardShell>
  );
}