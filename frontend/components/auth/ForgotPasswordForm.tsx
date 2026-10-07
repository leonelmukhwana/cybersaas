
"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { authService } from "@/services/auth.service";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function ForgotPasswordForm() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setError("");
    setSuccess(false);
    setLoading(true);

    try {
      await authService.forgotPassword({
        email: email.trim(),
      });

      setSuccess(true);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to process your request. Please try again.",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full max-w-md">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-bold tracking-tight text-slate-950">
          Forgot your password?
        </h1>

        <p className="mt-2 text-sm text-slate-500">
          Enter your email address and we&apos;ll send you a password
          reset link.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        {success ? (
          <div className="space-y-5">
            <div className="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
              If an account exists with that email, a password reset
              link has been sent.
            </div>

            <div className="text-center">
              <Link
                href="/login"
                className="text-sm font-semibold text-[#0757B8] hover:underline"
              >
                Back to login
              </Link>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div
                role="alert"
                className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600"
              >
                {error}
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="email">
                Email address
              </Label>

              <Input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(event) =>
                  setEmail(event.target.value)
                }
                autoComplete="email"
                required
                disabled={loading}
                className="h-11"
              />
            </div>

            <Button
              type="submit"
              disabled={loading}
              className="h-11 w-full bg-[#0757B8] text-sm font-semibold hover:bg-[#064A9D]"
            >
              {loading
                ? "Sending reset link..."
                : "Send reset link"}
            </Button>

            <div className="text-center">
              <Link
                href="/login"
                className="text-sm font-medium text-slate-500 hover:text-[#0757B8] hover:underline"
              >
                Back to login
              </Link>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}

