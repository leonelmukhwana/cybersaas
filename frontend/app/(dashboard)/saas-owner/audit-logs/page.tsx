"use client";

import { useCallback, useEffect, useState } from "react";
import { Search, ChevronLeft, ChevronRight } from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import {
  platformService,
  type AuditLog,
} from "@/services/platform.service";

import { useAuthStore } from "@/store/auth.store";

const PAGE_SIZE = 20;

function formatDate(value: string) {
  return new Date(value).toLocaleString("en-KE", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function AuditLogsPage() {
  const token = useAuthStore((state) => state.token);

  const [logs, setLogs] = useState<AuditLog[]>([]);
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

  const loadLogs = useCallback(async () => {
    if (typeof token !== "string" || !token) {
      setError("You are not authenticated.");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      const response =
        await platformService.auditLogs(
          token,
          activeSearch,
          PAGE_SIZE,
          (page - 1) * PAGE_SIZE
        );

      setLogs(response.items);
      setTotal(response.total);
    } catch (err) {
      console.error("AUDIT LOGS ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Failed to load audit logs."
      );
    } finally {
      setLoading(false);
    }
  }, [token, activeSearch, page]);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  function handleSearch(event: React.FormEvent) {
    event.preventDefault();

    setPage(1);
    setActiveSearch(search);
  }

  return (
    <DashboardShell role="platform_admin">
      <PageHeader
        title="Audit Logs"
        description="Review administrative actions and platform activity."
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
              placeholder="Search user, action, entity or reason..."
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
            <table className="w-full min-w-[1000px] text-sm">
              <thead className="border-b bg-gray-50">
                <tr className="text-left text-blue-900">
                  <th className="px-5 py-4 font-semibold">
                    Date
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    User
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Role
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Action
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Entity
                  </th>
                  <th className="px-5 py-4 font-semibold">
                    Reason
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y">
                {loading ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-5 py-12 text-center text-gray-500"
                    >
                      Loading audit logs...
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-5 py-12 text-center text-gray-500"
                    >
                      No audit logs found.
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <tr
                      key={log.id}
                      className="hover:bg-blue-50/40"
                    >
                      <td className="whitespace-nowrap px-5 py-4 text-gray-600">
                        {formatDate(log.created_at)}
                      </td>

                      <td className="px-5 py-4 font-medium text-blue-900">
                        {log.user_name || "System"}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {log.user_role || "—"}
                      </td>

                      <td className="px-5 py-4 font-medium text-gray-800">
                        {log.action}
                      </td>

                      <td className="px-5 py-4 text-gray-600">
                        {log.entity_type || "—"}
                      </td>

                      <td className="max-w-[300px] px-5 py-4 text-gray-600">
                        <span className="line-clamp-2">
                          {log.reason || "—"}
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