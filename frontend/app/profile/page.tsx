
"use client";

import { Mail, Phone, Shield, User } from "lucide-react";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";
import { useAuthStore } from "@/store/auth.store";

function formatRole(role?: string) {
  switch (role) {
    case "platform_admin":
      return "SaaS Owner";
    case "owner":
      return "Cyber Owner";
    case "attendant":
      return "Cyber Attendant";
    default:
      return role || "User";
  }
}

export default function ProfilePage() {
  const user = useAuthStore((state) => state.user);

  if (!user) {
    return null;
  }

  const initials = user.full_name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((name) => name[0]?.toUpperCase())
    .join("");

  const phone =
    "phone" in user && typeof user.phone === "string"
      ? user.phone
      : "Not provided";

  return (
    <DashboardShell role={user.role}>
      <PageHeader
        title="My Profile"
        description="View your personal account information."
      />

      <div className="space-y-6">
        <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-100 px-6 py-7 sm:px-8">
            <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
              <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-full bg-blue-50 text-2xl font-bold text-[#0757B8]">
                {initials || "U"}
              </div>

              <div className="min-w-0">
                <h2 className="text-xl font-bold text-slate-950">
                  {user.full_name}
                </h2>

                <p className="mt-1 truncate text-sm text-slate-500">
                  {user.email}
                </p>

                <div className="mt-3">
                  <span className="inline-flex rounded-full bg-blue-50 px-3 py-1 text-xs font-semibold text-[#0757B8]">
                    {formatRole(user.role)}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6 sm:p-8">
            <div className="mb-6">
              <h3 className="text-base font-semibold text-slate-950">
                Personal Information
              </h3>

              <p className="mt-1 text-sm text-slate-500">
                Your account details associated with CyberSaaS.
              </p>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                    <User className="h-4 w-4" />
                  </div>

                  <div className="min-w-0">
                    <p className="text-xs font-medium text-slate-500">
                      Full Name
                    </p>

                    <p className="mt-1 truncate text-sm font-semibold text-slate-900">
                      {user.full_name}
                    </p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                    <Mail className="h-4 w-4" />
                  </div>

                  <div className="min-w-0">
                    <p className="text-xs font-medium text-slate-500">
                      Email Address
                    </p>

                    <p className="mt-1 break-all text-sm font-semibold text-slate-900">
                      {user.email}
                    </p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                    <Phone className="h-4 w-4" />
                  </div>

                  <div className="min-w-0">
                    <p className="text-xs font-medium text-slate-500">
                      Phone Number
                    </p>

                    <p className="mt-1 text-sm font-semibold text-slate-900">
                      {phone}
                    </p>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
                    <Shield className="h-4 w-4" />
                  </div>

                  <div className="min-w-0">
                    <p className="text-xs font-medium text-slate-500">
                      Account Role
                    </p>

                    <p className="mt-1 text-sm font-semibold text-slate-900">
                      {formatRole(user.role)}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </DashboardShell>
  );
}
