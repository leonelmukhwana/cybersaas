"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  CalendarDays,
  CheckCircle2,
  CreditCard,
  Package,
  RefreshCw,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import { useAuthStore } from "@/store/auth.store";
import { subscriptionService } from "@/services/subscription.service";

import type { SubscriptionOverview } from "@/types/subscription";

export default function SubscriptionPage() {
  const router = useRouter();

  const token = useAuthStore((state) => state.token);

  const [overview, setOverview] =
    useState<SubscriptionOverview | null>(null);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token) {
      router.replace("/login/owner");
      return;
    }

    const currentToken = token;

    async function loadSubscription() {
      try {
        setLoading(true);
        setError("");

        const response =
          await subscriptionService.getOverview(currentToken);

        setOverview(response);
      } catch (err) {
        console.error("LOAD SUBSCRIPTION ERROR:", err);

        setError(
          err instanceof Error
            ? err.message
            : "Unable to load your subscription."
        );
      } finally {
        setLoading(false);
      }
    }

    loadSubscription();
  }, [token, router]);

  function formatDate(value?: string | null) {
    if (!value) {
      return "—";
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
      return "—";
    }

    return new Intl.DateTimeFormat("en-KE", {
      day: "numeric",
      month: "long",
      year: "numeric",
    }).format(date);
  }

  function formatAmount(amount?: string | null) {
    if (!amount) {
      return "KES 0";
    }

    const value = Number(amount);

    if (Number.isNaN(value)) {
      return `KES ${amount}`;
    }

    return new Intl.NumberFormat("en-KE", {
      style: "currency",
      currency: "KES",
      maximumFractionDigits: 0,
    }).format(value);
  }

  function getStatusLabel() {
    if (!overview) {
      return "Unknown";
    }

    if (overview.is_lifetime) {
      return "Lifetime";
    }

    if (overview.is_expired) {
      return "Expired";
    }

    if (overview.is_trial) {
      return overview.is_trial_expired
        ? "Trial Expired"
        : "Trial";
    }

    if (overview.is_active) {
      return "Active";
    }

    return overview.subscription?.status || "Inactive";
  }

  function getStatusClass() {
    if (!overview) {
      return "bg-slate-100 text-slate-700";
    }

    if (overview.is_lifetime) {
      return "bg-blue-100 text-blue-700";
    }

    if (overview.is_expired || overview.is_trial_expired) {
      return "bg-red-100 text-red-700";
    }

    if (overview.is_active) {
      return "bg-green-100 text-green-700";
    }

    return "bg-amber-100 text-amber-700";
  }

  function handlePayment() {
    if (!overview?.subscription?.plan_id) {
      return;
    }

    router.push(
      `/cyber-owner/subscription/payment?plan=${encodeURIComponent(
        overview.subscription.plan_id
      )}`
    );
  }

  async function handleRefresh() {
    if (!token) {
      return;
    }

    try {
      setLoading(true);
      setError("");

      const response =
        await subscriptionService.getOverview(token);

      setOverview(response);
    } catch (err) {
      console.error("REFRESH SUBSCRIPTION ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Unable to refresh your subscription."
      );
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return (
      <DashboardShell role="owner">
        <PageHeader
          title="Subscription"
          description="View and manage your current CyberSaaS subscription."
        />

        <div className="flex min-h-[40vh] items-center justify-center">
          <p className="text-sm text-slate-500">
            Loading your subscription...
          </p>
        </div>
      </DashboardShell>
    );
  }

  if (error) {
    return (
      <DashboardShell role="owner">
        <PageHeader
          title="Subscription"
          description="View and manage your current CyberSaaS subscription."
        />

        <div className="rounded-xl border border-red-200 bg-red-50 p-6">
          <h2 className="font-semibold text-red-700">
            Unable to load subscription
          </h2>

          <p className="mt-2 text-sm text-red-600">
            {error}
          </p>

          <Button
            onClick={handleRefresh}
            className="mt-4 bg-[#0757B8] hover:bg-[#064A9D]"
          >
            Try Again
          </Button>
        </div>
      </DashboardShell>
    );
  }

  if (!overview || !overview.subscription) {
    return (
      <DashboardShell role="owner">
        <PageHeader
          title="Subscription"
          description="View and manage your current CyberSaaS subscription."
        />

        <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
          <Package className="mx-auto h-10 w-10 text-slate-400" />

          <h2 className="mt-4 text-lg font-semibold text-slate-950">
            No subscription selected
          </h2>

          <p className="mx-auto mt-2 max-w-md text-sm text-slate-500">
            Choose a CyberSaaS package to activate your subscription.
          </p>

          <Button
            onClick={() => router.push("/subscription/plans")}
            className="mt-6 bg-[#0757B8] hover:bg-[#064A9D]"
          >
            Choose a Package
          </Button>
        </div>
      </DashboardShell>
    );
  }

  const subscription = overview.subscription;
  const plan = overview.plan;

  return (
    <DashboardShell role="owner">
      <PageHeader
        title="Subscription"
        description="View your current package, subscription status, and payment information."
      />

      <div className="space-y-6">
        {/* CURRENT PACKAGE */}
        <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-col gap-5 md:flex-row md:items-start md:justify-between">
            <div>
              <div className="flex items-center gap-3">
                <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-50">
                  <Package className="h-5 w-5 text-[#0757B8]" />
                </div>

                <div>
                  <p className="text-sm text-slate-500">
                    Current Package
                  </p>

                  <h2 className="text-2xl font-bold text-slate-950">
                    {plan?.name ||
                      subscription.plan_name ||
                      "Subscription Package"}
                  </h2>
                </div>
              </div>

              {plan && (
                <div className="mt-5">
                  <p className="text-3xl font-bold text-[#0757B8]">
                    {formatAmount(plan.monthly_price)}
                  </p>

                  <p className="mt-1 text-sm text-slate-500">
                    {plan.is_lifetime
                      ? "One-time payment"
                      : "Subscription price"}
                  </p>
                </div>
              )}
            </div>

            <div
              className={`inline-flex w-fit items-center gap-2 rounded-full px-4 py-2 text-sm font-semibold ${getStatusClass()}`}
            >
              <CheckCircle2 className="h-4 w-4" />
              {getStatusLabel()}
            </div>
          </div>
        </div>

        {/* SUBSCRIPTION DETAILS */}
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <CalendarDays className="h-5 w-5 text-[#0757B8]" />

              <p className="text-sm text-slate-500">
                Subscribed On
              </p>
            </div>

            <p className="mt-3 font-semibold text-slate-950">
              {formatDate(subscription.created_at)}
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <CalendarDays className="h-5 w-5 text-[#0757B8]" />

              <p className="text-sm text-slate-500">
                Current Period
              </p>
            </div>

            <p className="mt-3 font-semibold text-slate-950">
              {formatDate(subscription.current_period_start)}
            </p>

            {!subscription.is_lifetime &&
              subscription.current_period_end && (
                <p className="mt-1 text-xs text-slate-500">
                  Ends {formatDate(subscription.current_period_end)}
                </p>
              )}
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <Package className="h-5 w-5 text-[#0757B8]" />

              <p className="text-sm text-slate-500">
                Branches
              </p>
            </div>

            <p className="mt-3 text-xl font-bold text-slate-950">
              {overview.branches_used} / {overview.branches_limit}
            </p>

            <p className="mt-1 text-xs text-slate-500">
              Branches used
            </p>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <Package className="h-5 w-5 text-[#0757B8]" />

              <p className="text-sm text-slate-500">
                Account Balance
              </p>
            </div>

            <p className="mt-3 text-xl font-bold text-slate-950">
              {formatAmount(subscription.account_balance)}
            </p>
          </div>
        </div>

        {/* EXPIRY / PAYMENT */}
        <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 className="text-lg font-bold text-slate-950">
                Subscription Payment
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Keep your CyberSaaS subscription active by making your payment.
              </p>

              <div className="mt-4">
                {subscription.is_lifetime ? (
                  <p className="text-sm font-medium text-blue-700">
                    Your package does not expire.
                  </p>
                ) : subscription.current_period_end ? (
                  <p className="text-sm">
                    <span className="text-slate-500">
                      Expires on:
                    </span>{" "}
                    <span className="font-semibold text-slate-950">
                      {formatDate(subscription.current_period_end)}
                    </span>
                  </p>
                ) : (
                  <p className="text-sm text-slate-500">
                    Expiry date is not available.
                  </p>
                )}
              </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row">
              <Button
                type="button"
                variant="outline"
                onClick={handleRefresh}
                className="h-11 border-slate-300"
              >
                <RefreshCw className="mr-2 h-4 w-4" />
                Refresh
              </Button>

              {!subscription.is_lifetime && (
                <Button
                  type="button"
                  onClick={handlePayment}
                  className="h-11 bg-[#0757B8] px-6 font-semibold hover:bg-[#064A9D]"
                >
                  <CreditCard className="mr-2 h-4 w-4" />
                  Pay for Subscription
                </Button>
              )}
            </div>
          </div>
        </div>

        {/* PACKAGE LIMITS */}
        {plan && (
          <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-lg font-bold text-slate-950">
              Package Limits
            </h2>

            <div className="mt-5 grid gap-4 md:grid-cols-2">
              <div className="rounded-xl bg-slate-50 p-4">
                <p className="text-sm text-slate-500">
                  Included Branches
                </p>

                <p className="mt-1 text-xl font-bold text-slate-950">
                  {plan.included_branches}
                </p>
              </div>

              <div className="rounded-xl bg-slate-50 p-4">
                <p className="text-sm text-slate-500">
                  Included Terminals
                </p>

                <p className="mt-1 text-xl font-bold text-slate-950">
                  {plan.included_terminals}
                </p>
              </div>
            </div>
          </div>
        )}
      </div>
    </DashboardShell>
  );
}