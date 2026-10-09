"use client";

import { useEffect, useState } from "react";
import { BadgeDollarSign, RefreshCw, Save, Settings2 } from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
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
import { branchService } from "@/services/branch.service";
import { billingService } from "@/services/billing.service";

import type { Branch } from "@/types/branch";
import type { BillingConfig, BillingConfigInput } from "@/types/billing";

const EMPTY_CONFIG: BillingConfig = {
  branch_id: "",
  rate_per_minute: "",
  minimum_charge: "",
  billing_interval_minutes: 1,
  rounding_mode: "up",
  currency: "KES",
};

export default function BillingPage() {
  const token = useAuthStore((state) => state.token);

  const [branches, setBranches] = useState<Branch[]>([]);
  const [branchId, setBranchId] = useState("");
  const [config, setConfig] = useState<BillingConfig>(EMPTY_CONFIG);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  async function loadBranches() {
    if (!token) return;

    setLoading(true);
    setError("");

    try {
      const response = await branchService.list(token);
      setBranches(response.branches);

      setBranchId((current) =>
        current && response.branches.some((b) => b.id === current)
          ? current
          : response.branches[0]?.id ?? "",
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load branches.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadBranches();
    // Load the initial branch list when the authentication token changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    if (!token || !branchId) {
      setConfig({ ...EMPTY_CONFIG, branch_id: branchId });
      return;
    }

    let cancelled = false;

    async function loadConfig() {
      setLoading(true);
      setError("");
      setSuccess("");

      try {
        const result = await billingService.get(token!, branchId);
        if (!cancelled) setConfig(result);
      } catch (err) {
        if (!cancelled) {
          setConfig({ ...EMPTY_CONFIG, branch_id: branchId });
          setError(
            err instanceof Error
              ? err.message
              : "Could not load billing settings.",
          );
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void loadConfig();

    return () => {
      cancelled = true;
    };
  }, [token, branchId]);

  async function handleSave(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!token || !branchId) {
      setError("Select a branch first.");
      return;
    }

    const rate = Number(config.rate_per_minute);
    const minimum = Number(config.minimum_charge);

    if (!Number.isFinite(rate) || rate <= 0) {
      setError("Rate per minute must be greater than zero.");
      return;
    }

    if (!Number.isFinite(minimum) || minimum < 0) {
      setError("Minimum charge cannot be negative.");
      return;
    }

    setSaving(true);
    setError("");
    setSuccess("");

    try {
      const input: BillingConfigInput = {
        rate_per_minute: rate.toFixed(2),
        minimum_charge: minimum.toFixed(2),
        billing_interval_minutes: config.billing_interval_minutes,
        rounding_mode: config.rounding_mode,
        currency: "KES",
      };

      const saved = await billingService.save(token, branchId, input);
      setConfig(saved);
      setSuccess("Billing settings saved successfully.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save billing settings.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <DashboardShell role="owner">
      <PageHeader
        title="Billing Configuration"
        description="Configure computer usage pricing separately for each branch."
        action={
          <Button
            variant="outline"
            onClick={() => void loadBranches()}
            disabled={loading || saving}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        }
      />

      {error && (
        <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-6 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {success}
        </div>
      )}

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.5fr)_minmax(280px,1fr)]">
        <Card className="border-slate-200 shadow-sm">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
                <Settings2 className="h-5 w-5" />
              </div>
              <div>
                <CardTitle>Branch usage rates</CardTitle>
                <CardDescription>
                  These rates are used by terminal session billing.
                </CardDescription>
              </div>
            </div>
          </CardHeader>

          <CardContent>
            <form onSubmit={handleSave} className="space-y-5">
              <div className="space-y-2">
                <Label htmlFor="billing-branch">Branch</Label>
                <Select
                  value={branchId}
                  onValueChange={(value) => {
                    setBranchId(value ?? "");
                    setError("");
                    setSuccess("");
                  }}
                >
                  <SelectTrigger id="billing-branch" className="h-12">
                    <SelectValue placeholder="Select a branch" />
                  </SelectTrigger>
                  <SelectContent>
                    {branches.map((branch) => (
                      <SelectItem key={branch.id} value={branch.id}>
                        {branch.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-5 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="rate">Rate per minute (KES)</Label>
                  <Input
                    id="rate"
                    type="number"
                    min="0.01"
                    step="0.01"
                    required
                    className="h-12"
                    value={config.rate_per_minute}
                    onChange={(e) =>
                      setConfig({ ...config, rate_per_minute: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="minimum">Minimum charge (KES)</Label>
                  <Input
                    id="minimum"
                    type="number"
                    min="0"
                    step="0.01"
                    required
                    className="h-12"
                    value={config.minimum_charge}
                    onChange={(e) =>
                      setConfig({ ...config, minimum_charge: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="interval">Billing interval</Label>
                  <Select
                    value={String(config.billing_interval_minutes)}
                    onValueChange={(value) =>
                      setConfig({
                        ...config,
                        billing_interval_minutes: Number(value ?? "1"),
                      })
                    }
                  >
                    <SelectTrigger id="interval" className="h-12">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {[1, 5, 10, 15].map((minutes) => (
                        <SelectItem key={minutes} value={String(minutes)}>
                          Every {minutes} minute{minutes > 1 ? "s" : ""}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="rounding">Rounding mode</Label>
                  <Select
                    value={config.rounding_mode}
                    onValueChange={(value) =>
                      setConfig({
                        ...config,
                        rounding_mode: (value ?? "up") as BillingConfig["rounding_mode"],
                      })
                    }
                  >
                    <SelectTrigger id="rounding" className="h-12">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="up">Round up</SelectItem>
                      <SelectItem value="down">Round down</SelectItem>
                      <SelectItem value="nearest">Round to nearest</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="rounded-lg bg-slate-50 p-4 text-sm text-slate-600">
                Currency is fixed to Kenyan Shillings (KES).
              </div>

              <div className="flex justify-end">
                <Button
                  type="submit"
                  disabled={loading || saving || !branchId}
                  className="w-full bg-[#0757B8] text-white hover:bg-[#064A9D] sm:w-auto"
                >
                  <Save className="mr-2 h-4 w-4" />
                  {saving ? "Saving..." : "Save Billing Settings"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <Card className="h-fit border-slate-200 shadow-sm">
          <CardHeader>
            <BadgeDollarSign className="h-8 w-8 text-emerald-700" />
            <CardTitle>Current configuration</CardTitle>
            <CardDescription>
              Review the selected branch's pricing.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border p-4">
              <p className="text-sm text-slate-500">Rate per minute</p>
              <p className="mt-1 text-2xl font-semibold">
                KES {Number(config.rate_per_minute || 0).toFixed(2)}
              </p>
            </div>
            <div className="rounded-lg border p-4">
              <p className="text-sm text-slate-500">Minimum charge</p>
              <p className="mt-1 text-2xl font-semibold">
                KES {Number(config.minimum_charge || 0).toFixed(2)}
              </p>
            </div>
            <p className="text-xs leading-5 text-slate-500">
              The terminal reads billing settings from the backend. Saving
              requires the matching backend API to be implemented.
            </p>
          </CardContent>
        </Card>
      </div>
    </DashboardShell>
  );
}
