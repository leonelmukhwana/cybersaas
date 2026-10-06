import {
  LayoutDashboard,
  Building2,
  Monitor,
  Users,
  CreditCard,
  Package,
  Settings,
  Receipt,
  Activity,
} from "lucide-react";

export type UserRole =
  | "saas_owner"
  | "cyber_owner"
  | "cyber_attendant";

export interface NavigationItem {
  title: string;
  href: string;
  icon: React.ElementType;
}

export const navigationByRole: Record<UserRole, NavigationItem[]> = {
  saas_owner: [
    {
      title: "Dashboard",
      href: "/saas-owner/dashboard",
      icon: LayoutDashboard,
    },
    {
      title: "Cyber Owners",
      href: "/saas-owner/cyber-owners",
      icon: Users,
    },
    {
      title: "Packages",
      href: "/saas-owner/packages",
      icon: Package,
    },
    {
      title: "Subscriptions",
      href: "/saas-owner/subscriptions",
      icon: CreditCard,
    },
    {
      title: "Activity",
      href: "/saas-owner/activity",
      icon: Activity,
    },
    {
      title: "Settings",
      href: "/saas-owner/settings",
      icon: Settings,
    },
  ],

  cyber_owner: [
    {
      title: "Dashboard",
      href: "/owner/dashboard",
      icon: LayoutDashboard,
    },
    {
      title: "Branches",
      href: "/owner/branches",
      icon: Building2,
    },
    {
      title: "Terminals",
      href: "/owner/terminals",
      icon: Monitor,
    },
    {
      title: "Attendants",
      href: "/owner/attendants",
      icon: Users,
    },
    {
      title: "Subscriptions",
      href: "/owner/subscription",
      icon: CreditCard,
    },
    {
      title: "Settings",
      href: "/owner/settings",
      icon: Settings,
    },
  ],

  cyber_attendant: [
    {
      title: "Dashboard",
      href: "/attendant/dashboard",
      icon: LayoutDashboard,
    },
    {
      title: "Terminals",
      href: "/attendant/terminals",
      icon: Monitor,
    },
    {
      title: "Sessions",
      href: "/attendant/sessions",
      icon: Activity,
    },
    {
      title: "Payments",
      href: "/attendant/payments",
      icon: CreditCard,
    },
    {
      title: "Receipts",
      href: "/attendant/receipts",
      icon: Receipt,
    },
  ],
};