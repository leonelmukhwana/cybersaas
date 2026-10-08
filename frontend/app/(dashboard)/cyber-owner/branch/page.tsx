"use client";

import { useEffect, useState } from "react";
import {
  Building2,
  Edit,
  MapPin,
  Phone,
  Plus,
  Power,
  RefreshCw,
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

import { useAuthStore } from "@/store/auth.store";
import { branchService } from "@/services/branch.service";

import type { Branch } from "@/types/branch";

export default function BranchesPage() {
  const token = useAuthStore((state) => state.token);

  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [showForm, setShowForm] = useState(false);
  const [editingBranch, setEditingBranch] =
    useState<Branch | null>(null);

  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [phone, setPhone] = useState("");

  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function loadBranches() {
    if (!token) return;

    setLoading(true);
    setError("");

    try {
      const response = await branchService.list(token);
      setBranches(response.branches);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load branches.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadBranches();
  }, [token]);

  function openCreateForm() {
    setEditingBranch(null);
    setName("");
    setAddress("");
    setPhone("");
    setError("");
    setSuccess("");
    setShowForm(true);
  }

  function openEditForm(branch: Branch) {
    setEditingBranch(branch);
    setName(branch.name);
    setAddress(branch.address ?? "");
    setPhone(branch.phone ?? "");
    setError("");
    setSuccess("");
    setShowForm(true);
  }

  function closeForm() {
    if (saving) return;

    setShowForm(false);
    setEditingBranch(null);
    setName("");
    setAddress("");
    setPhone("");
  }

  async function handleSubmit(
    event: React.FormEvent<HTMLFormElement>,
  ) {
    event.preventDefault();

    if (!token) {
      setError("You are not authenticated.");
      return;
    }

    const trimmedName = name.trim();
    const trimmedAddress = address.trim();
    const trimmedPhone = phone.trim();

    if (!trimmedName) {
      setError("Branch name is required.");
      return;
    }

    setSaving(true);
    setError("");
    setSuccess("");

    try {
      if (editingBranch) {
        await branchService.update(
          editingBranch.id,
          {
            name: trimmedName,
            address: trimmedAddress || undefined,
            phone: trimmedPhone || undefined,
          },
          token,
        );

        setSuccess("Branch updated successfully.");
      } else {
        /*
         * Your current backend CreateBranchRequest requires
         * business_name.
         *
         * For now we use the branch name for this field.
         */
        await branchService.create(
          {
            business_name: trimmedName,
            name: trimmedName,
            address: trimmedAddress || undefined,
            phone: trimmedPhone || undefined,
          },
          token,
        );

        setSuccess("Branch created successfully.");
      }

      closeForm();
      await loadBranches();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to save branch.",
      );
    } finally {
      setSaving(false);
    }
  }

  async function handleStatusChange(branch: Branch) {
    if (!token) {
      setError("You are not authenticated.");
      return;
    }

    const newStatus =
      branch.status === "active" ? "inactive" : "active";

    setError("");
    setSuccess("");

    try {
      await branchService.changeStatus(
        branch.id,
        {
          status: newStatus,
        },
        token,
      );

      setSuccess(
        newStatus === "active"
          ? "Branch activated successfully."
          : "Branch deactivated successfully.",
      );

      await loadBranches();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to change branch status.",
      );
    }
  }

  return (
    <DashboardShell role="owner">
      <PageHeader
        title="Branches"
        description="Create and manage your cyber café branches."
        action={
          <Button
            onClick={openCreateForm}
            className="bg-[#0757B8] text-white hover:bg-[#064A9D]"
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Branch
          </Button>
        }
      />

      {/* MESSAGES */}
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

      {/* CREATE / EDIT FORM */}
      {showForm && (
        <Card className="mb-6 border-slate-200 shadow-sm">
          <CardHeader>
            <CardTitle>
              {editingBranch
                ? "Update Branch"
                : "Add New Branch"}
            </CardTitle>
          </CardHeader>

          <CardContent>
            <form
              onSubmit={handleSubmit}
              className="space-y-5"
            >
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="branch-name">
                    Branch name
                  </Label>

                  <Input
                    id="branch-name"
                    value={name}
                    onChange={(event) =>
                      setName(event.target.value)
                    }
                    placeholder="e.g. CBD Branch"
                    disabled={saving}
                    className="h-12"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="branch-phone">
                    Phone number
                  </Label>

                  <Input
                    id="branch-phone"
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
                <Label htmlFor="branch-address">
                  Address
                </Label>

                <Input
                  id="branch-address"
                  value={address}
                  onChange={(event) =>
                    setAddress(event.target.value)
                  }
                  placeholder="e.g. Tom Mboya Street"
                  disabled={saving}
                  className="h-12"
                />
              </div>

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
                  disabled={saving}
                  className="bg-[#0757B8] text-white hover:bg-[#064A9D]"
                >
                  {saving
                    ? "Saving..."
                    : editingBranch
                      ? "Update Branch"
                      : "Create Branch"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* BRANCHES */}
      <Card className="border-slate-200 shadow-sm">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Your Branches</CardTitle>

            <p className="mt-1 text-sm text-slate-500">
              {branches.length}{" "}
              {branches.length === 1
                ? "branch"
                : "branches"}
            </p>
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={loadBranches}
            disabled={loading}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${
                loading ? "animate-spin" : ""
              }`}
            />
            Refresh
          </Button>
        </CardHeader>

        <CardContent className="p-0">
          {loading ? (
            <div className="flex min-h-40 items-center justify-center">
              <p className="text-sm text-slate-500">
                Loading branches...
              </p>
            </div>
          ) : branches.length === 0 ? (
            <div className="m-6 flex min-h-52 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-[#0757B8]">
                <Building2 className="h-6 w-6" />
              </div>

              <h3 className="mt-4 text-sm font-semibold text-slate-900">
                No branches yet
              </h3>

              <p className="mt-1 text-sm text-slate-500">
                Create your first cyber café branch.
              </p>

              <Button
                onClick={openCreateForm}
                className="mt-4 bg-[#0757B8] hover:bg-[#064A9D]"
              >
                <Plus className="mr-2 h-4 w-4" />
                Add Branch
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[760px] text-left text-sm">
                <thead className="border-y border-slate-200 bg-slate-50">
                  <tr>
                    <th className="px-6 py-3 font-semibold text-slate-600">
                      Branch
                    </th>

                    <th className="px-6 py-3 font-semibold text-slate-600">
                      Address
                    </th>

                    <th className="px-6 py-3 font-semibold text-slate-600">
                      Phone
                    </th>

                    <th className="px-6 py-3 font-semibold text-slate-600">
                      Status
                    </th>

                    <th className="px-6 py-3 text-right font-semibold text-slate-600">
                      Actions
                    </th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-slate-100">
                  {branches.map((branch) => {
                    const active =
                      branch.status === "active";

                    return (
                      <tr
                        key={branch.id}
                        className="transition hover:bg-slate-50"
                      >
                        {/* BRANCH */}
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-[#0757B8]">
                              <Building2 className="h-4 w-4" />
                            </div>

                            <span className="font-medium text-slate-900">
                              {branch.name}
                            </span>
                          </div>
                        </td>

                        {/* ADDRESS */}
                        <td className="max-w-xs px-6 py-4 text-slate-500">
                          {branch.address ? (
                            <div className="flex items-center gap-2">
                              <MapPin className="h-4 w-4 shrink-0" />

                              <span
                                className="block truncate"
                                title={branch.address}
                              >
                                {branch.address}
                              </span>
                            </div>
                          ) : (
                            "—"
                          )}
                        </td>

                        {/* PHONE */}
                        <td className="whitespace-nowrap px-6 py-4 text-slate-500">
                          {branch.phone ? (
                            <div className="flex items-center gap-2">
                              <Phone className="h-4 w-4 shrink-0" />
                              <span>
                                {branch.phone}
                              </span>
                            </div>
                          ) : (
                            "—"
                          )}
                        </td>

                        {/* STATUS */}
                        <td className="px-6 py-4">
                          <span
                            className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${
                              active
                                ? "bg-emerald-50 text-emerald-700"
                                : "bg-slate-100 text-slate-600"
                            }`}
                          >
                            {active
                              ? "Active"
                              : "Inactive"}
                          </span>
                        </td>

                        {/* ACTIONS */}
                        <td className="px-6 py-4">
                          <div className="flex justify-end gap-2">
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                openEditForm(branch)
                              }
                            >
                              <Edit className="mr-2 h-4 w-4" />
                              Edit
                            </Button>

                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              className={
                                active
                                  ? "text-red-600 hover:bg-red-50 hover:text-red-700"
                                  : "text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
                              }
                              onClick={() =>
                                handleStatusChange(branch)
                              }
                            >
                              <Power className="mr-2 h-4 w-4" />
                              {active
                                ? "Deactivate"
                                : "Activate"}
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </DashboardShell>
  );
}