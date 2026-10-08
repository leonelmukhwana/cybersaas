"use client";

import {
  Activity,
  ArrowRight,
  Building2,
  CreditCard,
  KeyRound,
  Laptop,
  Receipt,
  UserRound,
  Users,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import StatCard from "@/components/dashboard/StatCard";

import { branchService } from "@/services/branch.service";
import { attendantService } from "@/services/attendant.service";
import { subscriptionService } from "@/services/subscription.service";
import { terminalService } from "@/services/terminal.service";

import { useAuthStore } from "@/store/auth.store";

export default function OwnerDashboard() {
  const { token } = useAuthStore();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [branches, setBranches] = useState<
    {
      id: string;
      name: string;
      address?: string | null;
      status: string;
    }[]
  >([]);

  const [attendantCount, setAttendantCount] = useState(0);
  const [terminalCount, setTerminalCount] = useState(0);

  const [subscription, setSubscription] = useState<
    Awaited<ReturnType<typeof subscriptionService.getOverview>> | null
  >(null);

  useEffect(() => {
    if (typeof token !== "string" || token.length === 0) {
      setLoading(false);
      return;
    }

    const authToken = token;
    let mounted = true;

    async function loadDashboard() {
      try {
        setLoading(true);
        setError("");

        const [
          branchResponse,
          attendantResponse,
          terminalResponse,
          subscriptionResponse,
        ] = await Promise.all([
          branchService.list(authToken),
          attendantService.list(),
          terminalService.list(),
          subscriptionService.getOverview(authToken),
        ]);

        if (!mounted) return;

        setBranches(branchResponse.branches ?? []);

        setAttendantCount(
          attendantResponse.attendants?.length ??
            attendantResponse.total ??
            0
        );

        setTerminalCount(
          terminalResponse.terminals?.length ??
            terminalResponse.total ??
            0
        );

        setSubscription(subscriptionResponse);
      } catch (err) {
        if (!mounted) return;

        setError(
          err instanceof Error
            ? err.message
            : "Unable to load dashboard data."
        );
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    }

    loadDashboard();

    return () => {
      mounted = false;
    };
  }, [token]);

  const activeBranches = useMemo(
    () =>
      branches.filter(
        (branch) => branch.status.toLowerCase() === "active"
      ).length,
    [branches]
  );

  const subscriptionLabel = useMemo(() => {
    if (!subscription) return "—";

    if (subscription.is_lifetime) {
      return "Lifetime";
    }

    if (subscription.is_trial) {
      return subscription.is_trial_expired
        ? "Trial expired"
        : "Trial";
    }

    if (subscription.is_active) {
      return subscription.plan?.name ?? "Active";
    }

    if (subscription.is_expired) {
      return "Expired";
    }

    if (subscription.must_select_package) {
      return "Select plan";
    }

    return subscription.subscription?.status ?? "Inactive";
  }, [subscription]);

  const subscriptionDescription = useMemo(() => {
    if (!subscription) {
      return "Subscription information";
    }

    if (subscription.is_lifetime) {
      return "Lifetime subscription";
    }

    if (subscription.must_select_package) {
      return "Choose a subscription plan";
    }

    if (subscription.is_active) {
      return `${subscription.branches_used}/${subscription.branches_limit} branches · ${subscription.terminals_used}/${subscription.terminals_limit} terminals`;
    }

    if (subscription.is_expired) {
      return "Subscription has expired";
    }

    return "Subscription status";
  }, [subscription]);

  return (
    <DashboardShell role="owner">
      <PageHeader
        title="Dashboard"
        description="Manage your cyber café business."
      />

      {error && (
        <div className="mb-6 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* STATS */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Branches"
          value={loading ? "…" : String(branches.length)}
          description={
            loading
              ? "Loading branches"
              : `${activeBranches} active`
          }
          icon={<Building2 className="h-5 w-5" />}
        />

        <StatCard
          title="Attendants"
          value={loading ? "…" : String(attendantCount)}
          description="Registered attendants"
          icon={<Users className="h-5 w-5" />}
        />

        <StatCard
          title="Terminals"
          value={loading ? "…" : String(terminalCount)}
          description="Registered computers"
          icon={<Laptop className="h-5 w-5" />}
        />

        <StatCard
          title="Subscription"
          value={loading ? "…" : subscriptionLabel}
          description={subscriptionDescription}
          icon={<CreditCard className="h-5 w-5" />}
        />
      </div>

      {/* MAIN GRID */}
      <div className="mt-6 grid gap-6 lg:grid-cols-3">
        {/* BRANCHES */}
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm lg:col-span-2">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h2 className="text-base font-semibold text-slate-950">
                Your Branches
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Manage the locations registered under your business.
              </p>
            </div>

            <Link
              href="/cyber-owner/branch"
              className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-blue-600 hover:text-blue-700"
            >
              View all
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>

          <div className="mt-6 space-y-3">
            {loading ? (
              <>
                <div className="h-16 animate-pulse rounded-lg bg-slate-100" />
                <div className="h-16 animate-pulse rounded-lg bg-slate-100" />
              </>
            ) : branches.length === 0 ? (
              <div className="flex min-h-32 flex-col items-center justify-center rounded-lg border border-dashed border-slate-200 px-6 text-center">
                <Building2 className="h-8 w-8 text-slate-300" />

                <p className="mt-3 text-sm font-medium text-slate-700">
                  No branches yet
                </p>

                <p className="mt-1 text-sm text-slate-400">
                  Create your first cyber café branch to get started.
                </p>

                <Link
                  href="/dashboard/cyber-owner/branches"
                  className="mt-4 inline-flex h-9 items-center gap-2 rounded-md bg-blue-600 px-3.5 text-sm font-medium text-white hover:bg-blue-700"
                >
                  <Building2 className="h-4 w-4" />
                  Add branch
                </Link>
              </div>
            ) : (
              branches.slice(0, 5).map((branch) => (
                <div
                  key={branch.id}
                  className="flex items-center justify-between gap-4 rounded-lg border border-slate-100 px-4 py-3"
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
                      <Building2 className="h-4 w-4" />
                    </div>

                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-slate-900">
                        {branch.name}
                      </p>

                      <p className="truncate text-xs text-slate-500">
                        {branch.address || "No address provided"}
                      </p>
                    </div>
                  </div>

                  <span
                    className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-medium ${
                      branch.status.toLowerCase() === "active"
                        ? "bg-green-50 text-green-700"
                        : "bg-slate-100 text-slate-600"
                    }`}
                  >
                    {branch.status}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* QUICK ACTIONS */}
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-base font-semibold text-slate-950">
            Quick Actions
          </h2>

          <p className="mt-1 text-sm text-slate-500">
            Common management tasks.
          </p>

          <div className="mt-6 space-y-2">
            <Link
              href="/dashboard/cyber-owner/branches"
              className="flex items-center gap-3 rounded-lg border border-slate-200 p-3 transition hover:border-blue-200 hover:bg-blue-50/50"
            >
              <Building2 className="h-4 w-4 text-blue-600" />

              <div>
                <p className="text-sm font-medium text-slate-900">
                  Manage branches
                </p>

                <p className="text-xs text-slate-500">
                  Add or update locations
                </p>
              </div>
            </Link>

            <Link
              href="/dashboard/cyber-owner/attendants"
              className="flex items-center gap-3 rounded-lg border border-slate-200 p-3 transition hover:border-blue-200 hover:bg-blue-50/50"
            >
              <UserRound className="h-4 w-4 text-blue-600" />

              <div>
                <p className="text-sm font-medium text-slate-900">
                  Manage attendants
                </p>

                <p className="text-xs text-slate-500">
                  Add and assign attendants
                </p>
              </div>
            </Link>

            <Link
              href="/dashboard/cyber-owner/licence-keys"
              className="flex items-center gap-3 rounded-lg border border-slate-200 p-3 transition hover:border-blue-200 hover:bg-blue-50/50"
            >
              <KeyRound className="h-4 w-4 text-blue-600" />

              <div>
                <p className="text-sm font-medium text-slate-900">
                  Licence keys
                </p>

                <p className="text-xs text-slate-500">
                  Register terminal computers
                </p>
              </div>
            </Link>

            <Link
              href="/dashboard/cyber-owner/subscription"
              className="flex items-center gap-3 rounded-lg border border-slate-200 p-3 transition hover:border-blue-200 hover:bg-blue-50/50"
            >
              <CreditCard className="h-4 w-4 text-blue-600" />

              <div>
                <p className="text-sm font-medium text-slate-900">
                  Subscription
                </p>

                <p className="text-xs text-slate-500">
                  Manage your subscription
                </p>
              </div>
            </Link>
          </div>
        </div>
      </div>

      {/* MANAGEMENT OVERVIEW */}
      <div className="mt-6 grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Link
          href="/dashboard/cyber-owner/terminals"
          className="group rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-blue-200 hover:shadow-md"
        >
          <Laptop className="h-5 w-5 text-blue-600" />

          <h3 className="mt-4 text-sm font-semibold text-slate-950">
            Terminals
          </h3>

          <p className="mt-1 text-sm text-slate-500">
            Manage registered computers.
          </p>

          <div className="mt-4 flex items-center gap-1 text-sm font-medium text-blue-600">
            Manage
            <ArrowRight className="h-4 w-4 transition group-hover:translate-x-1" />
          </div>
        </Link>

        <Link
          href="/dashboard/cyber-owner/customers"
          className="group rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-blue-200 hover:shadow-md"
        >
          <Users className="h-5 w-5 text-blue-600" />

          <h3 className="mt-4 text-sm font-semibold text-slate-950">
            Customers
          </h3>

          <p className="mt-1 text-sm text-slate-500">
            View and manage registered customers.
          </p>

          <div className="mt-4 flex items-center gap-1 text-sm font-medium text-blue-600">
            Manage
            <ArrowRight className="h-4 w-4 transition group-hover:translate-x-1" />
          </div>
        </Link>

        <Link
          href="/dashboard/cyber-owner/receipts"
          className="group rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-blue-200 hover:shadow-md"
        >
          <Receipt className="h-5 w-5 text-blue-600" />

          <h3 className="mt-4 text-sm font-semibold text-slate-950">
            Receipts
          </h3>

          <p className="mt-1 text-sm text-slate-500">
            View issued business receipts.
          </p>

          <div className="mt-4 flex items-center gap-1 text-sm font-medium text-blue-600">
            View receipts
            <ArrowRight className="h-4 w-4 transition group-hover:translate-x-1" />
          </div>
        </Link>

        <Link
          href="/dashboard/cyber-owner/reports"
          className="group rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-blue-200 hover:shadow-md"
        >
          <Activity className="h-5 w-5 text-blue-600" />

          <h3 className="mt-4 text-sm font-semibold text-slate-950">
            Reports
          </h3>

          <p className="mt-1 text-sm text-slate-500">
            Review business performance and records.
          </p>

          <div className="mt-4 flex items-center gap-1 text-sm font-medium text-blue-600">
            View reports
            <ArrowRight className="h-4 w-4 transition group-hover:translate-x-1" />
          </div>
        </Link>
      </div>
    </DashboardShell>
  );
}