import { Suspense } from "react";
import Link from "next/link";

export const dynamic = "force-dynamic";

export default function NotFound() {
  return (
    <Suspense fallback={<div className="p-10 text-center text-slate-500">Loading...</div>}>
      <div className="flex min-h-screen flex-col items-center justify-center p-5 text-center font-sans">
        <h2 className="text-2xl font-bold text-slate-900">404 - Page Not Found</h2>
        <p className="mt-2 text-sm text-slate-600">The page you are looking for does not exist.</p>
        <Link
          href="/"
          className="mt-4 rounded-lg bg-[#0757B8] px-4 py-2 text-sm font-medium text-white hover:bg-[#064A9D] transition"
        >
          Return Home
        </Link>
      </div>
    </Suspense>
  );
}