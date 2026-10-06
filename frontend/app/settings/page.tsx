"use client";

import { Check, Edit3, Loader2, Plus, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import { subscriptionService } from "@/services/subscription.service";
import { useAuthStore } from "@/store/auth.store";
import { SubscriptionPlan } from "@/types/subscription";

type FormState = {
  name: string;
  monthlyPrice: string;
  includedBranches: string;
  includedTerminals: string;
  extraBranchRate: string;
  extraTerminalRate: string;
  isLifetime: boolean;
  isActive: boolean;
};

const emptyForm: FormState = {
  name: "",
  monthlyPrice: "",
  includedBranches: "",
  includedTerminals: "",
  extraBranchRate: "",
  extraTerminalRate: "",
  isLifetime: false,
  isActive: true,
};

function formatMoney(value: string) {
  const amount = Number(value);

  if (!Number.isFinite(amount)) {
    return `KSh ${value}`;
  }

  return `KSh ${amount.toLocaleString("en-KE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

function formatLimit(value: number) {
  return value === 0 ? "0" : value.toString();
}

function formatRate(value: string) {
  return formatMoney(value);
}

export default function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const token = useAuthStore((state) => state.token);

  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const isSaaSOwner = user?.role === "platform_admin";

  const loadPlans = useCallback(async () => {
    if (!token) {
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const data =
        await subscriptionService.getPlatformPlans(token);

      setPlans(data);
    } catch (err) {
      console.error(
        "Failed to load subscription plans:",
        err
      );

      setError(
        err instanceof Error
          ? err.message
          : "Failed to load subscription plans."
      );
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (isSaaSOwner) {
      void loadPlans();
    } else {
      setLoading(false);
    }
  }, [isSaaSOwner, loadPlans]);

  if (!user) {
    return null;
  }

  function resetForm() {
    setEditingId(null);
    setForm(emptyForm);
    setError(null);
    setSuccess(null);
  }

  function startAdd() {
    setEditingId("new");
    setForm(emptyForm);
    setError(null);
    setSuccess(null);

    window.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  }

  function startEdit(plan: SubscriptionPlan) {
    setEditingId(plan.id);

    setForm({
      name: plan.name,
      monthlyPrice: plan.monthly_price,
      includedBranches:
        plan.included_branches.toString(),
      includedTerminals:
        plan.included_terminals.toString(),
      extraBranchRate: plan.extra_branch_rate,
      extraTerminalRate: plan.extra_terminal_rate,
      isLifetime: plan.is_lifetime,
      isActive: plan.is_active,
    });

    setError(null);
    setSuccess(null);

    window.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  }

  function updateField<K extends keyof FormState>(
    field: K,
    value: FormState[K]
  ) {
    setForm((current) => ({
      ...current,
      [field]: value,
    }));
  }

  async function handleSave() {
    if (!editingId || !token) {
      return;
    }

    setError(null);
    setSuccess(null);

    const name = form.name.trim();
    const monthlyPrice = form.monthlyPrice.trim();
    const extraBranchRate =
      form.extraBranchRate.trim();
    const extraTerminalRate =
      form.extraTerminalRate.trim();

    if (!name) {
      setError("Package name is required.");
      return;
    }

    if (!monthlyPrice) {
      setError("Monthly price is required.");
      return;
    }

    if (!extraBranchRate) {
      setError("Extra branch rate is required.");
      return;
    }

    if (!extraTerminalRate) {
      setError("Extra terminal rate is required.");
      return;
    }

    const monthlyPriceNumber = Number(monthlyPrice);
    const extraBranchRateNumber =
      Number(extraBranchRate);
    const extraTerminalRateNumber =
      Number(extraTerminalRate);

    const includedBranches = Number(
      form.includedBranches
    );
    const includedTerminals = Number(
      form.includedTerminals
    );

    if (
      !Number.isFinite(monthlyPriceNumber) ||
      monthlyPriceNumber < 0
    ) {
      setError(
        "Monthly price must be a valid non-negative amount."
      );
      return;
    }

    if (
      !Number.isFinite(extraBranchRateNumber) ||
      extraBranchRateNumber < 0
    ) {
      setError(
        "Extra branch rate must be a valid non-negative amount."
      );
      return;
    }

    if (
      !Number.isFinite(extraTerminalRateNumber) ||
      extraTerminalRateNumber < 0
    ) {
      setError(
        "Extra terminal rate must be a valid non-negative amount."
      );
      return;
    }

    if (
      !Number.isInteger(includedBranches) ||
      includedBranches < 0
    ) {
      setError(
        "Included branches must be a non-negative whole number."
      );
      return;
    }

    if (
      !Number.isInteger(includedTerminals) ||
      includedTerminals < 0
    ) {
      setError(
        "Included terminals must be a non-negative whole number."
      );
      return;
    }

    if (
      !form.isActive &&
      !window.confirm(
        "Are you sure you want to deactivate this subscription plan?\n\nCyber Owners will no longer be able to select it for new subscriptions or payments."
      )
    ) {
      return;
    }

    setSaving(true);

    const payload = {
      name,
      included_branches: includedBranches,
      included_terminals: includedTerminals,
      extra_branch_rate: extraBranchRate,
      extra_terminal_rate: extraTerminalRate,
      monthly_price: monthlyPrice,
      is_lifetime: form.isLifetime,
      is_active: form.isActive,
    };

    try {
      if (editingId === "new") {
        const created =
          await subscriptionService.createPlatformPlan(
            payload,
            token
          );

        setPlans((current) => [
          ...current,
          created,
        ]);

        setSuccess(
          `${created.name} package was created successfully.`
        );
      } else {
        const updated =
          await subscriptionService.updatePlatformPlan(
            editingId,
            payload,
            token
          );

        setPlans((current) =>
          current.map((plan) =>
            plan.id === updated.id
              ? updated
              : plan
          )
        );

        setSuccess(
          `${updated.name} package was updated successfully.`
        );
      }

      setEditingId(null);
      setForm(emptyForm);
    } catch (err) {
      console.error(
        editingId === "new"
          ? "Failed to create subscription plan:"
          : "Failed to update subscription plan:",
        err
      );

      setError(
        err instanceof Error
          ? err.message
          : editingId === "new"
            ? "Failed to create subscription plan."
            : "Failed to update subscription plan."
      );
    } finally {
      setSaving(false);
    }
  }

  async function handleToggleActive(
    plan: SubscriptionPlan
  ) {
    if (!token || saving) {
      return;
    }

    const nextActiveState = !plan.is_active;

    if (
      !nextActiveState &&
      !window.confirm(
        `Deactivate "${plan.name}"?\n\nCyber Owners will no longer be able to select this plan for new subscriptions or payments.`
      )
    ) {
      return;
    }

    setError(null);
    setSuccess(null);
    setSaving(true);

    try {
      const updated =
        await subscriptionService.updatePlatformPlan(
          plan.id,
          {
            name: plan.name,
            included_branches:
              plan.included_branches,
            included_terminals:
              plan.included_terminals,
            extra_branch_rate:
              plan.extra_branch_rate,
            extra_terminal_rate:
              plan.extra_terminal_rate,
            monthly_price:
              plan.monthly_price,
            is_lifetime:
              plan.is_lifetime,
            is_active:
              nextActiveState,
          },
          token
        );

      setPlans((current) =>
        current.map((item) =>
          item.id === updated.id
            ? updated
            : item
        )
      );

      setSuccess(
        `${updated.name} is now ${
          updated.is_active
            ? "active"
            : "inactive"
        }.`
      );
    } catch (err) {
      console.error(
        "Failed to change subscription plan status:",
        err
      );

      setError(
        err instanceof Error
          ? err.message
          : "Failed to update subscription plan."
      );
    } finally {
      setSaving(false);
    }
  }

  if (!isSaaSOwner) {
    return (
      <DashboardShell role={user.role}>
        <PageHeader
          title="Settings"
          description="Manage your CyberSaaS application settings."
        />

        <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-bold text-slate-950">
            Account Settings
          </h2>

          <p className="mt-2 text-sm text-slate-500">
            Settings available for your account
            will appear here.
          </p>
        </section>
      </DashboardShell>
    );
  }

  return (
    <DashboardShell role={user.role}>
      <PageHeader
        title="Settings"
        description="Manage your CyberSaaS subscription packages."
      />

      <div className="space-y-6">
        {/* STATUS MESSAGES */}

        {error && (
          <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {success && (
          <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
            {success}
          </div>
        )}

        {/* HEADER */}

        <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-lg font-bold text-slate-950">
                Subscription Packages
              </h2>

              <p className="mt-1 max-w-3xl text-sm leading-6 text-slate-500">
                These packages are loaded directly
                from your subscription database.
                Changes made here are saved to
                PostgreSQL and will be reflected
                wherever active subscription packages
                are displayed.
              </p>
            </div>

            <button
              type="button"
              onClick={startAdd}
              disabled={saving || editingId === "new"}
              className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg bg-[#0757B8] px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-[#064a9d] disabled:cursor-not-allowed disabled:opacity-60"
            >
              <Plus className="h-4 w-4" />
              Add Package
            </button>
          </div>
        </section>

        {/* ADD / EDIT FORM */}

        {editingId && (
          <section className="rounded-2xl border border-blue-200 bg-white p-6 shadow-sm">
            <div className="mb-6 flex items-start justify-between gap-4">
              <div>
                <h3 className="text-base font-bold text-slate-950">
                  {editingId === "new"
                    ? "Add Subscription Package"
                    : "Update Subscription Package"}
                </h3>

                <p className="mt-1 text-sm text-slate-500">
                  {editingId === "new"
                    ? "Create a new subscription package for Cyber Owners."
                    : "Changes will be saved to the actual subscription plan in PostgreSQL."}
                </p>
              </div>

              <button
                type="button"
                onClick={resetForm}
                disabled={saving}
                aria-label="Close package form"
                className="rounded-lg p-2 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="grid gap-5 md:grid-cols-2">
              {/* Package Name */}

              <div>
                <label
                  htmlFor="package-name"
                  className="mb-2 block text-sm font-medium text-slate-700"
                >
                  Package Name
                </label>

                <input
                  id="package-name"
                  value={form.name}
                  onChange={(event) =>
                    updateField(
                      "name",
                      event.target.value
                    )
                  }
                  disabled={saving}
                  placeholder="e.g. Enterprise"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />
              </div>

              {/* Monthly Price */}

              <div>
                <label
                  htmlFor="monthly-price"
                  className="mb-2 block text-sm font-medium text-slate-700"
                >
                  Monthly Price (KSh)
                </label>

                <input
                  id="monthly-price"
                  type="number"
                  min="0"
                  step="0.01"
                  value={form.monthlyPrice}
                  onChange={(event) =>
                    updateField(
                      "monthlyPrice",
                      event.target.value
                    )
                  }
                  disabled={saving}
                  placeholder="10000.00"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />

                <p className="mt-1 text-xs text-slate-400">
                  Used as the subscription price for
                  this plan.
                </p>
              </div>

              {/* Included Branches */}

              <div>
                <label
                  htmlFor="included-branches"
                  className="mb-2 block text-sm font-medium text-slate-700"
                >
                  Included Branches
                </label>

                <input
                  id="included-branches"
                  type="number"
                  min="0"
                  step="1"
                  value={form.includedBranches}
                  onChange={(event) =>
                    updateField(
                      "includedBranches",
                      event.target.value
                    )
                  }
                  disabled={saving}
                  placeholder="25"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />
              </div>

              {/* Included Terminals */}

              <div>
                <label
                  htmlFor="included-terminals"
                  className="mb-2 block text-sm font-medium text-slate-700"
                >
                  Included Terminals
                </label>

                <input
                  id="included-terminals"
                  type="number"
                  min="0"
                  step="1"
                  value={form.includedTerminals}
                  onChange={(event) =>
                    updateField(
                      "includedTerminals",
                      event.target.value
                    )
                  }
                  disabled={saving}
                  placeholder="150"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />
              </div>

              {/* Extra Branch Rate */}

              <div>
                <label
                  htmlFor="extra-branch-rate"
                  className="mb-2 block text-sm font-medium text-slate-700"
                >
                  Extra Branch Rate (KSh)
                </label>

                <input
                  id="extra-branch-rate"
                  type="number"
                  min="0"
                  step="0.01"
                  value={form.extraBranchRate}
                  onChange={(event) =>
                    updateField(
                      "extraBranchRate",
                      event.target.value
                    )
                  }
                  disabled={saving}
                  placeholder="1000.00"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />
              </div>

              {/* Extra Terminal Rate */}

              <div>
                <label
                  htmlFor="extra-terminal-rate"
                  className="mb-2 block text-sm font-medium text-slate-700"
                >
                  Extra Terminal Rate (KSh)
                </label>

                <input
                  id="extra-terminal-rate"
                  type="number"
                  min="0"
                  step="0.01"
                  value={form.extraTerminalRate}
                  onChange={(event) =>
                    updateField(
                      "extraTerminalRate",
                      event.target.value
                    )
                  }
                  disabled={saving}
                  placeholder="200.00"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-blue-100 disabled:bg-slate-50"
                />
              </div>
            </div>

            {/* Lifetime */}

            <div className="mt-5 rounded-xl border border-slate-200 bg-slate-50 p-4">
              <label className="flex cursor-pointer items-start gap-3">
                <input
                  type="checkbox"
                  checked={form.isLifetime}
                  onChange={(event) =>
                    updateField(
                      "isLifetime",
                      event.target.checked
                    )
                  }
                  disabled={saving}
                  className="mt-0.5 h-4 w-4 rounded border-slate-300 text-[#0757B8] focus:ring-blue-200"
                />

                <span>
                  <span className="block text-sm font-semibold text-slate-800">
                    Lifetime subscription
                  </span>

                  <span className="mt-1 block text-xs leading-5 text-slate-500">
                    This plan represents a lifetime
                    subscription. The backend will
                    handle it using
                    <code className="mx-1 rounded bg-white px-1 py-0.5">
                      is_lifetime
                    </code>
                    rather than a separate billing
                    type.
                  </span>
                </span>
              </label>
            </div>

            {/* Active */}

            <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
              <label className="flex cursor-pointer items-start gap-3">
                <input
                  type="checkbox"
                  checked={form.isActive}
                  onChange={(event) =>
                    updateField(
                      "isActive",
                      event.target.checked
                    )
                  }
                  disabled={saving}
                  className="mt-0.5 h-4 w-4 rounded border-slate-300 text-[#0757B8] focus:ring-blue-200"
                />

                <span>
                  <span className="block text-sm font-semibold text-slate-800">
                    Package is active
                  </span>

                  <span className="mt-1 block text-xs leading-5 text-slate-500">
                    Inactive packages remain in the
                    database for historical purposes
                    but cannot be selected for new
                    subscriptions.
                  </span>
                </span>
              </label>
            </div>

            {/* Actions */}

            <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                type="button"
                onClick={resetForm}
                disabled={saving}
                className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                Cancel
              </button>

              <button
                type="button"
                onClick={() => void handleSave()}
                disabled={saving}
                className="inline-flex items-center justify-center gap-2 rounded-lg bg-[#0757B8] px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-[#064a9d] disabled:cursor-not-allowed disabled:opacity-60"
              >
                {saving ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    {editingId === "new"
                      ? "Creating..."
                      : "Saving..."}
                  </>
                ) : (
                  <>
                    {editingId === "new" ? (
                      <Plus className="h-4 w-4" />
                    ) : (
                      <Check className="h-4 w-4" />
                    )}

                    {editingId === "new"
                      ? "Create Package"
                      : "Save Changes"}
                  </>
                )}
              </button>
            </div>
          </section>
        )}

        {/* PLANS */}

        <section>
          <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-lg font-bold text-slate-950">
                Subscription Packages
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                These are the actual packages currently
                stored in the CyberSaaS subscription
                system.
              </p>
            </div>

            {!editingId && (
              <button
                type="button"
                onClick={startAdd}
                disabled={saving}
                className="inline-flex items-center justify-center gap-2 rounded-lg border border-[#0757B8] px-4 py-2.5 text-sm font-semibold text-[#0757B8] transition hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <Plus className="h-4 w-4" />
                Add Package
              </button>
            )}
          </div>

          {loading ? (
            <div className="flex min-h-48 items-center justify-center rounded-2xl border border-slate-200 bg-white shadow-sm">
              <div className="flex items-center gap-3 text-sm text-slate-500">
                <Loader2 className="h-5 w-5 animate-spin" />
                Loading subscription packages...
              </div>
            </div>
          ) : plans.length === 0 ? (
            <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
              <h3 className="font-semibold text-slate-900">
                No subscription packages found
              </h3>

              <p className="mt-2 text-sm text-slate-500">
                There are currently no subscription
                plans in the database.
              </p>

              <button
                type="button"
                onClick={() => void loadPlans()}
                className="mt-4 rounded-lg bg-[#0757B8] px-4 py-2.5 text-sm font-semibold text-white hover:bg-[#064a9d]"
              >
                Reload
              </button>
            </div>
          ) : (
            <div className="grid gap-5 lg:grid-cols-2">
              {plans.map((plan) => (
                <article
                  key={plan.id}
                  className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm"
                >
                  {/* Header */}

                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-3">
                        <h3 className="text-lg font-bold text-slate-950">
                          {plan.name}
                        </h3>

                        <span
                          className={`rounded-full px-2.5 py-1 text-xs font-semibold ${
                            plan.is_active
                              ? "bg-emerald-50 text-emerald-700"
                              : "bg-slate-100 text-slate-500"
                          }`}
                        >
                          {plan.is_active
                            ? "Active"
                            : "Inactive"}
                        </span>

                        {plan.is_lifetime && (
                          <span className="rounded-full bg-blue-50 px-2.5 py-1 text-xs font-semibold text-blue-700">
                            Lifetime
                          </span>
                        )}
                      </div>

                      <p className="mt-4 text-2xl font-bold text-[#0757B8]">
                        {formatMoney(
                          plan.monthly_price
                        )}
                      </p>

                      <p className="mt-1 text-xs text-slate-500">
                        {plan.is_lifetime
                          ? "Lifetime subscription"
                          : "Monthly subscription"}
                      </p>
                    </div>

                    <button
                      type="button"
                      onClick={() =>
                        startEdit(plan)
                      }
                      disabled={saving}
                      className="inline-flex shrink-0 items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      <Edit3 className="h-4 w-4" />
                      Edit
                    </button>
                  </div>

                  {/* Limits */}

                  <div className="mt-6 grid grid-cols-2 gap-3">
                    <div className="rounded-xl bg-slate-50 p-4">
                      <p className="text-xs text-slate-500">
                        Included Branches
                      </p>

                      <p className="mt-1 text-sm font-bold text-slate-900">
                        {formatLimit(
                          plan.included_branches
                        )}
                      </p>
                    </div>

                    <div className="rounded-xl bg-slate-50 p-4">
                      <p className="text-xs text-slate-500">
                        Included Terminals
                      </p>

                      <p className="mt-1 text-sm font-bold text-slate-900">
                        {formatLimit(
                          plan.included_terminals
                        )}
                      </p>
                    </div>
                  </div>

                  {/* Extra rates */}

                  <div className="mt-3 grid grid-cols-2 gap-3">
                    <div className="rounded-xl bg-slate-50 p-4">
                      <p className="text-xs text-slate-500">
                        Extra Branch
                      </p>

                      <p className="mt-1 text-sm font-bold text-slate-900">
                        {formatRate(
                          plan.extra_branch_rate
                        )}
                      </p>
                    </div>

                    <div className="rounded-xl bg-slate-50 p-4">
                      <p className="text-xs text-slate-500">
                        Extra Terminal
                      </p>

                      <p className="mt-1 text-sm font-bold text-slate-900">
                        {formatRate(
                          plan.extra_terminal_rate
                        )}
                      </p>
                    </div>
                  </div>

                  {/* Footer */}

                  <div className="mt-5 flex flex-col gap-3 border-t border-slate-100 pt-4 sm:flex-row sm:items-center sm:justify-between">
                    <span className="text-xs text-slate-400">
                      Database plan ID: {plan.id}
                    </span>

                    <button
                      type="button"
                      onClick={() =>
                        void handleToggleActive(
                          plan
                        )
                      }
                      disabled={saving}
                      className={`text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-50 ${
                        plan.is_active
                          ? "text-red-600 hover:text-red-700"
                          : "text-emerald-600 hover:text-emerald-700"
                      }`}
                    >
                      {plan.is_active
                        ? "Deactivate"
                        : "Activate"}
                    </button>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </div>
    </DashboardShell>
  );
}