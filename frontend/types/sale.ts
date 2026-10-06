export type SaleStatus = "completed" | "voided" | "refunded";

export type PaymentMethod = "cash" | "mpesa" | "other";

export type PaymentStatus =
  | "pending"
  | "confirmed"
  | "failed"
  | "refunded";

export interface SaleItemInput {
  service_id?: string | null;
  description: string;
  quantity: number;
  unit_price: string;
}

export interface CreateSaleInput {
  branch_id: string;
  customer_id?: string | null;
  session_id?: string | null;
  session_started_at?: string | null;
  terminal_id?: string | null;
  attendant_id?: string | null;
  discount_type?: string | null;
  discount_value: string;
  items: SaleItemInput[];
}

export interface SaleResponse {
  id: string;
  tenant_id: string;
  branch_id: string;
  customer_id?: string | null;
  session_id?: string | null;
  session_started_at?: string | null;
  terminal_id?: string | null;
  attendant_id?: string | null;

  status: SaleStatus;

  subtotal: string;
  discount_type?: string | null;
  discount_value: string;
  discount_amount: string;
  total_amount: string;
  currency: "KES";

  created_at: string;
  updated_at: string;

  items: SaleItemResponse[];
}

export interface SaleItemResponse {
  id: string;
  sale_id: string;
  service_id?: string | null;
  description: string;
  quantity: number;
  unit_price: string;
  line_total: string;
}


export interface CreatePaymentInput {
  branch_id: string;
  sale_id: string;
  method: PaymentMethod;
  amount: string;
  phone?: string | null;
  external_reference?: string | null;
  client_operation_id?: string | null;
}


export interface PaymentResponse {
  id: string;
  tenant_id: string;
  branch_id: string;
  sale_id: string;
  method: PaymentMethod;
  status: PaymentStatus;
  amount: string;
  phone?: string | null;
  external_reference?: string | null;
  mpesa_receipt_number?: string | null;
  provider_request_id?: string | null;
  provider_transaction_id?: string | null;
  failure_reason?: string | null;
  confirmed_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ReceiptResponse {
  id: string;
  tenant_id: string;
  branch_id: string;
  sale_id: string;
  payment_id?: string | null;
  receipt_number: string;
  receipt_type: "sale";
  issued_at: string;
  printed_at?: string | null;
  reprint_count: number;
  created_at: string;
  updated_at: string;
}

export interface CompletedOfflineSale {
  sale: SaleResponse;
  payment: PaymentResponse;
  receipt: ReceiptResponse;
}