
export interface OwnerReportSummary {
  tenant_id: string;
  branch_id?: string | null;

  period_start: string;
  period_end: string;

  session_revenue: string;
  sales_revenue: string;
  total_revenue: string;

  // Kept for compatibility with existing report consumers.
  total_sales: string;
  total_collections: string;
  total_expenses: string;
  net_amount: string;

  session_count: number;
  sales_count: number;
  payments_count: number;
  expenses_count: number;
}

export interface OwnerPaymentMethodTotal {
  method: string;
  amount: string;
  count: number;
}

export interface OwnerBranchTotal {
  branch_id: string;
  branch_name: string;

  session_revenue: string;
  sales_revenue: string;
  total_revenue: string;

  // Compatibility fields.
  total_sales: string;
  total_collections: string;
  total_expenses: string;
  net_amount: string;

  session_count: number;
  sales_count: number;
}

export interface OwnerDailyTotal {
  date: string;

  session_revenue: string;
  sales_revenue: string;
  total_revenue: string;

  total_sales: string;
  total_collections: string;
  total_expenses: string;
  net_amount: string;

  session_count: number;
  sales_count: number;
}

export interface OwnerMonthlyTotal {
  month: string;

  session_revenue: string;
  sales_revenue: string;
  total_revenue: string;

  total_sales: string;
  total_collections: string;
  total_expenses: string;
  net_amount: string;

  session_count: number;
  sales_count: number;
}

export interface OwnerSalesReport {
  sales: OwnerSaleRow[];
  total: number;
}

export interface OwnerSaleRow {
  id: string;
  branch_id: string;
  customer_id?: string | null;
  session_id?: string | null;
  terminal_id?: string | null;
  attendant_id?: string | null;

  status: string;
  total_amount: string;
  currency: string;
  created_at: string;

  items?: OwnerSaleItem[];
}

export interface OwnerSaleItem {
  id: string;
  service_id?: string | null;
  description: string;
  quantity: string;
  unit_price: string;
  line_total: string;
}

export interface OwnerPaymentReport {
  payments: OwnerPaymentRow[];
  total: number;
}

export interface OwnerPaymentRow {
  id: string;
  branch_id: string;
  sale_id: string;
  method: string;
  status: string;
  amount: string;
  external_reference?: string | null;
  mpesa_receipt_number?: string | null;
  confirmed_at?: string | null;
  created_at: string;
}

export interface OwnerExpenseReport {
  expenses: OwnerExpenseRow[];
  total: number;
}

export interface OwnerExpenseRow {
  id: string;
  branch_id: string;
  recorded_by: string;
  category: string;
  description?: string | null;
  amount: string;
  currency: string;
  payment_method: string;
  reference?: string | null;
  expense_date: string;
  status: string;
  created_at: string;
}

export interface OwnerReportSummaryResponse {
  summary: OwnerReportSummary;
  payment_methods: OwnerPaymentMethodTotal[];
  branches: OwnerBranchTotal[];
  daily_totals?: OwnerDailyTotal[];
  monthly_totals?: OwnerMonthlyTotal[];
}

export interface FetchOwnerReportFilters {
  branch_id?: string;
  period_start?: string;
  period_end?: string;
}
