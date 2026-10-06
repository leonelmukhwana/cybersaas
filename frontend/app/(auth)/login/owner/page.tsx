import { Suspense } from "react";
import Link from "next/link";
import LoginForm from "@/components/auth/LoginForm";

// Prevents Next.js prerender error during build by forcing dynamic rendering
export const dynamic = "force-dynamic";

export default function CyberOwnerLoginPage() {
  return (
    <main className="min-h-screen bg-slate-50">
      <div className="mx-auto flex min-h-screen max-w-7xl flex-col px-5">
        {/* HEADER */}
        <header className="flex h-20 items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-[#0757B8] font-bold text-white">
              C
            </div>

            <span className="text-lg font-bold tracking-tight text-slate-950">
              Cyber<span className="text-[#0757B8]">SaaS</span>
            </span>
          </Link>

          <Link
            href="/"
            className="text-sm font-medium text-slate-500 hover:text-[#0757B8]"
          >
            Back to website
          </Link>
        </header>

        {/* FORM */}
        <div className="flex flex-1 items-center justify-center py-12">
          {/* Suspense boundary handles dynamic search params/hooks inside LoginForm */}
          <Suspense fallback={<div className="text-sm text-slate-500">Loading form...</div>}>
            <LoginForm />
          </Suspense>
        </div>

        {/* FOOTER */}
        <footer className="py-6 text-center text-xs text-slate-400">
          © {new Date().getFullYear()} CyberSaaS. All rights reserved.
        </footer>
      </div>
    </main>
  );
}