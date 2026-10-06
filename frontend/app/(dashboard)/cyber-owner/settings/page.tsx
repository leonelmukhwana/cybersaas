
"use client";

import { useEffect, useState } from "react";
import {
  CheckCircle2,
  CreditCard,
  Eye,
  EyeOff,
  Loader2,
  Save,
  ShieldCheck,
  Trash2,
} from "lucide-react";



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
import {
  mpesaService,
  type MpesaConfiguration,
  type SaveMpesaConfigurationRequest,
} from "@/services/mpesa.service";

import type { Branch } from "@/types/branch";
import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

export default function CyberOwnerSettingsPage() {
  const { user, token } = useAuthStore();

  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedBranchId, setSelectedBranchId] =
    useState("");

  const [configuration, setConfiguration] =
    useState<MpesaConfiguration | null>(null);

  const [loadingBranches, setLoadingBranches] =
    useState(true);
  const [loadingConfiguration, setLoadingConfiguration] =
    useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const [showConsumerSecret, setShowConsumerSecret] =
    useState(false);
  const [showPasskey, setShowPasskey] = useState(false);

  const [provider, setProvider] = useState("daraja");
  const [environment, setEnvironment] = useState<
    "sandbox" | "production"
  >("sandbox");

  const [businessShortCode, setBusinessShortCode] =
    useState("");
  const [tillNumber, setTillNumber] = useState("");
  const [paybillNumber, setPaybillNumber] =
    useState("");

  const [consumerKey, setConsumerKey] = useState("");
  const [consumerSecret, setConsumerSecret] =
    useState("");
  const [passkey, setPasskey] = useState("");

  const [accountReference, setAccountReference] =
    useState("");
  const [callbackUrl, setCallbackUrl] = useState("");

  const [active, setActive] = useState(true);

  useEffect(() => {
    if (!token || user?.role !== "owner") {
      return;
    }

    const loadBranches = async () => {
      setLoadingBranches(true);
      setError("");

      try {
        const response = await branchService.list(token);

        setBranches(response.branches);

        if (response.branches.length > 0) {
          setSelectedBranchId(response.branches[0].id);
        }
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "Failed to load branches.",
        );
      } finally {
        setLoadingBranches(false);
      }
    };

    void loadBranches();
  }, [token, user?.role]);

  useEffect(() => {
    if (!selectedBranchId) {
      return;
    }

    const loadConfiguration = async () => {
      setLoadingConfiguration(true);
      setError("");
      setSuccess("");

      try {
        const data =
          await mpesaService.getBranch(
            selectedBranchId,
          );

        setConfiguration(data);

        setProvider(data.provider || "daraja");

        setEnvironment(
          data.environment === "production"
            ? "production"
            : "sandbox",
        );

        setBusinessShortCode(
          data.business_short_code || "",
        );

        setTillNumber(data.till_number || "");

        setPaybillNumber(
          data.paybill_number || "",
        );

        setAccountReference(
          data.account_reference || "",
        );

        setCallbackUrl(
          data.callback_url || "",
        );

        setActive(data.active);

        // Never load encrypted secrets into the form.
        setConsumerKey("");
        setConsumerSecret("");
        setPasskey("");
      } catch (err) {
        const message =
          err instanceof Error
            ? err.message
            : "Failed to load M-Pesa configuration.";

        // A missing configuration is allowed.
        if (
          message.toLowerCase().includes("not found") ||
          message.toLowerCase().includes("no mpesa")
        ) {
          setConfiguration(null);
          resetForm();
        } else {
          setError(message);
        }
      } finally {
        setLoadingConfiguration(false);
      }
    };

    void loadConfiguration();
  }, [selectedBranchId]);

  const resetForm = () => {
    setConfiguration(null);
    setProvider("daraja");
    setEnvironment("sandbox");
    setBusinessShortCode("");
    setTillNumber("");
    setPaybillNumber("");
    setConsumerKey("");
    setConsumerSecret("");
    setPasskey("");
    setAccountReference("");
    setCallbackUrl("");
    setActive(true);
  };

  const handleSave = async () => {
    if (!selectedBranchId) {
      setError("Please select a branch.");
      return;
    }

    if (!businessShortCode.trim()) {
      setError("Business shortcode is required.");
      return;
    }

    if (
      !tillNumber.trim() &&
      !paybillNumber.trim()
    ) {
      setError(
        "Enter either a Till Number or a Paybill Number.",
      );
      return;
    }

    if (
      tillNumber.trim() &&
      paybillNumber.trim()
    ) {
      setError(
        "Use either a Till Number or a Paybill Number, not both.",
      );
      return;
    }

    if (!accountReference.trim()) {
      setError("Account reference is required.");
      return;
    }

    if (!callbackUrl.trim()) {
      setError("Callback URL is required.");
      return;
    }

    if (
      !callbackUrl
        .trim()
        .toLowerCase()
        .startsWith("https://")
    ) {
      setError(
        "Callback URL must use HTTPS.",
      );
      return;
    }

    setSaving(true);
    setError("");
    setSuccess("");

    try {
      const data: SaveMpesaConfigurationRequest = {
        provider,
        environment,
        business_short_code:
          businessShortCode.trim(),
        account_reference:
          accountReference.trim(),
        callback_url: callbackUrl.trim(),
        active,
      };

      if (tillNumber.trim()) {
        data.till_number =
          tillNumber.trim();
      }

      if (paybillNumber.trim()) {
        data.paybill_number =
          paybillNumber.trim();
      }

      /*
       * IMPORTANT:
       * Only send credentials when the owner
       * actually entered them.
       *
       * This prevents an existing encrypted
       * credential from being accidentally
       * replaced by an empty value.
       */
      if (consumerKey.trim()) {
        data.consumer_key =
          consumerKey.trim();
      }

      if (consumerSecret.trim()) {
        data.consumer_secret =
          consumerSecret.trim();
      }

      if (passkey.trim()) {
        data.passkey = passkey.trim();
      }

      const saved =
        await mpesaService.saveBranch(
          selectedBranchId,
          data,
        );

      setConfiguration(saved);

      setConsumerKey("");
      setConsumerSecret("");
      setPasskey("");

      setSuccess(
        "M-Pesa configuration saved successfully.",
      );
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to save M-Pesa configuration.",
      );
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedBranchId) {
      return;
    }

    const confirmed = window.confirm(
      "Delete the M-Pesa configuration for this branch?",
    );

    if (!confirmed) {
      return;
    }

    setDeleting(true);
    setError("");
    setSuccess("");

    try {
      await mpesaService.deleteBranch(
        selectedBranchId,
      );

      resetForm();

      setSuccess(
        "M-Pesa configuration deleted.",
      );
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to delete M-Pesa configuration.",
      );
    } finally {
      setDeleting(false);
    }
  };

  const selectedBranch = branches.find(
    (branch) =>
      branch.id === selectedBranchId,
  );

  return (
    <DashboardShell role="owner">
      <div className="space-y-6">
        <PageHeader
          title="Settings"
          description="Manage your cyber branch payment configuration."
        />

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {success && (
          <div className="flex items-center gap-2 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
            <CheckCircle2 className="h-4 w-4" />
            {success}
          </div>
        )}

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CreditCard className="h-5 w-5 text-[#0757B8]" />
              Branch Payment Settings
            </CardTitle>
          </CardHeader>

          <CardContent className="space-y-6">
            <div className="max-w-xl space-y-2">
              <Label htmlFor="branch">
                Branch
              </Label>

              {loadingBranches ? (
                <div className="flex h-10 items-center gap-2 rounded-md border px-3 text-sm text-slate-500">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading branches...
                </div>
              ) : (
                <select
                  id="branch"
                  value={selectedBranchId}
                  onChange={(event) =>
                    setSelectedBranchId(
                      event.target.value,
                    )
                  }
                  className="flex h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-[#0757B8] focus:ring-1 focus:ring-[#0757B8]"
                >
                  <option value="">
                    Select a branch
                  </option>

                  {branches.map((branch) => (
                    <option
                      key={branch.id}
                      value={branch.id}
                    >
                      {branch.name}
                    </option>
                  ))}
                </select>
              )}

              {selectedBranch && (
                <p className="text-xs text-slate-500">
                  Configure M-Pesa payments specifically
                  for {selectedBranch.name}.
                </p>
              )}
            </div>
          </CardContent>
        </Card>

        {selectedBranchId && (
          <>
            {loadingConfiguration ? (
              <Card>
                <CardContent className="flex min-h-[180px] items-center justify-center">
                  <div className="flex items-center gap-2 text-sm text-slate-500">
                    <Loader2 className="h-5 w-5 animate-spin" />
                    Loading M-Pesa configuration...
                  </div>
                </CardContent>
              </Card>
            ) : (
              <>
                <Card>
                  <CardHeader>
                    <CardTitle>
                      M-Pesa / Daraja
                    </CardTitle>
                  </CardHeader>

                  <CardContent className="space-y-6">
                    <div className="grid gap-5 md:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="provider">
                          Provider
                        </Label>

                        <select
                          id="provider"
                          value={provider}
                          onChange={(event) =>
                            setProvider(
                              event.target.value,
                            )
                          }
                          className="flex h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-[#0757B8] focus:ring-1 focus:ring-[#0757B8]"
                        >
                          <option value="daraja">
                            Safaricom Daraja
                          </option>
                        </select>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="environment">
                          Environment
                        </Label>

                        <select
                          id="environment"
                          value={environment}
                          onChange={(event) =>
                            setEnvironment(
                              event.target
                                .value as
                                | "sandbox"
                                | "production",
                            )
                          }
                          className="flex h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-[#0757B8] focus:ring-1 focus:ring-[#0757B8]"
                        >
                          <option value="sandbox">
                            Sandbox
                          </option>
                          <option value="production">
                            Production
                          </option>
                        </select>
                      </div>

                      <div className="space-y-2 md:col-span-2">
                        <Label htmlFor="business-short-code">
                          Business Shortcode
                        </Label>

                        <Input
                          id="business-short-code"
                          value={businessShortCode}
                          onChange={(event) =>
                            setBusinessShortCode(
                              event.target.value,
                            )
                          }
                          placeholder="Enter business shortcode"
                        />

                        <p className="text-xs text-slate-500">
                          The shortcode used by your
                          M-Pesa Daraja application.
                        </p>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="till-number">
                          Till Number
                        </Label>

                        <Input
                          id="till-number"
                          value={tillNumber}
                          onChange={(event) =>
                            setTillNumber(
                              event.target.value,
                            )
                          }
                          placeholder="Enter Till Number"
                        />
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="paybill-number">
                          Paybill Number
                        </Label>

                        <Input
                          id="paybill-number"
                          value={paybillNumber}
                          onChange={(event) =>
                            setPaybillNumber(
                              event.target.value,
                            )
                          }
                          placeholder="Enter Paybill Number"
                        />
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="account-reference">
                          Account Reference
                        </Label>

                        <Input
                          id="account-reference"
                          value={accountReference}
                          onChange={(event) =>
                            setAccountReference(
                              event.target.value,
                            )
                          }
                          placeholder="e.g. CYBER"
                        />
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="callback-url">
                          Callback URL
                        </Label>

                        <Input
                          id="callback-url"
                          value={callbackUrl}
                          onChange={(event) =>
                            setCallbackUrl(
                              event.target.value,
                            )
                          }
                          placeholder="https://..."
                        />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <ShieldCheck className="h-5 w-5 text-[#0757B8]" />
                      API Credentials
                    </CardTitle>
                  </CardHeader>

                  <CardContent className="space-y-5">
                    <div className="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm text-blue-800">
                      Your M-Pesa credentials are stored
                      securely by the backend. Saved
                      credentials are never displayed
                      here.
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="consumer-key">
                        Consumer Key
                        {configuration?.consumer_key_configured && (
                          <span className="ml-2 text-xs text-green-600">
                            Configured
                          </span>
                        )}
                      </Label>

                      <Input
                        id="consumer-key"
                        type="password"
                        value={consumerKey}
                        onChange={(event) =>
                          setConsumerKey(
                            event.target.value,
                          )
                        }
                        placeholder={
                          configuration?.consumer_key_configured
                            ? "Leave blank to keep existing key"
                            : "Enter consumer key"
                        }
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="consumer-secret">
                        Consumer Secret
                        {configuration?.consumer_secret_configured && (
                          <span className="ml-2 text-xs text-green-600">
                            Configured
                          </span>
                        )}
                      </Label>

                      <div className="relative">
                        <Input
                          id="consumer-secret"
                          type={
                            showConsumerSecret
                              ? "text"
                              : "password"
                          }
                          value={consumerSecret}
                          onChange={(event) =>
                            setConsumerSecret(
                              event.target.value,
                            )
                          }
                          placeholder={
                            configuration?.consumer_secret_configured
                              ? "Leave blank to keep existing secret"
                              : "Enter consumer secret"
                          }
                          className="pr-10"
                        />

                        <button
                          type="button"
                          onClick={() =>
                            setShowConsumerSecret(
                              (value) => !value,
                            )
                          }
                          className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-700"
                          aria-label={
                            showConsumerSecret
                              ? "Hide consumer secret"
                              : "Show consumer secret"
                          }
                        >
                          {showConsumerSecret ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="passkey">
                        Passkey
                        {configuration?.passkey_configured && (
                          <span className="ml-2 text-xs text-green-600">
                            Configured
                          </span>
                        )}
                      </Label>

                      <div className="relative">
                        <Input
                          id="passkey"
                          type={
                            showPasskey
                              ? "text"
                              : "password"
                          }
                          value={passkey}
                          onChange={(event) =>
                            setPasskey(
                              event.target.value,
                            )
                          }
                          placeholder={
                            configuration?.passkey_configured
                              ? "Leave blank to keep existing passkey"
                              : "Enter passkey"
                          }
                          className="pr-10"
                        />

                        <button
                          type="button"
                          onClick={() =>
                            setShowPasskey(
                              (value) => !value,
                            )
                          }
                          className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-700"
                          aria-label={
                            showPasskey
                              ? "Hide passkey"
                              : "Show passkey"
                          }
                        >
                          {showPasskey ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>
                      Configuration Status
                    </CardTitle>
                  </CardHeader>

                  <CardContent>
                    <label className="flex cursor-pointer items-center justify-between gap-4 rounded-lg border border-slate-200 p-4">
                      <div>
                        <p className="font-medium text-slate-900">
                          Enable M-Pesa payments
                        </p>

                        <p className="text-sm text-slate-500">
                          When disabled, this branch
                          will not initiate M-Pesa
                          payments.
                        </p>
                      </div>

                      <input
                        type="checkbox"
                        checked={active}
                        onChange={(event) =>
                          setActive(
                            event.target.checked,
                          )
                        }
                        className="h-5 w-5 rounded border-slate-300"
                      />
                    </label>
                  </CardContent>
                </Card>

                <div className="flex flex-col gap-3 sm:flex-row sm:justify-between">
                  <div>
                    {configuration && (
                      <Button
                        type="button"
                        variant="outline"
                        onClick={handleDelete}
                        disabled={deleting || saving}
                        className="border-red-200 text-red-600 hover:bg-red-50 hover:text-red-700"
                      >
                        {deleting ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="mr-2 h-4 w-4" />
                        )}

                        Delete Configuration
                      </Button>
                    )}
                  </div>

                  <Button
                    type="button"
                    onClick={handleSave}
                    disabled={
                      saving ||
                      deleting ||
                      loadingConfiguration
                    }
                    className="bg-[#0757B8] hover:bg-[#064A9D]"
                  >
                    {saving ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                      <Save className="mr-2 h-4 w-4" />
                    )}

                    Save M-Pesa Settings
                  </Button>
                </div>
              </>
            )}
          </>
        )}

        {!loadingBranches &&
          branches.length === 0 && (
            <Card>
              <CardContent className="py-10 text-center">
                <p className="font-medium text-slate-900">
                  No branches found
                </p>

                <p className="mt-1 text-sm text-slate-500">
                  Create a branch before configuring
                  M-Pesa payments.
                </p>
              </CardContent>
            </Card>
          )}
      </div>
    </DashboardShell>
  );
}
