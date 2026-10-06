import type { ReactNode } from "react";
import { Suspense } from "react";

// Forces all routes inside the (dashboard) group to render dynamically during request time
export const dynamic = "force-dynamic";

export default function DashboardLayout({
  children,
}: {
  children: ReactNode;
}) {
  return (
    <Suspense fallback={<div className="flex min-h-screen items-center justify-center text-sm text-slate-500">Loading dashboard...</div>}>
      {children}
    </Suspense>
  );
}