"use client";

import { useEffect, useMemo, useState } from "react";

import {
  Plus,
  RefreshCw,
  Save,
  X,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import { useAuthStore } from "@/store/auth.store";
import { useExpenseStore } from "@/store/expense.store";
import { attendantService } from "@/services/attendant.service";

import type {
  CreateExpenseRequest,
  ExpensePaymentMethod,
} from "@/types/expense";

const CATEGORY_OPTIONS = [
  "KPLC Token",
  "Printing Paper",
  "Envelopes",
  "Stationery",
  "Internet",
  "Cleaning",
  "Transport",
  "Repairs",
  "Other",
];

const PAYMENT_METHODS: ExpensePaymentMethod[] = [
  "cash",
  "mpesa",
  "bank",
  "card",
  "other",
];

function todayDate() {
  const date = new Date();

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

function formatAmount(amount: string, currency: string) {
  const value = Number(amount);

  if (Number.isNaN(value)) {
    return `${currency} ${amount}`;
  }

  return `${currency} ${value.toLocaleString("en-KE", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

function formatDate(value: string) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleDateString("en-KE", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export default function ExpensesPage() {
  const { token, user, isOffline } = useAuthStore();

  const {
    expenses,
    loading,
    saving,
    error,
    fetchExpenses,
    createExpense,
    clearError,
  } = useExpenseStore();

  const [branchId, setBranchId] = useState<string | null>(null);
  const [branchName, setBranchName] = useState<string | null>(null);
  const [branchLoading, setBranchLoading] = useState(true);
  const [branchError, setBranchError] = useState<string | null>(null);

  const [showForm, setShowForm] = useState(false);

  const [category, setCategory] = useState("");
  const [amount, setAmount] = useState("");
  const [paymentMethod, setPaymentMethod] =
    useState<ExpensePaymentMethod>("cash");
  const [expenseDate, setExpenseDate] = useState(todayDate());
  const [description, setDescription] = useState("");
  const [reference, setReference] = useState("");

  useEffect(() => {
    let mounted = true;

    async function loadAttendantBranch() {
      if (!token || !user || user.role !== "attendant") {
        if (mounted) {
          setBranchLoading(false);
          setBranchError("Attendant authentication is required.");
        }

        return;
      }

      if (isOffline) {
        if (mounted) {
          setBranchLoading(false);
          setBranchError(
            "Expenses are currently available only while the attendant dashboard is online."
          );
        }

        return;
      }

      try {
        setBranchLoading(true);
        setBranchError(null);

        const attendant = await attendantService.me(token);

        if (!mounted) {
          return;
        }

        if (!attendant.branch_id) {
          setBranchId(null);
          setBranchName(null);
          setBranchError(
            "This attendant has not been assigned to any branch. Ask the Cyber Owner to assign a branch."
          );

          return;
        }

        setBranchId(attendant.branch_id);
        setBranchName(attendant.branch_name ?? null);
      } catch (err) {
        if (!mounted) {
          return;
        }

        setBranchId(null);
        setBranchName(null);

        setBranchError(
          err instanceof Error
            ? err.message
            : "Failed to load the attendant branch."
        );
      } finally {
        if (mounted) {
          setBranchLoading(false);
        }
      }
    }

    loadAttendantBranch();

    return () => {
      mounted = false;
    };
  }, [token, user, isOffline]);

  useEffect(() => {
    if (!branchId || isOffline) {
      return;
    }

    fetchExpenses(branchId);
  }, [branchId, isOffline, fetchExpenses]);

  const today = todayDate();

  /*
   * ATTENDANT VIEW RULE
   *
   * The attendant dashboard displays only today's
   * recorded expenses.
   *
   * Historical expenses remain stored in the database
   * and are still available to the Cyber Owner and
   * compliance reports.
   */
  const todayExpenses = useMemo(
    () =>
      expenses.filter(
        (expense) =>
          expense.expense_date.slice(0, 10) === today &&
          expense.status === "recorded"
      ),
    [expenses, today]
  );

  const todayTotal = useMemo(
    () =>
      todayExpenses.reduce(
        (sum, expense) => sum + Number(expense.amount || 0),
        0
      ),
    [todayExpenses]
  );

  const currency =
    todayExpenses.length > 0
      ? todayExpenses[0].currency || "KES"
      : "KES";

  function resetForm() {
    setCategory("");
    setAmount("");
    setPaymentMethod("cash");
    setExpenseDate(todayDate());
    setDescription("");
    setReference("");
    clearError();
  }

  function closeForm() {
    if (saving) {
      return;
    }

    setShowForm(false);
    resetForm();
  }

  async function handleSubmit(
    event: React.FormEvent<HTMLFormElement>
  ) {
    event.preventDefault();

    clearError();

    if (!branchId) {
      setBranchError(
        "This attendant has not been assigned to any branch."
      );

      return;
    }

    const numericAmount = Number(amount);

    if (!category.trim()) {
      return;
    }

    if (
      !amount ||
      Number.isNaN(numericAmount) ||
      numericAmount <= 0
    ) {
      return;
    }

    /*
     * Attendants can only record expenses for today.
     */
    const currentToday = todayDate();

    if (expenseDate !== currentToday) {
      return;
    }

    const data: CreateExpenseRequest = {
      branch_id: branchId,
      category: category.trim(),
      amount: numericAmount.toFixed(2),
      currency: "KES",
      payment_method: paymentMethod,
      expense_date: currentToday,
      description: description.trim() || null,
      reference: reference.trim() || null,
    };

    try {
      await createExpense(data);

      setShowForm(false);
      resetForm();

      await fetchExpenses(branchId);
    } catch {
      /*
       * The store keeps the API error.
       * Keep the form open so the attendant can correct
       * the data.
       */
    }
  }

  async function handleRefresh() {
    if (!branchId || isOffline || loading) {
      return;
    }

    clearError();

    await fetchExpenses(branchId);
  }

  return (
    <DashboardShell role="attendant">
      <div className="space-y-6">
        <PageHeader
          title="Expenses"
          description={
            branchName
              ? `Record and view today's expenses for ${branchName}.`
              : "Record and view today's expenses for your assigned branch."
          }
        />

        {isOffline && (
          <Card>
            <CardContent className="pt-6">
              <div className="text-sm text-muted-foreground">
                Expenses require an active internet connection. Offline
                attendant sessions continue operating from the terminal.
              </div>
            </CardContent>
          </Card>
        )}

        {branchError && (
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-3">
                <p className="text-sm text-destructive">
                  {branchError}
                </p>

                {!branchLoading &&
                  branchError.includes("has not been assigned") && (
                    <p className="text-sm text-muted-foreground">
                      The Cyber Owner must assign this attendant to a
                      branch before expenses can be recorded.
                    </p>
                  )}
              </div>
            </CardContent>
          </Card>
        )}

        {error && (
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="font-medium text-destructive">
                    Unable to save expense
                  </p>

                  <p className="mt-1 text-sm text-destructive">
                    {error}
                  </p>
                </div>

                <Button
                  type="button"
                  variant="outline"
                  className="h-10 shrink-0 gap-2 px-4"
                  onClick={clearError}
                >
                  <X className="h-4 w-4" />
                  Dismiss
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">
                Today&apos;s Expenses
              </p>

              <p className="mt-2 text-2xl font-semibold">
                {currency}{" "}
                {todayTotal.toLocaleString("en-KE", {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">
                Today&apos;s Records
              </p>

              <p className="mt-2 text-2xl font-semibold">
                {todayExpenses.length}
              </p>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardContent className="flex flex-col gap-4 pt-6 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-lg font-semibold">
                Today&apos;s Branch Expenses
              </h2>

              <p className="text-sm text-muted-foreground">
                Only expenses recorded today are shown here.
              </p>
            </div>

            <div className="flex gap-3">
              <Button
                type="button"
                onClick={handleRefresh}
                disabled={
                  branchLoading ||
                  !branchId ||
                  isOffline ||
                  loading
                }
                className="h-10 gap-2 rounded-md bg-blue-600 px-4 font-medium text-white shadow-sm hover:bg-blue-700"
              >
                <RefreshCw
                  className={`h-4 w-4 ${
                    loading ? "animate-spin" : ""
                  }`}
                />
                Refresh
              </Button>

              <Button
                type="button"
                onClick={() => {
                  clearError();
                  setExpenseDate(todayDate());
                  setShowForm(true);
                }}
                disabled={
                  branchLoading ||
                  !branchId ||
                  isOffline
                }
                className="h-10 gap-2 rounded-md bg-blue-600 px-4 font-medium text-white shadow-sm hover:bg-blue-700"
              >
                <Plus className="h-4 w-4" />
                Add Expense
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            {loading || branchLoading ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                Loading today&apos;s expenses...
              </div>
            ) : todayExpenses.length === 0 ? (
              <div className="py-10 text-center text-sm text-muted-foreground">
                No expenses have been recorded today.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full min-w-[900px] text-sm">
                  <thead>
                    <tr className="border-b text-left">
                      <th className="px-3 py-3 font-medium">
                        Category
                      </th>

                      <th className="px-3 py-3 font-medium">
                        Description
                      </th>

                      <th className="px-3 py-3 font-medium">
                        Amount
                      </th>

                      <th className="px-3 py-3 font-medium">
                        Payment
                      </th>

                      <th className="px-3 py-3 font-medium">
                        Date
                      </th>

                      <th className="px-3 py-3 font-medium">
                        Reference
                      </th>

                      <th className="px-3 py-3 font-medium">
                        Status
                      </th>
                    </tr>
                  </thead>

                  <tbody>
                    {todayExpenses.map((expense) => (
                      <tr
                        key={expense.id}
                        className="border-b last:border-0"
                      >
                        <td className="px-3 py-3 font-medium">
                          {expense.category}
                        </td>

                        <td className="px-3 py-3 text-muted-foreground">
                          {expense.description || "-"}
                        </td>

                        <td className="px-3 py-3">
                          {formatAmount(
                            expense.amount,
                            expense.currency
                          )}
                        </td>

                        <td className="px-3 py-3 capitalize">
                          {expense.payment_method}
                        </td>

                        <td className="px-3 py-3">
                          {formatDate(expense.expense_date)}
                        </td>

                        <td className="px-3 py-3 text-muted-foreground">
                          {expense.reference || "-"}
                        </td>

                        <td className="px-3 py-3 capitalize">
                          {expense.status}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>

        {showForm && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <Card className="w-full max-w-2xl">
              <CardContent className="pt-6">
                <div className="mb-6 flex items-start justify-between gap-4">
                  <div>
                    <h2 className="text-xl font-semibold">
                      Add Expense
                    </h2>

                    <p className="text-sm text-muted-foreground">
                      Record an expense for{" "}
                      {branchName || "your assigned branch"}.
                    </p>
                  </div>

                  <Button
                    type="button"
                    variant="outline"
                    className="h-10 gap-2 px-4"
                    onClick={closeForm}
                    disabled={saving}
                  >
                    <X className="h-4 w-4" />
                    Close
                  </Button>
                </div>

                <form
                  onSubmit={handleSubmit}
                  className="space-y-5"
                >
                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-2">
                      <label className="text-sm font-medium">
                        Category
                      </label>

                      <select
                        value={category}
                        onChange={(event) =>
                          setCategory(event.target.value)
                        }
                        className="h-10 w-full rounded-md border bg-background px-3 text-sm"
                        required
                      >
                        <option value="">
                          Select category
                        </option>

                        {CATEGORY_OPTIONS.map((option) => (
                          <option
                            key={option}
                            value={option}
                          >
                            {option}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">
                        Amount (KES)
                      </label>

                      <Input
                        type="number"
                        min="0.01"
                        step="0.01"
                        value={amount}
                        onChange={(event) =>
                          setAmount(event.target.value)
                        }
                        placeholder="0.00"
                        required
                      />
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">
                        Payment Method
                      </label>

                      <select
                        value={paymentMethod}
                        onChange={(event) =>
                          setPaymentMethod(
                            event.target
                              .value as ExpensePaymentMethod
                          )
                        }
                        className="h-10 w-full rounded-md border bg-background px-3 text-sm"
                        required
                      >
                        {PAYMENT_METHODS.map((method) => (
                          <option
                            key={method}
                            value={method}
                          >
                            {method.charAt(0).toUpperCase() +
                              method.slice(1)}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div className="space-y-2">
                      <label className="text-sm font-medium">
                        Expense Date
                      </label>

                      <Input
                        type="date"
                        value={expenseDate}
                        min={today}
                        max={today}
                        readOnly
                        required
                      />

                      <p className="text-xs text-muted-foreground">
                        Attendants can record expenses for today only.
                      </p>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label className="text-sm font-medium">
                      Description
                    </label>

                    <textarea
                      value={description}
                      onChange={(event) =>
                        setDescription(event.target.value)
                      }
                      placeholder="Optional description"
                      rows={3}
                      className="w-full rounded-md border bg-background px-3 py-2 text-sm outline-none focus:border-blue-600 focus:ring-1 focus:ring-blue-600"
                    />
                  </div>

                  <div className="space-y-2">
                    <label className="text-sm font-medium">
                      Reference
                    </label>

                    <Input
                      value={reference}
                      onChange={(event) =>
                        setReference(event.target.value)
                      }
                      placeholder="Receipt number, M-Pesa reference, etc."
                    />
                  </div>

                  <div className="flex justify-end gap-3 border-t pt-5">
                    <Button
                      type="button"
                      variant="outline"
                      className="h-10 gap-2 px-4"
                      onClick={closeForm}
                      disabled={saving}
                    >
                      <X className="h-4 w-4" />
                      Cancel
                    </Button>

                    <Button
                      type="submit"
                      disabled={
                        saving ||
                        !branchId ||
                        branchLoading
                      }
                      className="h-10 gap-2 rounded-md bg-blue-600 px-5 font-medium text-white shadow-sm hover:bg-blue-700"
                    >
                      {saving ? (
                        <>
                          <RefreshCw className="h-4 w-4 animate-spin" />
                          Saving...
                        </>
                      ) : (
                        <>
                          <Save className="h-4 w-4" />
                          Save Expense
                        </>
                      )}
                    </Button>
                  </div>
                </form>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </DashboardShell>
  );
}