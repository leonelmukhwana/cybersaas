import type { UserRole } from "@/types/auth";

export interface ComplianceBranch {
  id: string;
  name: string;
  address?: string | null;
  phone?: string | null;
}

export interface ComplianceCustomer {
  id: string;
  customer_type: "adult" | "child";
  full_name: string;
  phone?: string | null;

  id_number?: string | null;

  parent_name?: string | null;
  parent_phone?: string | null;
  parent_id_number?: string | null;

  registered_at: string;
  updated_at: string;
}

export interface ComplianceTerminal {
  id: string;
  terminal_code: string;
  machine_name?: string | null;
  device_identifier?: string | null;
  status: string;
  registered_at?: string | null;
  last_seen_at?: string | null;
  disabled_at?: string | null;
  created_at: string;
}

export interface ComplianceSession {
  id: string;

  branch_id: string;

  terminal_id: string;
  terminal_code?: string | null;
  terminal_name?: string | null;

  customer_id: string;
  customer_name?: string | null;
  customer_type?: "adult" | "child" | null;

  attendant_id?: string | null;
  attendant_name?: string | null;

  session_type: "pay_after" | "prepaid";
  status: "active" | "completed" | "expired" | "cancelled";

  started_at: string;
  ended_at?: string | null;

  allowed_minutes?: number | null;
  rate_per_minute?: number | null;
  minimum_charge?: number | null;
  prepaid_amount?: number | null;
  final_amount?: number | null;

  used_minutes?: number | null;
  remaining_minutes?: number | null;
}

export interface ComplianceSale {
  id: string;

  customer_id?: string | null;
  session_id?: string | null;
  terminal_id?: string | null;
  attendant_id?: string | null;

  customer_name?: string | null;
  terminal_name?: string | null;
  attendant_name?: string | null;

  subtotal: number;
  discount_type?: "fixed" | "percentage" | null;
  discount_value: number;
  discount_amount: number;
  total_amount: number;

  status: "completed" | "voided" | "refunded";

  void_reason?: string | null;
  voided_by?: string | null;
  voided_at?: string | null;

  created_at: string;
}

export interface ComplianceSaleItem {
  id: string;
  sale_id: string;

  service_id?: string | null;
  description: string;

  quantity: number;
  unit_price: number;
  line_total: number;

  created_at: string;
}

export interface CompliancePayment {
  id: string;
  sale_id: string;

  method: "cash" | "mpesa" | "other";
  status:
    | "pending"
    | "confirmed"
    | "failed"
    | "cancelled"
    | "refunded";

  amount: number;

  phone?: string | null;
  external_reference?: string | null;
  mpesa_receipt_number?: string | null;

  provider_request_id?: string | null;
  provider_transaction_id?: string | null;

  failure_reason?: string | null;

  confirmed_at?: string | null;
  created_at: string;
}

export interface ComplianceReceipt {
  id: string;

  sale_id: string;
  payment_id?: string | null;

  receipt_number: string;
  receipt_type: string;

  issued_at: string;
  printed_at?: string | null;

  reprint_count: number;
}

export interface ComplianceReceiptReprint {
  id: string;
  receipt_id: string;

  user_id?: string | null;
  terminal_id?: string | null;

  reason?: string | null;

  created_at: string;
}

export interface ComplianceAuditEvent {
  id: string;

  user_id?: string | null;
  user_role?: UserRole | null;

  action: string;

  entity_type?: string | null;
  entity_id?: string | null;

  old_value?: Record<string, unknown> | null;
  new_value?: Record<string, unknown> | null;

  reason?: string | null;

  ip_address?: string | null;
  device_id?: string | null;

  created_at: string;
}

export interface ComplianceLimitations {
  pause_resume?: string;
  extensions?: string;
}

export interface ComplianceRecords {
  customers: ComplianceCustomer[];
  terminals: ComplianceTerminal[];
  sessions: ComplianceSession[];
  sales: ComplianceSale[];
  sale_items: ComplianceSaleItem[];
  payments: CompliancePayment[];
  receipts: ComplianceReceipt[];
  receipt_reprints?: ComplianceReceiptReprint[];
  audit_events: ComplianceAuditEvent[];
}

export interface ComplianceSnapshotData {
  report_type: "cak_compliance";

  tenant_id: string;
  branch_id: string;

  period: {
    start: string;
    end: string;
  };

  generated_at: string;

  retention: {
    years: number;
  };

  branch: ComplianceBranch;

  records: ComplianceRecords;

  limitations?: ComplianceLimitations;

  excluded?: string[];
}

export interface ComplianceReportSnapshot {
  id: string;
  tenant_id: string;
  branch_id?: string | null;
  generated_by?: string | null;

  report_type: string;

  period_start: string;
  period_end: string;

  generated_at: string;
  retention_until: string;

  snapshot: ComplianceSnapshotData;

  snapshot_hash?: string | null;

  created_at: string;
}

export interface GenerateComplianceReportRequest {
  branch_id: string;
  report_type: "cak_compliance";
  period_start: string;
  period_end: string;
}

export interface ComplianceReportResponse {
  snapshot: ComplianceReportSnapshot;
}

export interface ComplianceReportListResponse {
  snapshots: ComplianceReportSnapshot[];
}