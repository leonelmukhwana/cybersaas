"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { authService } from "@/services/auth.service";
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

export default function RegisterForm() {
  const router = useRouter();

  const setAuth = useAuthStore((state) => state.setAuth);

  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setError("");

    const trimmedFullName = fullName.trim();
    const trimmedEmail = email.trim();
    const trimmedPhone = phone.trim();

    if (!trimmedFullName) {
      setError("Full name is required.");
      return;
    }

    if (!trimmedEmail) {
      setError("Email address is required.");
      return;
    }

    if (!trimmedPhone) {
      setError("Phone number is required.");
      return;
    }

    if (!password) {
      setError("Password is required.");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setLoading(true);

    try {
      const response = await authService.register({
        full_name: trimmedFullName,
        email: trimmedEmail,
        phone: trimmedPhone,
        password,
      });

      console.log("REGISTER RESPONSE:", response);

      // Save authenticated user + JWT
      setAuth(response.user, response.token);

      // Owner registration does not create the business yet.
      // The backend response currently has no tenant_id,
      // so send the owner to business setup.
      if (response.user.role === "owner") {
        if (!response.user.tenant_id) {
          router.replace("/business");
          return;
        }

        router.replace("/dashboard/owner");
        return;
      }

      // These are kept for completeness if the backend
      // ever returns another role during registration.
      if (response.user.role === "attendant") {
        router.replace("/dashboard/attendant");
        return;
      }

      if (response.user.role === "platform_admin") {
        router.replace("/dashboard/saas-owner");
        return;
      }

      setError("Your account role is not supported.");
    } catch (err) {
      console.error("REGISTER ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Registration failed. Please try again."
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="w-full max-w-md border-slate-200 shadow-xl shadow-slate-900/5">
      <CardHeader className="space-y-3 text-center">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#0757B8] text-xl font-bold text-white">
          C
        </div>

        <div>
          <CardTitle className="text-2xl font-bold text-slate-950">
            Create your account
          </CardTitle>

          <CardDescription className="mt-2 text-slate-500">
            Create your CyberSaaS Cyber Owner account
          </CardDescription>
        </div>
      </CardHeader>

      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* FULL NAME */}
          <div className="space-y-2">
            <Label htmlFor="full-name">Full name</Label>

            <Input
              id="full-name"
              type="text"
              placeholder="e.g. Leon Ambume"
              value={fullName}
              onChange={(event) => setFullName(event.target.value)}
              autoComplete="name"
              required
              disabled={loading}
              className="h-11"
            />
          </div>

          {/* EMAIL */}
          <div className="space-y-2">
            <Label htmlFor="email">Email address</Label>

            <Input
              id="email"
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

          {/* PHONE */}
          <div className="space-y-2">
            <Label htmlFor="phone">Phone number</Label>

            <Input
              id="phone"
              type="tel"
              placeholder="0712345678"
              value={phone}
              onChange={(event) => setPhone(event.target.value)}
              autoComplete="tel"
              required
              disabled={loading}
              className="h-11"
            />
          </div>

          {/* PASSWORD */}
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>

            <div className="relative">
              <Input
                id="password"
                type={showPassword ? "text" : "password"}
                placeholder="Create a password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                autoComplete="new-password"
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

          {/* CONFIRM PASSWORD */}
          <div className="space-y-2">
            <Label htmlFor="confirm-password">Confirm password</Label>

            <div className="relative">
              <Input
                id="confirm-password"
                type={showConfirmPassword ? "text" : "password"}
                placeholder="Confirm your password"
                value={confirmPassword}
                onChange={(event) =>
                  setConfirmPassword(event.target.value)
                }
                autoComplete="new-password"
                required
                disabled={loading}
                className="h-11 pr-20"
              />

              <button
                type="button"
                onClick={() =>
                  setShowConfirmPassword((value) => !value)
                }
                disabled={loading}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-semibold text-slate-500 hover:text-[#0757B8]"
              >
                {showConfirmPassword ? "Hide" : "Show"}
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

          {/* SUBMIT */}
          <Button
            type="submit"
            disabled={loading}
            className="mt-2 h-11 w-full bg-[#0757B8] text-sm font-semibold hover:bg-[#064A9D]"
          >
            {loading ? "Creating account..." : "Create Account"}
          </Button>

          {/* LOGIN */}
          <p className="pt-1 text-center text-sm text-slate-500">
            Already have an account?{" "}
            <Link
              href="/login/owner"
              className="font-semibold text-[#0757B8] hover:underline"
            >
              Login
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  );
}