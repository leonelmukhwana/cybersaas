"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { authService } from "@/services/auth.service";
import { attendantService } from "@/services/attendant.service";
import {
  createOfflineCredential,
  loginOffline,
} from "@/services/offline-auth.service";

import { useAuthStore } from "@/store/auth.store";

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

type LoginRole = "owner" | "attendant" | "platform_admin";

interface LoginFormProps {
  role?: LoginRole;
}

export default function LoginForm({ role = "owner" }: LoginFormProps) {
  const router = useRouter();

  const setAuth = useAuthStore((state) => state.setAuth);
  const setOfflineAuth = useAuthStore((state) => state.setOfflineAuth);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const isSaaSOwner = role === "platform_admin";
  const isAttendant = role === "attendant";

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    setError("");
    setLoading(true);

    try {
      /*
       * ATTENDANT OFFLINE LOGIN
       *
       * Offline login is only available to
       * attendants who have previously completed
       * a successful online login on this
       * workstation.
       */
      if (isAttendant && !navigator.onLine) {
        const offlineResult = await loginOffline(email.trim(), password);

        setOfflineAuth({
          user: offlineResult.user,
          branchId: offlineResult.branchId,
          branchName: offlineResult.branchName,
        });

        router.replace("/cyber-attendant");
        return;
      }

      /*
       * NORMAL ONLINE LOGIN
       */
      const response = await authService.login({
        email: email.trim(),
        password,
      });

      /*
       * Make sure the account matches
       * the login page.
       */
      if (response.user.role !== role) {
        setError(
          isSaaSOwner
            ? "This account is not a SaaS Owner account."
            : isAttendant
              ? "This account is not a Cyber Attendant account."
              : "This account is not a Cyber Owner account."
        );
        return;
      }

      /*
       * ATTENDANT ONLINE SETUP
       *
       * The login response contains a fresh JWT.
       * Pass that token directly to the attendant
       * request because setAuth() has not happened
       * yet.
       */
      if (isAttendant) {
        if (!response.user.tenant_id) {
          setError("Your attendant account is missing a tenant ID.");
          return;
        }

        const attendant = await attendantService.me(response.token);

        if (!attendant.branch_id) {
          setError("Your attendant account is not assigned to a branch.");
          return;
        }

        if (attendant.status !== "active") {
          setError("Your attendant account is not active.");
          return;
        }

        await createOfflineCredential({
          user: response.user,
          tenantId: response.user.tenant_id,
          branchId: attendant.branch_id,
          branchName: attendant.branch_name || "Assigned Branch",
          password,
          credentialVersion: 1,
        });
      }

      /*
       * Persist the online session only after
       * all validation and attendant setup
       * succeeds.
       */
      setAuth(response.user, response.token);

      /*
       * Route according to role.
       */
      if (role === "owner") {
        if (!response.user.tenant_id) {
          router.replace("/business");
          return;
        }

        router.replace("/cyber-owner");
        return;
      }

      if (role === "attendant") {
        router.replace("/cyber-attendant");
        return;
      }

      if (role === "platform_admin") {
        router.replace("/saas-owner");
        return;
      }

      setError("Your account role is not supported.");
    } catch (err) {
      console.error("LOGIN ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Something went wrong. Please try again."
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="w-full max-w-md shadow-md">
      <CardHeader>
        <div>
          <CardTitle className="text-2xl font-bold text-slate-950">
            Welcome back
          </CardTitle>

          <CardDescription className="mt-2 text-slate-500">
            {isSaaSOwner
              ? "Login to your CyberSaaS management account"
              : isAttendant
                ? "Login to your CyberSaaS attendant account"
                : "Login to your CyberSaaS owner account"}
          </CardDescription>
        </div>
      </CardHeader>

      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-5">
          {/* EMAIL */}
          <div className="space-y-2">
            <Label htmlFor="email">Email address</Label>

            <Input
              id="email"
              name="email"
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              autoComplete="email"
              required
              disabled={loading}
              className="h-11"
            />
          </div>

          {/* PASSWORD */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="password">Password</Label>

              <Link
                href={
                  isSaaSOwner
                    ? "/forgot-password/saas-owner"
                    : isAttendant
                      ? "/forgot-password/attendant"
                      : "/forgot-password/owner"
                }
                className="text-sm font-medium text-[#0757B8] hover:underline"
              >
                Forgot password?
              </Link>
            </div>

            <div className="relative">
              <Input
                id="password"
                name="password"
                type={showPassword ? "text" : "password"}
                placeholder="Enter your password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                autoComplete="current-password"
                required
                disabled={loading}
                className="h-11 pr-20"
              />

              <button
                type="button"
                onClick={() => setShowPassword((value) => !value)}
                disabled={loading}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-semibold text-slate-500 hover:text-[#0757B8]"
              >
                {showPassword ? "Hide" : "Show"}
              </button>
            </div>
          </div>

          {/* ERROR */}
          {error && (
            <div
              role="alert"
              className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600"
            >
              {error}
            </div>
          )}

          {/* LOGIN BUTTON */}
          <Button
            type="submit"
            disabled={loading}
            className="h-11 w-full bg-[#0757B8] text-sm font-semibold hover:bg-[#064A9D]"
          >
            {loading
              ? "Logging in..."
              : isSaaSOwner
                ? "Login as SaaS Owner"
                : isAttendant
                  ? "Login as Attendant"
                  : "Login"}
          </Button>

          {/* REGISTER - OWNER ONLY */}
          {role === "owner" && (
            <p className="text-center text-sm text-slate-500">
              Don't have an account?{" "}
              <Link
                href="/register/owner"
                className="font-semibold text-[#0757B8] hover:underline"
              >
                Create an account
              </Link>
            </p>
          )}
        </form>
      </CardContent>
    </Card>
  );
}