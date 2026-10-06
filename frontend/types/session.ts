
export type SessionType =
  | "walk_in"
  | "prepaid";

export type SessionStatus =
  | "active"
  | "paused"
  | "completed"
  | "cancelled";

export interface Session {
  id: string;

  tenant_id: string;
  branch_id: string;
  terminal_id: string;
  customer_id: string;
  attendant_id?: string | null;

  /*
   * Safe display fields.
   *
   * These are returned by the backend for normal
   * attendant session views.
   *
   * Sensitive customer ID numbers are NOT included.
   */
  customer_name: string;
  customer_type: string;
  parent_name?: string | null;
  terminal_name?: string | null;

  session_type: SessionType;
  status: SessionStatus;

  started_at: string;
  ended_at?: string | null;
  paused_at?: string | null;

  total_paused_seconds: number;

  prepaid_amount?: string | null;
  allowed_minutes?: number | null;

  rate_per_minute?: string | null;
  minimum_charge?: string | null;
  final_amount?: string | null;

  client_operation_id: string;

  created_at: string;
  updated_at: string;
}

/* =========================
   Start Session
========================= */

export interface StartSessionRequest {
  branch_id: string;
  terminal_id: string;
  customer_id: string;
  session_type: SessionType;

  prepaid_amount?: string | null;
  allowed_minutes?: number | null;

  started_at?: string | null;

  client_operation_id: string;
}

/* =========================
   End Session
========================= */

export interface EndSessionRequest {
  ended_at?: string | null;
}

/* =========================
   Cancel Session
========================= */

export interface CancelSessionRequest {
  reason: string;
}

/* =========================
   Session List
========================= */

export interface SessionListResponse {
  sessions: Session[];
  total: number;
}

/* =========================
   Billing Preview
========================= */

export interface BillingPreview {
  session_id: string;

  elapsed_minutes: number;
  billable_minutes: number;

  rate_per_minute: string;
  minimum_charge: string;

  calculated_amount: string;
  currency: string;
}

/* =========================
   Session With Billing
========================= */

export interface SessionWithBilling
  extends Session {
  elapsed_minutes: number;
  billable_minutes: number;
  calculated_amount: string;
  currency: string;
}

/* =========================
   Terminal Session
========================= */

export interface TerminalStartSessionRequest {
  branch_id: string;
  terminal_id: string;
  customer_id: string;
  session_type: SessionType;

  prepaid_amount?: string | null;
  allowed_minutes?: number | null;

  started_at?: string | null;

  client_operation_id: string;
}

export interface TerminalSessionPauseResponse {
  message: string;
}

export interface TerminalSessionResumeResponse {
  message: string;
}

export interface TerminalSessionEndResponse {
  message: string;
}

/* =========================
   Terminal Billing Config
========================= */

export interface TerminalBillingConfig {
  rate_per_minute: string;
  minimum_charge: string;
  billing_interval_seconds?: number;
  rounding_mode?: string;
  currency: string;
}
