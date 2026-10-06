"use client";

import type { ReactNode } from "react";

import Sidebar from "@/components/dashboard/Sidebar";
import TopNav from "@/components/dashboard/TopNav";

type UserRole = "platform_admin" | "owner" | "attendant";

interface DashboardShellProps {
  role: UserRole;
  children: ReactNode;
}

export default function DashboardShell({
  role,
  children,
}: DashboardShellProps) {
  return (
    <div className="min-h-screen bg-slate-50">
      <div className="flex min-h-screen">
        {/* SIDEBAR */}
        <Sidebar role={role} />

        {/* MAIN AREA */}
        <div className="flex min-w-0 flex-1 flex-col">
          <TopNav />

          <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
            <div className="mx-auto w-full max-w-7xl">
              {children}
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}