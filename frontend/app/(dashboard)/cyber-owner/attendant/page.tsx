"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Check,
  Loader2,
  MoreHorizontal,
  Pencil,
  UserPlus,
  UserX,
  Users,
} from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";

import { Button } from "@/components/ui/button";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { Badge } from "@/components/ui/badge";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import { useAttendantStore } from "@/store/attendant.store";
import { useAuthStore } from "@/store/auth.store";

import { branchService } from "@/services/branch.service";

import type {
  Attendant,
  CreateAttendantRequest,
  UpdateAttendantRequest,
} from "@/types/attendant";

import type { Branch } from "@/types/branch";

function statusLabel(status: Attendant["status"]) {
  switch (status) {
    case "active":
      return (
        <Badge className="border-0 bg-green-100 text-green-700 hover:bg-green-100">
          Active
        </Badge>
      );

    case "inactive":
      return (
        <Badge className="border-0 bg-slate-100 text-slate-600 hover:bg-slate-100">
          Inactive
        </Badge>
      );

    case "suspended":
      return (
        <Badge className="border-0 bg-red-100 text-red-700 hover:bg-red-100">
          Suspended
        </Badge>
      );

    default:
      return null;
  }
}

export default function CyberOwnerAttendantsPage() {
  const {
    attendants,
    total,
    loading,
    saving,
    error,
    fetchAttendants,
    createAttendant,
    updateAttendant,
    changeStatus,
    assignBranch,
  } = useAttendantStore();

  const token = useAuthStore((state) => state.token);

  // --------------------------------------------------
  // Branches
  // --------------------------------------------------

  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchesLoading, setBranchesLoading] =
    useState(false);
  const [branchesError, setBranchesError] =
    useState<string | null>(null);

  // --------------------------------------------------
  // Add dialog
  // --------------------------------------------------

  const [addOpen, setAddOpen] = useState(false);

  const [addFullName, setAddFullName] =
    useState("");

  const [addEmail, setAddEmail] =
    useState("");

  const [addPhone, setAddPhone] =
    useState("");

  const [addPassword, setAddPassword] =
    useState("");

  const [addBranchId, setAddBranchId] =
    useState("");

  // --------------------------------------------------
  // Edit dialog
  // --------------------------------------------------

  const [editOpen, setEditOpen] = useState(false);

  const [editingAttendant, setEditingAttendant] =
    useState<Attendant | null>(null);

  const [editFullName, setEditFullName] =
    useState("");

  const [editEmail, setEditEmail] =
    useState("");

  const [editPhone, setEditPhone] =
    useState("");

  const [editBranchId, setEditBranchId] =
    useState("");

  // --------------------------------------------------
  // Status action
  // --------------------------------------------------

  const [statusChangingId, setStatusChangingId] =
    useState<string | null>(null);

  // --------------------------------------------------
  // Load data
  // --------------------------------------------------

  useEffect(() => {
    fetchAttendants();
  }, [fetchAttendants]);

  useEffect(() => {
    async function loadBranches() {
      if (!token) {
        return;
      }

      setBranchesLoading(true);
      setBranchesError(null);

      try {
        const response =
          await branchService.list(token);

        setBranches(response.branches);
      } catch (error) {
        setBranchesError(
          error instanceof Error
            ? error.message
            : "Failed to load branches."
        );
      } finally {
        setBranchesLoading(false);
      }
    }

    loadBranches();
  }, [token]);

  // --------------------------------------------------
  // Derived values
  // --------------------------------------------------

  const activeAttendants = useMemo(
    () =>
      attendants.filter(
        (attendant) =>
          attendant.status === "active"
      ).length,
    [attendants]
  );

  const assignedAttendants = useMemo(
    () =>
      attendants.filter(
        (attendant) =>
          !!attendant.branch_id
      ).length,
    [attendants]
  );

  // --------------------------------------------------
  // Add attendant
  // --------------------------------------------------

  function resetAddForm() {
    setAddFullName("");
    setAddEmail("");
    setAddPhone("");
    setAddPassword("");
    setAddBranchId("");
  }

  function handleOpenAdd() {
    resetAddForm();
    setAddOpen(true);
  }

  async function handleCreate(
    event: React.FormEvent<HTMLFormElement>
  ) {
    event.preventDefault();

    if (!addBranchId) {
      return;
    }

    const data: CreateAttendantRequest = {
      full_name: addFullName.trim(),
      email: addEmail.trim() || null,
      phone: addPhone.trim() || null,
      password: addPassword,
      branch_id: addBranchId,
    };

    await createAttendant(data);

    resetAddForm();
    setAddOpen(false);
  }

  // --------------------------------------------------
  // Edit attendant
  // --------------------------------------------------

  function handleOpenEdit(attendant: Attendant) {
    setEditingAttendant(attendant);

    setEditFullName(
      attendant.full_name
    );

    setEditEmail(
      attendant.email ?? ""
    );

    setEditPhone(
      attendant.phone ?? ""
    );

    setEditBranchId(
      attendant.branch_id ?? ""
    );

    setEditOpen(true);
  }

  function resetEditForm() {
    setEditingAttendant(null);
    setEditFullName("");
    setEditEmail("");
    setEditPhone("");
    setEditBranchId("");
  }

  async function handleEdit(
    event: React.FormEvent<HTMLFormElement>
  ) {
    event.preventDefault();

    if (!editingAttendant) {
      return;
    }

    const data: UpdateAttendantRequest = {
      full_name: editFullName.trim(),
      email: editEmail.trim() || null,
      phone: editPhone.trim() || null,
    };

    await updateAttendant(
      editingAttendant.id,
      data
    );

    const currentBranchId =
      editingAttendant.branch_id ?? "";

    if (
      editBranchId &&
      editBranchId !== currentBranchId
    ) {
      await assignBranch(
        editingAttendant.id,
        editBranchId
      );
    }

    setEditOpen(false);
    resetEditForm();

    await fetchAttendants();
  }

  // --------------------------------------------------
  // Status
  // --------------------------------------------------

  async function handleStatus(
    attendant: Attendant
  ) {
    setStatusChangingId(attendant.id);

    try {
      const nextStatus =
        attendant.status === "active"
          ? "inactive"
          : "active";

      await changeStatus(
        attendant.id,
        {
          status: nextStatus,
        }
      );
    } finally {
      setStatusChangingId(null);
    }
  }

  return (
    <DashboardShell role="owner">
      <div className="space-y-6">
        {/* --------------------------------------------------
            PAGE HEADER
        -------------------------------------------------- */}

        <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950">
              Attendants
            </h1>

            <p className="mt-1 text-sm text-slate-500">
              Manage the attendants working in your cyber cafés.
            </p>
          </div>

          <Button
            onClick={handleOpenAdd}
            disabled={branchesLoading}
            className="bg-[#0757B8] shadow-sm hover:bg-[#064A9D]"
          >
            <UserPlus className="mr-2 h-4 w-4" />
            Add Attendant
          </Button>
        </div>

        {/* --------------------------------------------------
            BRANCH ERROR
        -------------------------------------------------- */}

        {branchesError && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
            {branchesError}
          </div>
        )}

        {/* --------------------------------------------------
            SUMMARY
        -------------------------------------------------- */}

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Card className="shadow-sm">
            <CardContent className="flex items-center gap-4 p-5">
              <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
                <Users className="h-5 w-5" />
              </div>

              <div>
                <p className="text-sm text-slate-500">
                  Total attendants
                </p>

                <p className="text-2xl font-bold text-slate-950">
                  {total}
                </p>
              </div>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="p-5">
              <p className="text-sm text-slate-500">
                Active attendants
              </p>

              <p className="mt-1 text-2xl font-bold text-green-600">
                {activeAttendants}
              </p>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardContent className="p-5">
              <p className="text-sm text-slate-500">
                Assigned to branch
              </p>

              <p className="mt-1 text-2xl font-bold text-slate-950">
                {assignedAttendants}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* --------------------------------------------------
            ERROR
        -------------------------------------------------- */}

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
            {error}
          </div>
        )}

        {/* --------------------------------------------------
            TABLE
        -------------------------------------------------- */}

        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle>
              Your attendants
            </CardTitle>
          </CardHeader>

          <CardContent>
            {loading ? (
              <div className="flex items-center justify-center py-12 text-sm text-slate-500">
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Loading attendants...
              </div>
            ) : attendants.length === 0 ? (
              <div className="py-12 text-center">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-slate-100">
                  <Users className="h-6 w-6 text-slate-400" />
                </div>

                <h3 className="mt-4 font-semibold text-slate-900">
                  No attendants yet
                </h3>

                <p className="mt-1 text-sm text-slate-500">
                  Add your first attendant to get started.
                </p>

                <Button
                  onClick={handleOpenAdd}
                  className="mt-5 bg-[#0757B8] hover:bg-[#064A9D]"
                >
                  <UserPlus className="mr-2 h-4 w-4" />
                  Add Attendant
                </Button>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>
                        Attendant
                      </TableHead>

                      <TableHead>
                        Contact
                      </TableHead>

                      <TableHead>
                        Branch
                      </TableHead>

                      <TableHead>
                        Status
                      </TableHead>

                      <TableHead className="text-right">
                        Actions
                      </TableHead>
                    </TableRow>
                  </TableHeader>

                  <TableBody>
                    {attendants.map(
                      (attendant) => (
                        <TableRow
                          key={attendant.id}
                        >
                          <TableCell>
                            <div className="font-medium text-slate-900">
                              {attendant.full_name}
                            </div>

                            <div className="text-xs text-slate-500">
                              Attendant
                            </div>
                          </TableCell>

                          <TableCell>
                            <div className="text-sm text-slate-700">
                              {attendant.email ||
                                "No email"}
                            </div>

                            <div className="text-xs text-slate-500">
                              {attendant.phone ||
                                "No phone"}
                            </div>
                          </TableCell>

                          <TableCell>
                            {attendant.branch_name ? (
                              <span className="inline-flex items-center rounded-md bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-700">
                                {attendant.branch_name}
                              </span>
                            ) : (
                              <span className="text-sm text-slate-400">
                                Not assigned
                              </span>
                            )}
                          </TableCell>

                          <TableCell>
                            {statusLabel(
                              attendant.status
                            )}
                          </TableCell>

                          <TableCell className="text-right">
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() =>
                                  handleOpenEdit(
                                    attendant
                                  )
                                }
                                className="h-8 border-slate-200 bg-white px-3 text-slate-700 shadow-sm hover:bg-slate-50"
                              >
                                <Pencil className="mr-1.5 h-3.5 w-3.5" />
                                Edit
                              </Button>

                              <DropdownMenu>
                                <DropdownMenuTrigger
                                  className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 shadow-sm hover:bg-slate-50"
                                  aria-label="More attendant actions"
                                >
                                  <MoreHorizontal className="h-4 w-4" />
                                </DropdownMenuTrigger>

                                <DropdownMenuContent align="end">
                                  <DropdownMenuItem
                                    disabled={
                                      statusChangingId ===
                                      attendant.id
                                    }
                                    onClick={() =>
                                      handleStatus(
                                        attendant
                                      )
                                    }
                                  >
                                    {statusChangingId ===
                                    attendant.id ? (
                                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    ) : attendant.status ===
                                      "active" ? (
                                      <UserX className="mr-2 h-4 w-4 text-red-500" />
                                    ) : (
                                      <Check className="mr-2 h-4 w-4 text-green-600" />
                                    )}

                                    {attendant.status ===
                                    "active"
                                      ? "Deactivate"
                                      : "Activate"}
                                  </DropdownMenuItem>
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    )}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* ==================================================
          ADD ATTENDANT DIALOG
      ================================================== */}

      <Dialog
        open={addOpen}
        onOpenChange={setAddOpen}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              Add attendant
            </DialogTitle>

            <DialogDescription>
              Create an attendant account and assign them to a branch.
            </DialogDescription>
          </DialogHeader>

          <form
            onSubmit={handleCreate}
            className="space-y-5"
          >
            {/* Full name */}

            <div className="space-y-2">
              <Label htmlFor="add-full-name">
                Full name
              </Label>

              <Input
                id="add-full-name"
                placeholder="e.g. John Kamau"
                value={addFullName}
                onChange={(event) =>
                  setAddFullName(
                    event.target.value
                  )
                }
                required
              />
            </div>

            {/* Email */}

            <div className="space-y-2">
              <Label htmlFor="add-email">
                Email
              </Label>

              <Input
                id="add-email"
                type="email"
                placeholder="attendant@example.com"
                value={addEmail}
                onChange={(event) =>
                  setAddEmail(
                    event.target.value
                  )
                }
              />
            </div>

            {/* Phone */}

            <div className="space-y-2">
              <Label htmlFor="add-phone">
                Phone
              </Label>

              <Input
                id="add-phone"
                placeholder="0712345678"
                value={addPhone}
                onChange={(event) =>
                  setAddPhone(
                    event.target.value
                  )
                }
              />
            </div>

            {/* Password */}

            <div className="space-y-2">
              <Label htmlFor="add-password">
                Initial password
              </Label>

              <Input
                id="add-password"
                type="password"
                placeholder="Minimum 8 characters"
                minLength={8}
                value={addPassword}
                onChange={(event) =>
                  setAddPassword(
                    event.target.value
                  )
                }
                required
              />
            </div>

            {/* Branch */}

            <div className="space-y-2">
              <Label htmlFor="add-branch">
                Branch
              </Label>

              <select
                id="add-branch"
                value={addBranchId}
                onChange={(event) =>
                  setAddBranchId(
                    event.target.value
                  )
                }
                required
                disabled={branchesLoading}
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-[#0757B8]/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="">
                  {branchesLoading
                    ? "Loading branches..."
                    : branches.length === 0
                      ? "No branches available"
                      : "Select a branch"}
                </option>

                {branches.map(
                  (branch) => (
                    <option
                      key={branch.id}
                      value={branch.id}
                      disabled={
                        branch.status !==
                        "active"
                      }
                    >
                      {branch.name}
                      {branch.status !==
                        "active"
                        ? " (Inactive)"
                        : ""}
                    </option>
                  )
                )}
              </select>

              {branches.length === 0 &&
                !branchesLoading && (
                  <p className="text-xs text-amber-600">
                    Create a branch before adding an attendant.
                  </p>
                )}
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  setAddOpen(false)
                }
                disabled={saving}
              >
                Cancel
              </Button>

              <Button
                type="submit"
                disabled={
                  saving ||
                  branchesLoading ||
                  branches.length === 0 ||
                  !addBranchId
                }
                className="bg-[#0757B8] hover:bg-[#064A9D]"
              >
                {saving ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  <>
                    <UserPlus className="mr-2 h-4 w-4" />
                    Create Attendant
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* ==================================================
          EDIT ATTENDANT DIALOG
      ================================================== */}

      <Dialog
        open={editOpen}
        onOpenChange={(value) => {
          setEditOpen(value);

          if (!value) {
            resetEditForm();
          }
        }}
      >
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              Edit attendant
            </DialogTitle>

            <DialogDescription>
              Update the attendant's details and branch assignment.
            </DialogDescription>
          </DialogHeader>

          <form
            onSubmit={handleEdit}
            className="space-y-5"
          >
            {/* Full name */}

            <div className="space-y-2">
              <Label htmlFor="edit-full-name">
                Full name
              </Label>

              <Input
                id="edit-full-name"
                value={editFullName}
                onChange={(event) =>
                  setEditFullName(
                    event.target.value
                  )
                }
                required
              />
            </div>

            {/* Email */}

            <div className="space-y-2">
              <Label htmlFor="edit-email">
                Email
              </Label>

              <Input
                id="edit-email"
                type="email"
                value={editEmail}
                onChange={(event) =>
                  setEditEmail(
                    event.target.value
                  )
                }
              />
            </div>

            {/* Phone */}

            <div className="space-y-2">
              <Label htmlFor="edit-phone">
                Phone
              </Label>

              <Input
                id="edit-phone"
                placeholder="0712345678"
                value={editPhone}
                onChange={(event) =>
                  setEditPhone(
                    event.target.value
                  )
                }
              />
            </div>

            {/* Branch */}

            <div className="space-y-2">
              <Label htmlFor="edit-branch">
                Branch
              </Label>

              <select
                id="edit-branch"
                value={editBranchId}
                onChange={(event) =>
                  setEditBranchId(
                    event.target.value
                  )
                }
                disabled={branchesLoading}
                className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-[#0757B8] focus:ring-2 focus:ring-[#0757B8]/20 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="">
                  Select a branch
                </option>

                {branches.map(
                  (branch) => (
                    <option
                      key={branch.id}
                      value={branch.id}
                      disabled={
                        branch.status !==
                        "active"
                      }
                    >
                      {branch.name}
                      {branch.status !==
                        "active"
                        ? " (Inactive)"
                        : ""}
                    </option>
                  )
                )}
              </select>

              <p className="text-xs text-slate-500">
                Changing this selection moves the attendant to the selected branch.
              </p>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setEditOpen(false);
                  resetEditForm();
                }}
                disabled={saving}
              >
                Cancel
              </Button>

              <Button
                type="submit"
                disabled={
                  saving ||
                  branchesLoading ||
                  !editFullName.trim()
                }
                className="bg-[#0757B8] hover:bg-[#064A9D]"
              >
                {saving ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Check className="mr-2 h-4 w-4" />
                    Save Changes
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </DashboardShell>
  );
}