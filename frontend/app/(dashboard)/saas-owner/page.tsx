"use client";

import { useEffect, useState } from "react";
import {
  Building2,
  CalendarDays,
  CreditCard,
  Users,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import StatCard from "@/components/dashboard/StatCard";

import { platformService } from "@/services/platform.service";
import { useAuthStore } from "@/store/auth.store";

import type { DashboardStats } from "@/services/platform.service";

export default function SaaSOwnerDashboardPage() {
  const token = useAuthStore((state) => state.token);

  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function loadDashboard() {
      if (typeof token !== "string" || token.length === 0) {
        setError("You are not authenticated.");
        setLoading(false);
        return;
      }

      try {
        const data = await platformService.dashboard(token);
        setStats(data);
      } catch (err) {
        console.error("PLATFORM DASHBOARD ERROR:", err);

        setError(
          err instanceof Error
            ? err.message
            : "Failed to load dashboard."
        );
      } finally {
        setLoading(false);
      }
    }

    loadDashboard();
  }, [token]);

  return (
    <DashboardShell role="platform_admin">
      <PageHeader
        title="SaaS Owner Dashboard"
        description="Monitor CyberSaaS owners, subscriptions, and platform revenue."
      />

      {loading && (
        <div className="mt-6 rounded-xl border border-slate-200 bg-white p-6">
          <p className="text-sm text-slate-500">
            Loading dashboard...
          </p>
        </div>
      )}

      {error && (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      {stats && (
        <>
          {/* PLATFORM STATISTICS */}
          <div className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <StatCard
              title="Cyber Owners"
              value={stats.total_cyber_owners}
              description={`${stats.active_cyber_owners} active`}
              icon={<Users className="h-5 w-5" />}
            />

            <StatCard
              title="Cyber Locations"
              value={stats.total_cyber_locations}
              description={`${stats.active_cyber_locations} active`}
              icon={<Building2 className="h-5 w-5" />}
            />

            <StatCard
              title="Active Subscriptions"
              value={stats.active_subscriptions}
              description={`${stats.past_due_subscriptions} past due`}
              icon={<CreditCard className="h-5 w-5" />}
            />

            <StatCard
              title="All-Time Revenue"
              value={`KES ${stats.subscription_revenue_all_time}`}
              description="Completed subscription payments"
              icon={<CreditCard className="h-5 w-5" />}
            />
          </div>

          {/* SUBSCRIPTION REVENUE */}
          <div className="mt-6 rounded-xl border border-slate-200 bg-white">
            <div className="border-b border-slate-200 px-6 py-5">
              <h2 className="text-lg font-semibold text-blue-900">
                Subscription Revenue
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Completed CyberSaaS subscription payments.
              </p>
            </div>

            <div className="grid gap-4 p-6 md:grid-cols-3">
              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-blue-50 p-2">
                    <CalendarDays className="h-5 w-5 text-blue-900" />
                  </div>

                  <div>
                    <p className="text-sm text-slate-500">
                      This Month
                    </p>

                    <p className="mt-1 text-2xl font-bold text-blue-900">
                      KES {stats.subscription_revenue_month}
                    </p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-blue-50 p-2">
                    <CalendarDays className="h-5 w-5 text-blue-900" />
                  </div>

                  <div>
                    <p className="text-sm text-slate-500">
                      This Year
                    </p>

                    <p className="mt-1 text-2xl font-bold text-blue-900">
                      KES {stats.subscription_revenue_year}
                    </p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-blue-50 p-2">
                    <CalendarDays className="h-5 w-5 text-blue-900" />
                  </div>

                  <div>
                    <p className="text-sm text-slate-500">
                      Last 3 Years
                    </p>

                    <p className="mt-1 text-2xl font-bold text-blue-900">
                      KES {stats.subscription_revenue_three_years}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* SUBSCRIPTION STATUS */}
          <div className="mt-6 rounded-xl border border-slate-200 bg-white">
            <div className="border-b border-slate-200 px-6 py-5">
              <h2 className="text-lg font-semibold text-blue-900">
                Subscription Overview
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Current platform subscription status.
              </p>
            </div>

            <div className="grid gap-6 p-6 sm:grid-cols-2 lg:grid-cols-4">
              <div>
                <p className="text-sm text-slate-500">
                  Active
                </p>

                <p className="mt-1 text-2xl font-bold text-blue-900">
                  {stats.active_subscriptions}
                </p>
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Past Due
                </p>

                <p className="mt-1 text-2xl font-bold text-blue-900">
                  {stats.past_due_subscriptions}
                </p>
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Expired
                </p>

                <p className="mt-1 text-2xl font-bold text-blue-900">
                  {stats.expired_subscriptions}
                </p>
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Lifetime
                </p>

                <p className="mt-1 text-2xl font-bold text-blue-900">
                  {stats.lifetime_subscriptions}
                </p>
              </div>
            </div>
          </div>

          {/* PLATFORM OVERVIEW */}
          <div className="mt-6 grid gap-6 lg:grid-cols-2">
            <div className="rounded-xl border border-slate-200 bg-white p-6">
              <h2 className="text-lg font-semibold text-blue-900">
                Cyber Owner Overview
              </h2>

              <div className="mt-5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-500">
                    Total owners
                  </span>

                  <span className="font-semibold text-blue-900">
                    {stats.total_cyber_owners}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-500">
                    Active owners
                  </span>

                  <span className="font-semibold text-blue-900">
                    {stats.active_cyber_owners}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-500">
                    Suspended owners
                  </span>

                  <span className="font-semibold text-blue-900">
                    {stats.suspended_cyber_owners}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-500">
                    Total locations
                  </span>

                  <span className="font-semibold text-blue-900">
                    {stats.total_cyber_locations}
                  </span>
                </div>
              </div>
            </div>

            <div className="rounded-xl border border-slate-200 bg-white p-6">
              <h2 className="text-lg font-semibold text-blue-900">
                Revenue Summary
              </h2>

              <div className="mt-5 space-y-5">
                <div>
                  <p className="text-sm text-slate-500">
                    This month
                  </p>

                  <p className="mt-1 text-xl font-semibold text-blue-900">
                    KES {stats.subscription_revenue_month}
                  </p>
                </div>

                <div>
                  <p className="text-sm text-slate-500">
                    This year
                  </p>

                  <p className="mt-1 text-xl font-semibold text-blue-900">
                    KES {stats.subscription_revenue_year}
                  </p>
                </div>

                <div>
                  <p className="text-sm text-slate-500">
                    Last 3 years
                  </p>

                  <p className="mt-1 text-xl font-semibold text-blue-900">
                    KES {stats.subscription_revenue_three_years}
                  </p>
                </div>

                <div>
                  <p className="text-sm text-slate-500">
                    All time
                  </p>

                  <p className="mt-1 text-xl font-semibold text-blue-900">
                    KES {stats.subscription_revenue_all_time}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </DashboardShell>
  );
}