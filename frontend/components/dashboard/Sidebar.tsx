
"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BarChart3,
  Building2,
  ClipboardList,
  CreditCard,
  FileText,
  Gauge,
  KeyRound,
  Laptop,
  LogOut,
  Menu,
  Receipt,
  Settings,
  ShieldCheck,
  ShoppingCart,
  Users,
  Wallet,
  X,
} from "lucide-react";

import { useAuthStore } from "@/store/auth.store";

type UserRole = "platform_admin" | "owner" | "attendant";

interface SidebarItem {
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface SidebarProps {
  role: UserRole;
}

const sidebarItems: Record<UserRole, SidebarItem[]> = {
  platform_admin: [
    { label: "Dashboard", href: "/saas-owner", icon: Gauge },
    { label: "Cyber Owners", href: "/saas-owner/owners", icon: Users },
    { label: "Branches", href: "/saas-owner/branches", icon: Building2 },
    { label: "Subscriptions", href: "/saas-owner/subscriptions", icon: ClipboardList },
    { label: "Revenue", href: "/saas-owner/revenue", icon: Wallet },
    { label: "Audit Logs", href: "/saas-owner/audit-logs", icon: ShieldCheck },
  ],

  owner: [
    { label: "Dashboard", href: "/cyber-owner", icon: Gauge },
    { label: "Branches", href: "/cyber-owner/branch", icon: Building2 },
    { label: "Attendants", href: "/cyber-owner/attendant", icon: Users },
    { label: "Services & Pricing", href: "/cyber-owner/services", icon: ClipboardList },
    { label: "Licence Keys", href: "/cyber-owner/licence-keys", icon: KeyRound },
    { label: "Subscription", href: "/cyber-owner/subscription", icon: CreditCard },
    { label: "Reports", href: "/cyber-owner/reports", icon: FileText },
    { label: "Analytics", href: "/cyber-owner/analytics", icon: BarChart3 },
    { label: "Billings", href: "/cyber-owner/billing", icon: Wallet },
    { label: "Settings", href: "/cyber-owner/settings", icon: Settings },
  ],

  attendant: [
    { label: "Dashboard", href: "/cyber-attendant", icon: Gauge },
    { label: "Customers", href: "/cyber-attendant/customer", icon: Users },
    { label: "Computers", href: "/cyber-attendant/terminals", icon: Laptop },
    { label: "Sessions", href: "/cyber-attendant/sessions", icon: ClipboardList },
    { label: "Sales", href: "/cyber-attendant/sales", icon: ShoppingCart },
    { label: "Revenue", href: "/cyber-attendant/revenue", icon: Wallet },
    { label: "Payment", href: "/cyber-attendant/payment", icon: CreditCard },
    { label: "Receipts", href: "/cyber-attendant/receipt", icon: Receipt },
    { label: "Expenses", href: "/cyber-attendant/expenses", icon: Wallet },
    { label: "Reports", href: "/cyber-attendant/report/compliance", icon: FileText },
    { label: "Health", href: "/cyber-attendant/health", icon: ShieldCheck },
  ],
};

const roleTitles: Record<UserRole, string> = {
  platform_admin: "SaaS Owner",
  owner: "Cyber Owner",
  attendant: "Cyber Attendant",
};

const dashboardPaths: Record<UserRole, string> = {
  platform_admin: "/saas-owner",
  owner: "/cyber-owner",
  attendant: "/cyber-attendant",
};

export default function Sidebar({ role }: SidebarProps) {
  const pathname = usePathname();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const [mobileOpen, setMobileOpen] = useState(false);

  const items = sidebarItems[role];

  function handleLogout() {
    setMobileOpen(false);
    clearAuth();
    window.location.href = "/login";
  }

  function isItemActive(href: string) {
    return pathname === href || pathname.startsWith(`${href}/`);
  }

  function renderNavigation(mobile = false) {
    return (
      <>
        <p className="mb-3 px-3 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
          Menu
        </p>

        {items.map((item) => {
          const Icon = item.icon;
          const active = isItemActive(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => {
                if (mobile) setMobileOpen(false);
              }}
              className={`flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition ${
                active
                  ? "bg-blue-50 text-[#0757B8]"
                  : "text-slate-600 hover:bg-slate-50 hover:text-slate-950"
              }`}
            >
              <Icon className="h-5 w-5 shrink-0" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </>
    );
  }

  function renderBrand() {
    return (
      <Link
        href={dashboardPaths[role]}
        onClick={() => setMobileOpen(false)}
        className="flex items-center gap-3"
      >
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#0757B8] text-sm font-bold text-white">
          C
        </div>

        <div>
          <p className="text-sm font-bold text-slate-950">CyberSaaS</p>
          <p className="text-xs text-slate-500">{roleTitles[role]}</p>
        </div>
      </Link>
    );
  }

  function renderLogout() {
    return (
      <button
        type="button"
        onClick={handleLogout}
        className="flex w-full items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium text-slate-600 transition hover:bg-red-50 hover:text-red-600"
      >
        <LogOut className="h-5 w-5 shrink-0" />
        <span>Logout</span>
      </button>
    );
  }

  return (
    <>
      {/* Mobile menu button */}
      <button
        type="button"
        aria-label={mobileOpen ? "Close navigation menu" : "Open navigation menu"}
        aria-expanded={mobileOpen}
        aria-controls="mobile-dashboard-navigation"
        onClick={() => setMobileOpen((open) => !open)}
        className="fixed left-4 top-4 z-[60] flex h-11 w-11 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-700 shadow-md transition hover:bg-slate-50 lg:hidden"
      >
        {mobileOpen ? (
          <X className="h-6 w-6" />
        ) : (
          <Menu className="h-6 w-6" />
        )}
      </button>

      {/* Mobile backdrop */}
      {mobileOpen && (
        <button
          type="button"
          aria-label="Close navigation menu"
          onClick={() => setMobileOpen(false)}
          className="fixed inset-0 z-40 bg-slate-950/50 lg:hidden"
        />
      )}

      {/* Mobile navigation drawer */}
      <aside
        id="mobile-dashboard-navigation"
        aria-label="Mobile dashboard navigation"
        aria-hidden={!mobileOpen}
        className={`fixed inset-y-0 left-0 z-50 flex w-[min(18rem,85vw)] flex-col border-r border-slate-200 bg-white shadow-xl transition-transform duration-300 ease-in-out lg:hidden ${
          mobileOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex h-16 shrink-0 items-center justify-between border-b border-slate-200 px-5">
          {renderBrand()}

          <button
            type="button"
            aria-label="Close navigation menu"
            onClick={() => setMobileOpen(false)}
            className="rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-900"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-5">
          {renderNavigation(true)}
        </nav>

        <div className="shrink-0 border-t border-slate-200 p-3">
          {renderLogout()}
        </div>
      </aside>

      {/* Desktop sidebar */}
      <aside className="hidden h-screen w-64 shrink-0 flex-col border-r border-slate-200 bg-white lg:flex">
        <div className="flex h-16 shrink-0 items-center border-b border-slate-200 px-6">
          {renderBrand()}
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-5">
          {renderNavigation()}
        </nav>

        <div className="shrink-0 border-t border-slate-200 p-3">
          {renderLogout()}
        </div>
      </aside>
    </>
  );
}