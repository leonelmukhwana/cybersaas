"use client";

import { useEffect, useState } from "react";
import {
  Edit,
  Plus,
  RefreshCw,
  Search,
  User,
  Users,
  WifiOff,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import { useAuthStore } from "@/store/auth.store";
import { useCustomerStore } from "@/store/customer.store";
import { attendantService } from "@/services/attendant.service";

import type {
  Customer,
  CustomerType,
} from "@/types/customer";

export default function AttendantCustomersPage() {
  const token = useAuthStore((state) => state.token);
  const isOffline = useAuthStore(
    (state) => state.isOffline,
  );
  const offlineBranchId = useAuthStore(
    (state) => state.offlineBranchId,
  );
  const offlineBranchName = useAuthStore(
    (state) => state.offlineBranchName,
  );

  const {
    customers,
    total,
    loading,
    saving,
    error,
    fetchCustomers,
    createCustomer,
    updateCustomer,
    clearError,
  } = useCustomerStore();

  const [branchId, setBranchId] = useState("");
  const [branchName, setBranchName] = useState("");

  const [search, setSearch] = useState("");

  const [showForm, setShowForm] = useState(false);

  const [editingCustomer, setEditingCustomer] =
    useState<Customer | null>(null);

  const [customerType, setCustomerType] =
    useState<CustomerType>("adult");

  const [fullName, setFullName] = useState("");
  const [phone, setPhone] = useState("");
  const [idNumber, setIdNumber] = useState("");
  const [parentName, setParentName] = useState("");
  const [parentPhone, setParentPhone] = useState("");
  const [parentIdNumber, setParentIdNumber] = useState("");

  const [success, setSuccess] = useState("");
  const [branchError, setBranchError] = useState("");

  /*
   * Load the attendant's assigned branch.
   *
   * Offline:
   * use the branch already stored locally.
   *
   * Online:
   * ask the backend for the authoritative assignment.
   */
  useEffect(() => {
    let cancelled = false;

    async function loadAssignedBranch() {
      setBranchError("");

      /*
       * Offline session.
       */
      if (isOffline) {
        if (!offlineBranchId) {
          setBranchError(
            "No assigned branch is available offline.",
          );
          return;
        }

        if (!cancelled) {
          setBranchId(offlineBranchId);
          setBranchName(
            offlineBranchName || "Assigned Branch",
          );
        }

        return;
      }

      /*
       * Online session.
       */
      if (!token) {
        setBranchError(
          "Your session has expired. Please login again.",
        );
        return;
      }

      try {
        const attendant =
          await attendantService.me(token);

        if (!attendant.branch_id) {
          setBranchError(
            "Your attendant account is not assigned to a branch.",
          );
          return;
        }

        if (!cancelled) {
          setBranchId(attendant.branch_id);
          setBranchName(
            attendant.branch_name ||
              "Assigned Branch",
          );
        }
      } catch (err) {
        if (!cancelled) {
          setBranchError(
            err instanceof Error
              ? err.message
              : "Failed to load your assigned branch.",
          );
        }
      }
    }

    loadAssignedBranch();

    return () => {
      cancelled = true;
    };
  }, [
    token,
    isOffline,
    offlineBranchId,
    offlineBranchName,
  ]);

  /*
   * Load customers when online.
   *
   * Offline customer storage will be connected
   * to Dexie/sync in the next step.
   */
  useEffect(() => {
    if (isOffline) {
      return;
    }

    if (!token || !branchId) {
      return;
    }

    fetchCustomers(
      token,
      branchId,
      search,
    ).catch(() => {});
  }, [
    token,
    branchId,
    search,
    isOffline,
    fetchCustomers,
  ]);

  function resetForm() {
    setEditingCustomer(null);
    setCustomerType("adult");
    setFullName("");
    setPhone("");
    setIdNumber("");
    setParentName("");
    setParentPhone("");
    setParentIdNumber("");
  }

  function openCreateForm() {
    clearError();
    setSuccess("");
    resetForm();
    setShowForm(true);
  }

  function openEditForm(customer: Customer) {
    clearError();
    setSuccess("");

    setEditingCustomer(customer);
    setCustomerType(customer.customer_type);
    setFullName(customer.full_name);
    setPhone(customer.phone ?? "");

    /*
     * ID numbers are intentionally not returned
     * by the backend.
     */
    setIdNumber("");

    setParentName(customer.parent_name ?? "");
    setParentPhone(customer.parent_phone ?? "");
    setParentIdNumber("");

    setShowForm(true);
  }

  function closeForm() {
    if (saving) return;

    setShowForm(false);
    resetForm();
  }

  async function handleSubmit(
    event: React.FormEvent<HTMLFormElement>,
  ) {
    event.preventDefault();

    if (!branchId) {
      setBranchError(
        "Your assigned branch could not be determined.",
      );
      return;
    }

    /*
     * Online customer operations currently use
     * the backend API.
     *
     * Offline create/update will be wired into
     * Dexie + sync queue next.
     */
    if (!token || isOffline) {
      setSuccess("");
      setBranchError(
        "Offline customer storage is being connected next.",
      );
      return;
    }

    const trimmedName = fullName.trim();
    const trimmedPhone = phone.trim();
    const trimmedID = idNumber.trim();
    const trimmedParentName = parentName.trim();
    const trimmedParentPhone = parentPhone.trim();
    const trimmedParentID = parentIdNumber.trim();

    if (!trimmedName) {
      setBranchError("");
      clearError();
      return;
    }

    if (
      customerType === "child" &&
      !trimmedParentName
    ) {
      return;
    }

    if (
      customerType === "child" &&
      !trimmedParentPhone
    ) {
      return;
    }

    clearError();
    setSuccess("");

    try {
      if (editingCustomer) {
        await updateCustomer(
          token,
          editingCustomer.id,
          branchId,
          {
            customer_type: customerType,
            full_name: trimmedName,
            phone: trimmedPhone || null,
            id_number: trimmedID || null,
            parent_name:
              customerType === "child"
                ? trimmedParentName || null
                : null,
            parent_phone:
              customerType === "child"
                ? trimmedParentPhone || null
                : null,
            parent_id_number:
              customerType === "child"
                ? trimmedParentID || null
                : null,
          },
        );

        setSuccess(
          "Customer updated successfully.",
        );
      } else {
        await createCustomer(
          token,
          {
            id: crypto.randomUUID(),
            branch_id: branchId,
            customer_type: customerType,
            full_name: trimmedName,
            phone: trimmedPhone || null,
            id_number: trimmedID || null,
            parent_name:
              customerType === "child"
                ? trimmedParentName || null
                : null,
            parent_phone:
              customerType === "child"
                ? trimmedParentPhone || null
                : null,
            parent_id_number:
              customerType === "child"
                ? trimmedParentID || null
                : null,
          },
        );

        setSuccess(
          "Customer registered successfully.",
        );
      }

      setShowForm(false);
      resetForm();

      await fetchCustomers(
        token,
        branchId,
        search,
      );
    } catch {
      // Customer store contains the error.
    }
  }

  async function handleRefresh() {
    if (isOffline) {
      setSuccess("");
      return;
    }

    if (!token || !branchId) {
      return;
    }

    setSuccess("");

    try {
      await fetchCustomers(
        token,
        branchId,
        search,
      );
    } catch {
      // Customer store contains the error.
    }
  }

  return (
    <DashboardShell role="attendant">
      <PageHeader
        title="Customers"
        description="Register and manage customers at your assigned branch."
        action={
          <Button
            onClick={openCreateForm}
            disabled={!branchId}
            className="bg-[#0757B8] text-white hover:bg-[#064A9D]"
          >
            <Plus className="mr-2 h-4 w-4" />
            Register Customer
          </Button>
        }
      />

      {/* BRANCH / CONNECTION INFO */}
      <div className="mb-6 flex flex-col gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-xs font-medium uppercase tracking-wide text-slate-400">
            Assigned Branch
          </p>

          <p className="mt-1 text-sm font-semibold text-slate-900">
            {branchName || "Loading branch..."}
          </p>
        </div>

        {isOffline && (
          <div className="inline-flex w-fit items-center gap-2 rounded-full bg-amber-50 px-3 py-1.5 text-xs font-semibold text-amber-700">
            <WifiOff className="h-3.5 w-3.5" />
            Offline mode
          </div>
        )}
      </div>

      {branchError && (
        <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
          {branchError}
        </div>
      )}

      {error && (
        <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-6 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {success}
        </div>
      )}

      {/* SEARCH */}
      <Card className="mb-6 border-slate-200 shadow-sm">
        <CardContent className="pt-6">
          <div className="space-y-2">
            <Label htmlFor="customer-search">
              Search customers
            </Label>

            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />

              <Input
                id="customer-search"
                value={search}
                onChange={(event) =>
                  setSearch(event.target.value)
                }
                placeholder="Search by customer name..."
                className="h-12 pl-10"
                disabled={!branchId}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* CUSTOMER FORM */}
      {showForm && (
        <Card className="mb-6 border-slate-200 shadow-sm">
          <CardHeader>
            <CardTitle>
              {editingCustomer
                ? "Update Customer"
                : "Register Customer"}
            </CardTitle>
          </CardHeader>

          <CardContent>
            <form
              onSubmit={handleSubmit}
              className="space-y-6"
            >
              <div className="space-y-2">
                <Label>Customer type</Label>

                <Select
                  value={customerType}
                  onValueChange={(value) => {
                    if (
                      value === "adult" ||
                      value === "child"
                    ) {
                      setCustomerType(value);

                      if (value === "adult") {
                        setParentName("");
                        setParentPhone("");
                        setParentIdNumber("");
                      }
                    }
                  }}
                  disabled={saving}
                >
                  <SelectTrigger className="h-12">
                    <SelectValue />
                  </SelectTrigger>

                  <SelectContent>
                    <SelectItem value="adult">
                      Adult
                    </SelectItem>

                    <SelectItem value="child">
                      Child
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="customer-name">
                    Full name
                  </Label>

                  <Input
                    id="customer-name"
                    value={fullName}
                    onChange={(event) =>
                      setFullName(event.target.value)
                    }
                    placeholder="Customer full name"
                    disabled={saving}
                    className="h-12"
                    required
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="customer-phone">
                    Phone number
                  </Label>

                  <Input
                    id="customer-phone"
                    type="tel"
                    value={phone}
                    onChange={(event) =>
                      setPhone(event.target.value)
                    }
                    placeholder="0712345678"
                    disabled={saving}
                    className="h-12"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="customer-id">
                  {customerType === "child"
                    ? "Child ID number (optional)"
                    : "ID number (optional)"}
                </Label>

                <Input
                  id="customer-id"
                  value={idNumber}
                  onChange={(event) =>
                    setIdNumber(event.target.value)
                  }
                  placeholder="National ID number"
                  disabled={saving}
                  className="h-12"
                />
              </div>

              {customerType === "child" && (
                <div className="rounded-xl border border-blue-100 bg-blue-50/50 p-5">
                  <div className="mb-5">
                    <h3 className="font-semibold text-slate-900">
                      Parent / Guardian
                    </h3>

                    <p className="mt-1 text-sm text-slate-500">
                      Parent or guardian details are required for child customers.
                    </p>
                  </div>

                  <div className="grid gap-5 md:grid-cols-2">
                    <div className="space-y-2">
                      <Label htmlFor="parent-name">
                        Parent / Guardian name
                      </Label>

                      <Input
                        id="parent-name"
                        value={parentName}
                        onChange={(event) =>
                          setParentName(
                            event.target.value,
                          )
                        }
                        placeholder="Parent or guardian name"
                        disabled={saving}
                        className="h-12"
                        required
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="parent-phone">
                        Parent / Guardian phone
                      </Label>

                      <Input
                        id="parent-phone"
                        type="tel"
                        value={parentPhone}
                        onChange={(event) =>
                          setParentPhone(
                            event.target.value,
                          )
                        }
                        placeholder="0712345678"
                        disabled={saving}
                        className="h-12"
                        required
                      />
                    </div>
                  </div>

                  <div className="mt-5 space-y-2">
                    <Label htmlFor="parent-id">
                      Parent / Guardian ID number
                    </Label>

                    <Input
                      id="parent-id"
                      value={parentIdNumber}
                      onChange={(event) =>
                        setParentIdNumber(
                          event.target.value,
                        )
                      }
                      placeholder="Parent ID number (optional)"
                      disabled={saving}
                      className="h-12"
                    />
                  </div>
                </div>
              )}

              <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                <Button
                  type="button"
                  variant="outline"
                  onClick={closeForm}
                  disabled={saving}
                >
                  Cancel
                </Button>

                <Button
                  type="submit"
                  disabled={saving || !branchId}
                  className="bg-[#0757B8] text-white hover:bg-[#064A9D]"
                >
                  {saving
                    ? "Saving..."
                    : editingCustomer
                      ? "Update Customer"
                      : "Register Customer"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* CUSTOMERS */}
      <Card className="border-slate-200 shadow-sm">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>
              Registered Customers
            </CardTitle>

            <p className="mt-1 text-sm text-slate-500">
              {total}{" "}
              {total === 1
                ? "customer"
                : "customers"}
            </p>
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={
              loading ||
              saving ||
              !branchId ||
              isOffline
            }
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${
                loading ? "animate-spin" : ""
              }`}
            />
            Refresh
          </Button>
        </CardHeader>

        <CardContent>
          {!branchId ? (
            <div className="flex min-h-52 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200">
              <Users className="h-8 w-8 text-slate-400" />

              <h3 className="mt-4 text-sm font-semibold text-slate-900">
                No assigned branch
              </h3>

              <p className="mt-1 text-sm text-slate-500">
                Your branch assignment is required.
              </p>
            </div>
          ) : loading ? (
            <div className="flex min-h-40 items-center justify-center">
              <p className="text-sm text-slate-500">
                Loading customers...
              </p>
            </div>
          ) : customers.length === 0 ? (
            <div className="flex min-h-52 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-[#0757B8]">
                <User className="h-6 w-6" />
              </div>

              <h3 className="mt-4 text-sm font-semibold text-slate-900">
                No customers yet
              </h3>

              <p className="mt-1 text-sm text-slate-500">
                Register the first customer at this branch.
              </p>

              <Button
                onClick={openCreateForm}
                className="mt-4 bg-[#0757B8] hover:bg-[#064A9D]"
              >
                <Plus className="mr-2 h-4 w-4" />
                Register Customer
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[700px]">
                <thead>
                  <tr className="border-b border-slate-200 text-left">
                    <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                      Customer
                    </th>

                    <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                      Type
                    </th>

                    <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                      Phone
                    </th>

                    <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                      Parent / Guardian
                    </th>

                    <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
                      Action
                    </th>
                  </tr>
                </thead>

                <tbody>
                  {customers.map(
                    (customer: Customer) => (
                      <tr
                        key={customer.id}
                        className="border-b border-slate-100 last:border-0"
                      >
                        <td className="px-4 py-4">
                          <div className="flex items-center gap-3">
                            <div className="flex h-9 w-9 items-center justify-center rounded-full bg-blue-50 text-[#0757B8]">
                              {customer.customer_type ===
                              "child" ? (
                                <User className="h-4 w-4" />
                              ) : (
                                <Users className="h-4 w-4" />
                              )}
                            </div>

                            <div>
                              <p className="font-medium text-slate-900">
                                {customer.full_name}
                              </p>

                              <p className="text-xs text-slate-400">
                                Registered customer
                              </p>
                            </div>
                          </div>
                        </td>

                        <td className="px-4 py-4">
                          <span
                            className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${
                              customer.customer_type ===
                              "child"
                                ? "bg-purple-50 text-purple-700"
                                : "bg-blue-50 text-blue-700"
                            }`}
                          >
                            {customer.customer_type ===
                            "child"
                              ? "Child"
                              : "Adult"}
                          </span>
                        </td>

                        <td className="px-4 py-4 text-sm text-slate-600">
                          {customer.phone || "—"}
                        </td>

                        <td className="px-4 py-4">
                          {customer.customer_type ===
                          "child" ? (
                            <div>
                              <p className="text-sm font-medium text-slate-700">
                                {customer.parent_name}
                              </p>

                              <p className="text-xs text-slate-500">
                                {customer.parent_phone}
                              </p>
                            </div>
                          ) : (
                            <span className="text-sm text-slate-400">
                              —
                            </span>
                          )}
                        </td>

                        <td className="px-4 py-4 text-right">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              openEditForm(customer)
                            }
                            disabled={
                              saving || isOffline
                            }
                          >
                            <Edit className="mr-2 h-4 w-4" />
                            Edit
                          </Button>
                        </td>
                      </tr>
                    ),
                  )}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </DashboardShell>
  );
}