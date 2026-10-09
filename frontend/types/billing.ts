export type BillingRoundingMode = "up" | "down" | "nearest";

export interface BillingConfig {
  id?: string;
  tenant_id?: string;
  branch_id: string;
  rate_per_minute: string;
  minimum_charge: string;
  billing_interval_minutes: number;
  rounding_mode: BillingRoundingMode;
  currency: string;
  created_at?: string;
  updated_at?: string;
}

export interface BillingConfigInput {
  rate_per_minute: string;
  minimum_charge: string;
  billing_interval_minutes: number;
  rounding_mode: BillingRoundingMode;
  currency: string;
}

export interface BillingConfigResponse {
  data: BillingConfig;
}
