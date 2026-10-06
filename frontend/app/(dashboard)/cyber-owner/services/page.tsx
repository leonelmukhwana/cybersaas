"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Edit,
  Package,
  Plus,
  Power,
  RefreshCw,
  Search,
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
import { useServiceStore } from "@/store/service.store";
import { branchService } from "@/services/branch.service";

import type { Branch } from "@/types/branch";
import type { Service } from "@/types/service";

export default function ServicesPage() {
  const token = useAuthStore((state) => state.token);

  const {
    services,
    total,
    loading,
    saving,
    error,
    fetchServices,
    createService,
    updateService,
    toggleStatus,
    clearError,
  } = useServiceStore();

  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchId, setBranchId] = useState("");

  const [search, setSearch] = useState("");

  const [showForm, setShowForm] = useState(false);
  const [editingService, setEditingService] =
    useState<Service | null>(null);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [price, setPrice] = useState("");

  const [savingBranch, setSavingBranch] = useState(false);
  const [branchError, setBranchError] = useState("");

  const [success, setSuccess] = useState("");

  async function loadBranches() {
    if (!token) return;

    setSavingBranch(true);
    setBranchError("");

    try {
      const response = await branchService.list(token);

      setBranches(response.branches);

      setBranchId((currentBranchId) => {
        if (
          currentBranchId &&
          response.branches.some(
            (branch: Branch) =>
              branch.id === currentBranchId,
          )
        ) {
          return currentBranchId;
        }

        return response.branches[0]?.id ?? "";
      });
    } catch (err) {
      setBranchError(
        err instanceof Error
          ? err.message
          : "Failed to load branches.",
      );
    } finally {
      setSavingBranch(false);
    }
  }

  useEffect(() => {
    loadBranches();
  }, [token]);

  useEffect(() => {
    if (!token || !branchId) return;

    fetchServices(
      token,
      branchId,
      false,
    ).catch(() => {});
  }, [token, branchId, fetchServices]);

  function openCreateForm() {
    clearError();
    setSuccess("");

    setEditingService(null);
    setName("");
    setDescription("");
    setPrice("");

    setShowForm(true);
  }

  function openEditForm(service: Service) {
    clearError();
    setSuccess("");

    setEditingService(service);
    setName(service.name);
    setDescription(service.description ?? "");
    setPrice(service.price);

    setShowForm(true);
  }

  function closeForm() {
    if (saving) return;

    setShowForm(false);
    setEditingService(null);

    setName("");
    setDescription("");
    setPrice("");
  }

  async function handleSubmit(
    event: React.FormEvent<HTMLFormElement>,
  ) {
    event.preventDefault();

    if (!token) {
      clearError();
      return;
    }

    if (!branchId) {
      return;
    }

    const trimmedName = name.trim();
    const trimmedDescription =
      description.trim();
    const trimmedPrice = price.trim();

    if (!trimmedName) {
      return;
    }

    if (!trimmedPrice) {
      return;
    }

    clearError();
    setSuccess("");

    try {
      if (editingService) {
        await updateService(
          token,
          editingService.id,
          branchId,
          {
            name: trimmedName,
            description:
              trimmedDescription || null,
            price: trimmedPrice,
          },
        );

        setSuccess(
          "Service updated successfully.",
        );
      } else {
        await createService(token, {
          branch_id: branchId,
          name: trimmedName,
          description:
            trimmedDescription || null,
          price: trimmedPrice,
        });

        setSuccess(
          "Service created successfully.",
        );
      }

      setShowForm(false);
      setEditingService(null);
      setName("");
      setDescription("");
      setPrice("");

      await fetchServices(
        token,
        branchId,
        false,
      );
    } catch {
      // Error is stored by the service store.
    }
  }

  async function handleStatusChange(
    service: Service,
  ) {
    if (!token || !branchId) {
      return;
    }

    clearError();
    setSuccess("");

    try {
      await toggleStatus(
        token,
        service.id,
        branchId,
        !service.active,
      );

      setSuccess(
        service.active
          ? "Service deactivated successfully."
          : "Service activated successfully.",
      );
    } catch {
      // Error is stored by the service store.
    }
  }

  async function handleRefresh() {
    if (!token || !branchId) {
      return;
    }

    clearError();
    setSuccess("");

    try {
      await loadBranches();

      const currentBranch =
        branchId;

      if (currentBranch) {
        await fetchServices(
          token,
          currentBranch,
          false,
        );
      }
    } catch {
      // Errors are handled above/store.
    }
  }

  const filteredServices = useMemo(() => {
    const query = search
      .trim()
      .toLowerCase();

    if (!query) {
      return services;
    }

    return services.filter(
      (service: Service) =>
        service.name
          .toLowerCase()
          .includes(query) ||
        (service.description ?? "")
          .toLowerCase()
          .includes(query),
    );
  }, [services, search]);

  return (
    <DashboardShell role="owner">
      <PageHeader
        title="Services"
        description="Create and manage services offered at your cyber café branches."
        action={
          <Button
            onClick={openCreateForm}
            disabled={!branchId}
            className="bg-[#0757B8] text-white hover:bg-[#064A9D]"
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Service
          </Button>
        }
      />

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

      <Card className="mb-6 border-slate-200 shadow-sm">
        <CardContent className="pt-6">
          <div className="grid gap-5 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="service-branch">
                Branch
              </Label>

              <Select
                value={branchId}
                onValueChange={(value) => {
                  setBranchId(value ?? "");
                  setSearch("");
                  setSuccess("");
                  clearError();
                }}
              >
                <SelectTrigger
                  id="service-branch"
                  className="h-12"
                >
                  <SelectValue placeholder="Select a branch" />
                </SelectTrigger>

                <SelectContent>
                  {branches.map(
                    (branch: Branch) => (
                      <SelectItem
                        key={branch.id}
                        value={branch.id}
                      >
                        {branch.name}
                      </SelectItem>
                    ),
                  )}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="service-search">
                Search services
              </Label>

              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />

                <Input
                  id="service-search"
                  value={search}
                  onChange={(event) =>
                    setSearch(
                      event.target.value,
                    )
                  }
                  placeholder="Search by service name..."
                  className="h-12 pl-10"
                  disabled={!branchId}
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {showForm && (
        <Card className="mb-6 border-slate-200 shadow-sm">
          <CardHeader>
            <CardTitle>
              {editingService
                ? "Update Service"
                : "Add New Service"}
            </CardTitle>
          </CardHeader>

          <CardContent>
            <form
              onSubmit={handleSubmit}
              className="space-y-5"
            >
              <div className="grid gap-5 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="service-name">
                    Service name
                  </Label>

                  <Input
                    id="service-name"
                    value={name}
                    onChange={(event) =>
                      setName(
                        event.target.value,
                      )
                    }
                    placeholder="e.g. Printing"
                    disabled={saving}
                    className="h-12"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="service-price">
                    Price (KES)
                  </Label>

                  <Input
                    id="service-price"
                    type="number"
                    min="0"
                    step="0.01"
                    value={price}
                    onChange={(event) =>
                      setPrice(
                        event.target.value,
                      )
                    }
                    placeholder="e.g. 20.00"
                    disabled={saving}
                    className="h-12"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="service-description">
                  Description
                </Label>

                <Input
                  id="service-description"
                  value={description}
                  onChange={(event) =>
                    setDescription(
                      event.target.value,
                    )
                  }
                  placeholder="Optional service description"
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
                  disabled={
                    saving || !branchId
                  }
                  className="bg-[#0757B8] text-white hover:bg-[#064A9D]"
                >
                  {saving
                    ? "Saving..."
                    : editingService
                      ? "Update Service"
                      : "Create Service"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <Card className="border-slate-200 shadow-sm">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>
              Services
            </CardTitle>

            <p className="mt-1 text-sm text-slate-500">
              {filteredServices.length}{" "}
              {filteredServices.length === 1
                ? "service"
                : "services"}
              {total !==
                filteredServices.length &&
                ` • ${total} total`}
            </p>
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={
              loading ||
              savingBranch ||
              !branchId
            }
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${
                loading ||
                savingBranch
                  ? "animate-spin"
                  : ""
              }`}
            />
            Refresh
          </Button>
        </CardHeader>

        <CardContent>
          {!branchId ? (
            <div className="flex min-h-52 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-[#0757B8]">
                <Package className="h-6 w-6" />
              </div>

              <h3 className="mt-4 text-sm font-semibold text-slate-900">
                No branch selected
              </h3>

              <p className="mt-1 text-sm text-slate-500">
                Create a branch before adding services.
              </p>
            </div>
          ) : loading ? (
            <div className="flex min-h-40 items-center justify-center">
              <p className="text-sm text-slate-500">
                Loading services...
              </p>
            </div>
          ) : filteredServices.length ===
            0 ? (
            <div className="flex min-h-52 flex-col items-center justify-center rounded-xl border border-dashed border-slate-200">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-[#0757B8]">
                <Package className="h-6 w-6" />
              </div>

              <h3 className="mt-4 text-sm font-semibold text-slate-900">
                {search
                  ? "No services found"
                  : "No services yet"}
              </h3>

              <p className="mt-1 text-sm text-slate-500">
                {search
                  ? "Try a different search term."
                  : "Create your first service for this branch."}
              </p>

              {!search && (
                <Button
                  onClick={openCreateForm}
                  className="mt-4 bg-[#0757B8] hover:bg-[#064A9D]"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Add Service
                </Button>
              )}
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {filteredServices.map(
                (service: Service) => {
                  const active =
                    service.active;

                  return (
                    <div
                      key={service.id}
                      className="rounded-xl border border-slate-200 bg-white p-5 transition hover:shadow-md"
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex min-w-0 items-center gap-3">
                          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-[#0757B8]">
                            <Package className="h-5 w-5" />
                          </div>

                          <div className="min-w-0">
                            <h3 className="truncate font-semibold text-slate-900">
                              {service.name}
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

                      <div className="mt-5">
                        <p className="text-2xl font-bold text-slate-900">
                          KES{" "}
                          {Number(
                            service.price,
                          ).toLocaleString(
                            "en-KE",
                            {
                              minimumFractionDigits: 2,
                              maximumFractionDigits: 2,
                            },
                          )}
                        </p>

                        {service.description && (
                          <p className="mt-2 line-clamp-2 text-sm text-slate-500">
                            {
                              service.description
                            }
                          </p>
                        )}
                      </div>

                      <div className="mt-5 flex gap-2 border-t border-slate-100 pt-4">
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          className="flex-1"
                          onClick={() =>
                            openEditForm(
                              service,
                            )
                          }
                          disabled={saving}
                        >
                          <Edit className="mr-2 h-4 w-4" />
                          Edit
                        </Button>

                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handleStatusChange(
                              service,
                            )
                          }
                          disabled={saving}
                          className={
                            active
                              ? "text-red-600 hover:bg-red-50 hover:text-red-700"
                              : "text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
                          }
                        >
                          <Power className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  );
                },
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </DashboardShell>
  );
}