"use client";

import { Bell, Menu } from "lucide-react";

import UserMenu from "@/components/dashboard/UserMenu";

interface TopNavProps {
  onMenuClick?: () => void;
}

export default function TopNav({ onMenuClick }: TopNavProps) {
  return (
    <header className="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-[#064A9C] bg-[#0757B8] px-4 shadow-sm sm:px-6">
      {/* MOBILE MENU */}
      <button
        type="button"
        onClick={onMenuClick}
        className="rounded-lg p-2 text-white transition hover:bg-white/10 lg:hidden"
        aria-label="Open menu"
      >
        <Menu className="h-5 w-5" />
      </button>

      {/* LEFT */}
      <div className="hidden lg:block">
        <p className="text-sm font-medium text-white">
          CyberSaaS Management
        </p>
      </div>

      {/* RIGHT */}
      <div className="ml-auto flex items-center gap-2">
        <button
          type="button"
          className="relative rounded-lg p-2.5 text-white transition hover:bg-white/10"
          aria-label="Notifications"
        >
          <Bell className="h-5 w-5" />

          <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-white ring-2 ring-[#0757B8]" />
        </button>

        <UserMenu />
      </div>
    </header>
  );
}