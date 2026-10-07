"use client";

import Link from "next/link";
import { Suspense } from "react";

import { ResetPasswordForm } from "@/components/auth/ResetPasswordForm";

export default function ResetPasswordPage() {
  return (
    <main className="relative min-h-screen bg-slate-50">
      <div className="absolute left-0 right-0 top-0">
        <div className="mx-auto flex h-20 max-w-7xl items-center justify-between px-5 lg:px-8">
          <Link
            href="/"
            className="flex items-center gap-2.5"
          >
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-[#0757B8] font-bold text-white">
              C
            </div>

            <span className="text-lg font-bold tracking-tight text-slate-950">
              Cyber
              <span className="text-[#0757B8]">
                SaaS
              </span>
            </span>
          </Link>

          <Link
            href="/login/owner"
            className="text-sm font-medium text-slate-500 transition hover:text-[#0757B8]"
          >
            Back to Login
          </Link>
        </div>
      </div>

      <div className="flex min-h-screen items-center justify-center px-5 py-24">
        <div className="w-full max-w-md space-y-6">
          <div className="text-center">
            <h1 className="text-2xl font-bold tracking-tight text-slate-950">
              Reset your password
            </h1>

            <p className="mt-2 text-sm text-slate-500">
              Enter a new password for your CyberSaaS account.
            </p>
          </div>

          <Suspense
            fallback={
              <div className="text-center text-sm text-slate-500">
                Loading password reset form...
              </div>
            }
          >
            <ResetPasswordForm />
          </Suspense>
        </div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 py-5 text-center text-xs text-slate-400">
        Copyright {new Date().getFullYear()} CyberSaaS. All rights reserved.
      </div>
    </main>
  );
}
