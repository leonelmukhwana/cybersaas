"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { tenantService } from "@/services/tenant.service";
import { useAuthStore } from "@/store/auth.store";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function BusinessPage() {
  const router = useRouter();

  const { user, token, updateUser } = useAuthStore();

  const [businessName, setBusinessName] = useState("");
  const [loading, setLoading] = useState(false);
  const [checking, setChecking] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token || !user) {
      router.replace("/login/owner");
      return;
    }

    // This page is only for Cyber Owners.
    if (user.role !== "owner") {
      router.replace("/login/owner");
      return;
    }

    // Business already exists.
    if (user.tenant_id) {
      router.replace("/cyber-owner");
      return;
    }

    setChecking(false);
  }, [token, user, router]);

  async function handleSubmit(
    event: FormEvent<HTMLFormElement>
  ) {
    event.preventDefault();

    if (!token || !user) {
      router.replace("/login/owner");
      return;
    }

    const name = businessName.trim();

    if (!name) {
      setError("Please enter your business name.");
      return;
    }

    if (name.length < 2) {
      setError(
        "Business name must contain at least 2 characters."
      );
      return;
    }

    if (name.length > 200) {
      setError("Business name is too long.");
      return;
    }

    setLoading(true);
    setError("");

    try {
      const response = await tenantService.create(
        {
          business_name: name,
        },
        token
      );

      // The backend has attached the tenant to the owner.
      updateUser({
        ...user,
        tenant_id: response.tenant.id,
      });

      router.replace("/dashboard/owner");
    } catch (err) {
      console.error("CREATE BUSINESS ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Failed to create your business."
      );
    } finally {
      setLoading(false);
    }
  }

  if (checking) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-white">
        <div className="text-sm text-slate-500">
          Loading...
        </div>
      </main>
    );
  }

  if (!user || user.role !== "owner") {
    return null;
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-white px-4 py-10">
      <Card className="w-full max-w-xl border-slate-200 shadow-lg">
        <CardHeader className="space-y-3 text-center">
          <CardTitle className="text-2xl font-bold text-slate-900">
            Set up your business
          </CardTitle>

          <CardDescription className="text-base text-slate-500">
            Enter your business name to continue.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form
            onSubmit={handleSubmit}
            className="space-y-6"
          >
            <div className="space-y-2">
              <Label htmlFor="business_name">
                Business name
              </Label>

              <Input
                id="business_name"
                name="business_name"
                placeholder="e.g. Leon Cyber Café"
                value={businessName}
                onChange={(event) =>
                  setBusinessName(event.target.value)
                }
                disabled={loading}
                maxLength={200}
                className="h-14 rounded-xl border-slate-300 px-4 text-base focus-visible:ring-blue-600"
                autoFocus
              />

              <p className="text-sm text-slate-500">
                Enter the name customers will know your business by.
              </p>
            </div>

            {error && (
              <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
                {error}
              </div>
            )}

            <Button
              type="submit"
              disabled={loading || !businessName.trim()}
              className="h-14 w-full rounded-xl bg-blue-600 text-base font-semibold text-white hover:bg-blue-700"
            >
              {loading
                ? "Creating business..."
                : "Create Business"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}