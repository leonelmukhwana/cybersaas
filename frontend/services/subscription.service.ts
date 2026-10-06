import { api } from "@/lib/api";
import {
  Subscription,
  SubscriptionOverview,
  SubscriptionPayment,
  SubscriptionPlan,
} from "@/types/subscription";

export interface UpdateSubscriptionPlanRequest {
  name: string;
  included_branches: number;
  included_terminals: number;
  extra_branch_rate: string;
  extra_terminal_rate: string;
  monthly_price: string;
  is_lifetime: boolean;
  is_active: boolean;
}

export interface CreateSubscriptionPlanRequest {
  name: string;
  included_branches: number;
  included_terminals: number;
  extra_branch_rate: string;
  extra_terminal_rate: string;
  monthly_price: string;
  is_lifetime: boolean;
  is_active: boolean;
}

interface PlansResponse {
  plans: SubscriptionPlan[];
}

interface PaymentsResponse {
  payments: SubscriptionPayment[];
  limit?: number;
  offset?: number;
}

export const subscriptionService = {
  // ============================================================
  // PUBLIC / CYBER OWNER
  // ============================================================

  /**
   * Get all active subscription plans.
   *
   * Backend:
   * GET /api/subscription/plans
   *
   * Used by:
   * - Landing page
   * - Cyber Owner package selection
   * - Other public subscription screens
   */
  async getPlans(token?: string): Promise<SubscriptionPlan[]> {
    const response = await api<PlansResponse>(
      "/api/subscription/plans",
      {
        method: "GET",
        ...(token ? { token } : {}),
      }
    );

    if (
      response &&
      typeof response === "object" &&
      Array.isArray(response.plans)
    ) {
      return response.plans;
    }

    console.error(
      "Invalid subscription plans response:",
      response
    );

    return [];
  },

  /**
   * Get the current tenant subscription overview.
   *
   * Backend:
   * GET /api/subscription
   */
  async getOverview(
    token: string
  ): Promise<SubscriptionOverview> {
    return api<SubscriptionOverview>(
      "/api/subscription",
      {
        method: "GET",
        token,
      }
    );
  },

  /**
   * Select/create the tenant's subscription plan.
   *
   * Backend:
   * POST /api/subscription
   */
  async createSubscription(
    planId: string,
    token: string
  ): Promise<Subscription> {
    return api<Subscription>(
      "/api/subscription",
      {
        method: "POST",
        token,
        body: JSON.stringify({
          plan_id: planId,
        }),
      }
    );
  },

  /**
   * Get subscription payment history.
   *
   * Backend:
   * GET /api/subscription/payments
   */
  async getPayments(
    token: string
  ): Promise<SubscriptionPayment[]> {
    const response = await api<PaymentsResponse>(
      "/api/subscription/payments",
      {
        method: "GET",
        token,
      }
    );

    if (
      response &&
      typeof response === "object" &&
      Array.isArray(response.payments)
    ) {
      return response.payments;
    }

    console.error(
      "Invalid subscription payments response:",
      response
    );

    return [];
  },

  /**
   * Initiate a subscription payment.
   *
   * The backend determines the amount from
   * the tenant's selected plan.
   *
   * Backend:
   * POST /api/subscription/payments
   */
  async createPayment(
    phoneNumber: string,
    token: string,
    paymentMethod = "mpesa_stk"
  ): Promise<SubscriptionPayment> {
    return api<SubscriptionPayment>(
      "/api/subscription/payments",
      {
        method: "POST",
        token,
        body: JSON.stringify({
          phone_number: phoneNumber,
          payment_method: paymentMethod,
        }),
      }
    );
  },

  /**
   * Get subscription ledger.
   *
   * Backend:
   * GET /api/subscription/ledger
   */
  async getLedger(token: string) {
    return api(
      "/api/subscription/ledger",
      {
        method: "GET",
        token,
      }
    );
  },

  // ============================================================
  // PLATFORM ADMIN / SAAS OWNER
  // ============================================================

  /**
   * Get ALL subscription plans.
   *
   * Unlike getPlans(), this includes inactive plans.
   *
   * Backend:
   * GET /api/platform/subscription-plans
   *
   * Used only by the SaaS Owner / Platform Admin Settings page.
   */
  async getPlatformPlans(
    token: string
  ): Promise<SubscriptionPlan[]> {
    const response = await api<PlansResponse>(
      "/api/platform/subscription-plans",
      {
        method: "GET",
        token,
      }
    );

    if (
      response &&
      typeof response === "object" &&
      Array.isArray(response.plans)
    ) {
      return response.plans;
    }

    console.error(
      "Invalid platform subscription plans response:",
      response
    );

    return [];
  },

  /**
   * Create a new subscription plan.
   *
   * Backend:
   * POST /api/platform/subscription-plans
   *
   * The backend:
   * - validates the plan
   * - creates the plan in subscription_plans
   * - records the creation in audit_logs
   */
  async createPlatformPlan(
    data: CreateSubscriptionPlanRequest,
    token: string
  ): Promise<SubscriptionPlan> {
    return api<SubscriptionPlan>(
      "/api/platform/subscription-plans",
      {
        method: "POST",
        token,
        body: JSON.stringify(data),
      }
    );
  },

  /**
   * Update an existing subscription plan.
   *
   * Backend:
   * PATCH /api/platform/subscription-plans/:id
   *
   * The backend:
   * - validates the plan
   * - updates subscription_plans
   * - records the change in audit_logs
   */
  async updatePlatformPlan(
    planId: string,
    data: UpdateSubscriptionPlanRequest,
    token: string
  ): Promise<SubscriptionPlan> {
    return api<SubscriptionPlan>(
      `/api/platform/subscription-plans/${planId}`,
      {
        method: "PATCH",
        token,
        body: JSON.stringify(data),
      }
    );
  },
};