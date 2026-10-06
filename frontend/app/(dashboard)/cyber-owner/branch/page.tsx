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
          : "Failed to load branches."
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
    event: React.FormEvent<HTMLFormElement>
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
          token
        );

        setSuccess("Branch updated successfully.");
      } else {
        /*
         * Your current backend CreateBranchRequest requires
         * business_name.
         *
         * For now we use the authenticated owner's business
         * name from the backend flow. If the backend changes
         * this request later, we can remove this field.
         */
        await branchService.create(
          {
            business_name: trimmedName,
            name: trimmedName,
            address: trimmedAddress || undefined,
            phone: trimmedPhone || undefined,
          },
          token
        );

        setSuccess("Branch created successfully.");
      }

      closeForm();
      await loadBranches();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to save branch."
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
        token
      );

      setSuccess(
        newStatus === "active"
          ? "Branch activated successfully."
          : "Branch deactivated successfully."
      );

      await loadBranches();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to change branch status."
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

        <CardContent>
          {loading ? (
            <div className="flex min-h-40 items-center justify-center">
              <p className="text-sm text-slate-500">
                Loading branches...
              </p>
            </div>
          ) : branches.length === 0 ? (
            <div className="flex min-h-52 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200">
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
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {branches.map((branch) => {
                const active = branch.status === "active";

                return (
                  <div
                    key={branch.id}
                    className="rounded-xl border border-slate-200 bg-white p-5 transition hover:shadow-md"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex min-w-0 items-center gap-3">
                        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-[#0757B8]">
                          <Building2 className="h-5 w-5" />
                        </div>

                        <div className="min-w-0">
                          <h3 className="truncate font-semibold text-slate-900">
                            {branch.name}
                          </h3>

                          <span
                            className={`mt-1 inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                              active
                                ? "bg-emerald-50 text-emerald-700"
                                : "bg-slate-100 text-slate-600"
                            }`}
                          >
                            {active
                              ? "Active"
                              : "Inactive"}
                          </span>
                        </div>
                      </div>
                    </div>

                    <div className="mt-5 space-y-3">
                      {branch.address && (
                        <div className="flex items-start gap-2 text-sm text-slate-500">
                          <MapPin className="mt-0.5 h-4 w-4 shrink-0" />
                          <span>{branch.address}</span>
                        </div>
                      )}

                      {branch.phone && (
                        <div className="flex items-center gap-2 text-sm text-slate-500">
                          <Phone className="h-4 w-4 shrink-0" />
                          <span>{branch.phone}</span>
                        </div>
                      )}
                    </div>

                    <div className="mt-5 flex gap-2 border-t border-slate-100 pt-4">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        className="flex-1"
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
                        <Power className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </DashboardShell>
  );
}