export interface SubscriptionPlan {
  id: string;
  name: string;
  included_branches: number;
  included_terminals: number;
  extra_branch_rate: string;
  extra_terminal_rate: string;
  monthly_price: string;
  is_lifetime: boolean;
  is_active: boolean;
  created_at: string;
}

export interface Subscription {
  id: string;
  tenant_id: string;
  plan_id?: string | null;
  plan_name?: string;

  status: string;
  is_trial: boolean;
  trial_ends_at?: string | null;
  is_lifetime: boolean;

  account_balance: string;

  current_period_start: string;
  current_period_end?: string | null;

  selected_branches_limit: number;
  selected_terminals_limit: number;

  peak_branches: number;
  peak_terminals: number;

  created_at: string;
  updated_at: string;
}

export interface SubscriptionPayment {
  id: string;
  subscription_id: string;
  tenant_id: string;

  amount: string;
  phone_number: string;

  payment_method: string;

  mpesa_checkout_request_id?: string | null;
  mpesa_receipt_number?: string | null;

  status: string;
  failure_reason?: string | null;

  created_at: string;
  updated_at: string;
}

export interface SubscriptionOverview {
  subscription: Subscription;
  plan?: SubscriptionPlan | null;

  branches_used: number;
  branches_limit: number;

  terminals_used: number;
  terminals_limit: number;

  is_active: boolean;
  is_expired: boolean;

  is_trial: boolean;
  is_trial_expired: boolean;

  must_select_package: boolean;

  suggested_monthly_amount?: string;

  is_lifetime: boolean;
}