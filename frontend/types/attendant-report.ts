export interface AttendantReportSummary {
  attendant_id: string;
  period_start: string;
  period_end: string;

  session_revenue: string;
  sales_revenue: string;
  total_revenue: string;
  confirmed_collections: string;

  session_count: number;
  sales_count: number;
  payments_count: number;
}

export interface AttendantPaymentMethodTotal {
  method: string;
  amount: string;
  count: number;
}

export interface AttendantDailyTotal {
  date: string;
  session_revenue: string;
  sales_revenue: string;
  total_revenue: string;
  session_count: number;
  sales_count: number;
}

export interface AttendantReportResponse {
  summary: AttendantReportSummary;
  payment_methods: AttendantPaymentMethodTotal[];
  daily_totals: AttendantDailyTotal[];
}