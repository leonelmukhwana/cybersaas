
import { api } from "@/lib/api";

export interface DashboardStats {
  total_cyber_owners: number;
  active_cyber_owners: number;
  suspended_cyber_owners: number;
  total_cyber_locations: number;
  active_cyber_locations: number;
  active_subscriptions: number;
  past_due_subscriptions: number;
  expired_subscriptions: number;
  lifetime_subscriptions: number;
  subscription_revenue_all_time: string;
  subscription_revenue_month: string;
  subscription_revenue_year: string;
  subscription_revenue_three_years: string;
}

export interface CyberOwnerListItem {
  id: string;
  full_name: string;
  email?: string | null;
  phone?: string | null;
  status: string;
  subscription_id?: string | null;
  package_name?: string | null;
  amount: string;
  subscription_status?: string | null;
  subscription_expires_at?: string | null;
  is_lifetime: boolean;
}

export interface CyberOwnerListResponse {
  items: CyberOwnerListItem[];
  total: number;
  limit: number;
  offset: number;
}

export interface SubscriptionListItem {
  id: string;
  owner_id: string;
  owner_name: string;
  owner_email?: string | null;
  owner_phone?: string | null;
  package_name: string;
  amount: string;
  status: string;
  is_lifetime: boolean;
  current_period_start: string;
  current_period_end?: string | null;
  created_at: string;
}

export interface SubscriptionListResponse {
  items: SubscriptionListItem[];
  total: number;
  limit: number;
  offset: number;
}

export interface RevenueListItem {
  id: string;
  owner_id: string;
  owner_name: string;
  owner_email?: string | null;
  amount: string;
  payment_method: string;
  status: string;
  mpesa_receipt_number?: string | null;
  mpesa_checkout_request_id?: string | null;
  created_at: string;
}

export interface RevenueListResponse {
  items: RevenueListItem[];
  total: number;
  limit: number;
  offset: number;
}

export interface AuditLog {
  id: string;
  user_id?: string | null;
  user_name?: string | null;
  user_role?: string | null;
  action: string;
  entity_type?: string | null;
  entity_id?: string | null;
  old_value?: unknown;
  new_value?: unknown;
  reason?: string | null;
  ip_address?: string | null;
  device_id?: string | null;
  created_at: string;
}

export interface AuditLogResponse {
  items: AuditLog[];
  total: number;
  limit: number;
  offset: number;
}

function paginationParams(
  search: string,
  limit: number,
  offset: number
) {
  const params = new URLSearchParams();

  if (search.trim()) {
    params.set("search", search.trim());
  }

  params.set("limit", String(limit));
  params.set("offset", String(offset));

  return params.toString();
}

export const platformService = {
  async dashboard(token: string) {
    return api<DashboardStats>("/api/platform/dashboard", {
      method: "GET",
      token,
    });
  },

  async owners(
    token: string,
    search = "",
    limit = 20,
    offset = 0
  ) {
    return api<CyberOwnerListResponse>(
      `/api/platform/owners?${paginationParams(
        search,
        limit,
        offset
      )}`,
      {
        method: "GET",
        token,
      }
    );
  },

  async subscriptions(
    token: string,
    search = "",
    limit = 20,
    offset = 0
  ) {
    return api<SubscriptionListResponse>(
      `/api/platform/subscriptions?${paginationParams(
        search,
        limit,
        offset
      )}`,
      {
        method: "GET",
        token,
      }
    );
  },

  async revenue(
    token: string,
    search = "",
    limit = 20,
    offset = 0
  ) {
    return api<RevenueListResponse>(
      `/api/platform/revenue?${paginationParams(
        search,
        limit,
        offset
      )}`,
      {
        method: "GET",
        token,
      }
    );
  },

  async auditLogs(
    token: string,
    search = "",
    limit = 20,
    offset = 0
  ) {
    return api<AuditLogResponse>(
      `/api/platform/audit-logs?${paginationParams(
        search,
        limit,
        offset
      )}`,
      {
        method: "GET",
        token,
      }
    );
  },
};
