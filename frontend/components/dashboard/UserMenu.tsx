"use client";

import { LogOut, Settings, User } from "lucide-react";
import { useRouter } from "next/navigation";

import { useAuthStore } from "@/store/auth.store";

export default function UserMenu() {
  const router = useRouter();

  const user = useAuthStore((state) => state.user);
  const clearAuth = useAuthStore((state) => state.clearAuth);

  function handleLogout() {
    clearAuth();
    router.replace("/login");
  }

  if (!user) {
    return null;
  }

  const initials = user.full_name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((name) => name[0]?.toUpperCase())
    .join("");

  return (
    <div className="relative">
      <details className="group">
        <summary className="flex cursor-pointer list-none items-center gap-3 rounded-lg px-2 py-1.5 hover:bg-slate-50">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-blue-50 text-sm font-semibold text-[#0757B8]">
            {initials || "U"}
          </div>

          <div className="hidden text-left sm:block">
            <p className="max-w-32 truncate text-sm font-semibold text-slate-900">
              {user.full_name}
            </p>

            <p className="text-xs capitalize text-slate-500">
              {user.role === "platform_admin"
                ? "SaaS Owner"
                : user.role}
            </p>
          </div>

          <svg
            className="hidden h-4 w-4 text-slate-400 transition-transform group-open:rotate-180 sm:block"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.51a.75.75 0 01-1.08 0l-4.25-4.51a.75.75 0 01.02-1.06z"
              clipRule="evenodd"
            />
          </svg>
        </summary>

        <div className="absolute right-0 top-12 z-50 w-56 rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl shadow-slate-900/10">
          <div className="border-b border-slate-100 px-3 py-3">
            <p className="truncate text-sm font-semibold text-slate-900">
              {user.full_name}
            </p>

            <p className="truncate text-xs text-slate-500">
              {user.email}
            </p>
          </div>

          <div className="py-1">
            <button
              type="button"
              onClick={() => router.push("/profile")}
              className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-slate-600 hover:bg-slate-50 hover:text-slate-950"
            >
              <User className="h-4 w-4" />
              Profile
            </button>

            <button
              type="button"
              onClick={() => router.push("/settings")}
              className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-slate-600 hover:bg-slate-50 hover:text-slate-950"
            >
              <Settings className="h-4 w-4" />
              Settings
            </button>
          </div>

          <div className="border-t border-slate-100 pt-1">
            <button
              type="button"
              onClick={handleLogout}
              className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-red-600 hover:bg-red-50"
            >
              <LogOut className="h-4 w-4" />
              Logout
            </button>
          </div>
        </div>
      </details>
    </div>
  );
}