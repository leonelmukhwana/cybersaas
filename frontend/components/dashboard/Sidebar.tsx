"use client";

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
  Receipt,
  Settings,
  ShieldCheck,
  ShoppingCart,
  Users,
  Wallet,
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
    {
      label: "Dashboard",
      href: "/saas-owner",
      icon: Gauge,
    },
    {
      label: "Cyber Owners",
      href: "/saas-owner/owners",
      icon: Users,
    },
    {
      label: "Branches",
      href: "//saas-owner/branches",
      icon: Building2,
    },
    {
      label: "Subscriptions",
      href: "/saas-owner/subscriptions",
      icon: ClipboardList,
    },
    {
      label: "Revenue",
      href: "/saas-owner/revenue",
      icon: Wallet,
    },
    {
      label: "Audit Logs",
      href: "/saas-owner/audit-logs",
      icon: ShieldCheck,
    },
  ],

  owner: [
  {
    label: "Dashboard",
    href: "/cyber-owner",
    icon: Gauge,
  },
  {
    label: "Branches",
    href: "/cyber-owner/branch",
    icon: Building2,
  },
  {
    label: "Attendants",
    href: "/cyber-owner/attendant",
    icon: Users,
  },
  {
    label: "Services & Pricing",
    href: "/cyber-owner/services",
    icon: ClipboardList,
  },
  {
    label: "Licence Keys",
    href: "/cyber-owner/licence-keys",
    icon: KeyRound,
  },
  {
    label: "Subscription",
    href: "/cyber-owner/subscription",
    icon: CreditCard,
  },
  {
    label: "Reports",
    href: "/cyber-owner/reports",
    icon: KeyRound,
  },
  {
    label: "Analytics",
    href: "/cyber-owner/analytics",
    icon: Settings,
  },
  {
    label: "Billings",
    href: "/cyber-owner/billing",
    icon: Settings,
  },
  {
    label: "Settings",
    href: "/cyber-owner/settings",
    icon: Settings,
  },
],

  attendant: [
    {
      label: "Dashboard",
      href: "/cyber-attendant",
      icon: Gauge,
    },
    {
      label: "Customers",
      href: "/cyber-attendant/customer",
      icon: Users,
    },
     {
      label: "Computers",
      href: "/cyber-attendant/terminals",
      icon: ClipboardList,
    },
    {
      label: "Sessions",
      href: "/cyber-attendant/sessions",
      icon: ClipboardList,
    },
    {
      label: "Sales",
      href: "/cyber-attendant/sales",
      icon: ShoppingCart,
    },
    {
      label: "Revenue",
      href: "/cyber-attendant/revenue",
      icon: Wallet,
    },
    {
      label: "Payment",
      href: "/cyber-attendant/payment",
      icon: FileText,
    },
    {
      label: "Receipts",
      href: "/cyber-attendant/receipt",
      icon: Receipt,
    },
     {
      label: "Expenses",
      href: "/cyber-attendant/expenses",
      icon: Receipt,
    },
    {
      label: "Reports",
      href: "/cyber-attendant/report/compliance",
      icon: Receipt,
    },
    
    {
      label: "Health",
      href: "/cyber-attendant/health",
      icon: Receipt,
    },
  ],
};

const roleTitles: Record<UserRole, string> = {
  platform_admin: "SaaS Owner",
  owner: "Cyber Owner",
  attendant: "Cyber Attendant",
};

export default function Sidebar({ role }: SidebarProps) {
  const pathname = usePathname();
  const clearAuth = useAuthStore((state) => state.clearAuth);

  const items = sidebarItems[role];

  function handleLogout() {
    clearAuth();
    window.location.href = "/login";
  }

  return (
    <aside className="hidden h-screen w-64 shrink-0 border-r border-slate-200 bg-white lg:flex lg:flex-col">
      {/* BRAND */}
      <div className="flex h-16 items-center border-b border-slate-200 px-6">
        <Link
          href={
            role === "platform_admin"
              ? "/dashboard/saas-owner"
              : role === "owner"
                ? "/dashboard/owner"
                : "/dashboard/attendant"
          }
          className="flex items-center gap-3"
        >
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-[#0757B8] text-sm font-bold text-white">
            C
          </div>

          <div>
            <p className="text-sm font-bold text-slate-950">
              CyberSaaS
            </p>

            <p className="text-xs text-slate-500">
              {roleTitles[role]}
            </p>
          </div>
        </Link>
      </div>

      {/* NAVIGATION */}
      <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-5">
        <p className="mb-3 px-3 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
          Menu
        </p>

        {items.map((item) => {
          const Icon = item.icon;

          const isActive =
            pathname === item.href ||
            (item.href !==
              (role === "platform_admin"
                ? "/dashboard/saas-owner"
                : role === "owner"
                  ? "/dashboard/owner"
                  : "/dashboard/attendant") &&
              pathname.startsWith(`${item.href}/`));

          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition ${
                isActive
                  ? "bg-blue-50 text-[#0757B8]"
                  : "text-slate-600 hover:bg-slate-50 hover:text-slate-950"
              }`}
            >
              <Icon className="h-4.5 w-4.5" />

              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>

      {/* BOTTOM */}
      <div className="border-t border-slate-200 p-3">
        <button
          type="button"
          onClick={handleLogout}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-slate-600 transition hover:bg-red-50 hover:text-red-600"
        >
          <LogOut className="h-4.5 w-4.5" />
          <span>Logout</span>
        </button>
      </div>
    </aside>
  );
}