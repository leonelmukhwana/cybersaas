export type ExpensePaymentMethod =
  | "cash"
  | "mpesa"
  | "bank"
  | "card"
  | "other";

export type ExpenseStatus =
  | "recorded"
  | "voided";

export interface Expense {
  id: string;

  tenant_id: string;
  branch_id: string;
  recorded_by: string;

  category: string;
  description?: string | null;

  amount: string;
  currency: string;

  payment_method: ExpensePaymentMethod;

  reference?: string | null;

  expense_date: string;

  status: ExpenseStatus;

  created_at: string;
  updated_at: string;
}

export interface CreateExpenseRequest {
  branch_id: string;
  category: string;
  description?: string | null;
  amount: string;
  currency?: string;
  payment_method: ExpensePaymentMethod;
  reference?: string | null;
  expense_date: string;
}

export interface ExpenseListResponse {
  expenses: Expense[];
  total: number;
}