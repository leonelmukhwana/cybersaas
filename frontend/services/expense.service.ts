import { api } from "@/lib/api";

import type {
  CreateExpenseRequest,
  Expense,
  ExpenseListResponse,
} from "@/types/expense";

export const expenseService = {
  async create(
    data: CreateExpenseRequest
  ): Promise<Expense> {
    return api<Expense>("/api/expenses", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async listMyBranch(
    branchId: string
  ): Promise<ExpenseListResponse> {
    return api<ExpenseListResponse>(
      `/api/expenses/my-branch?branch_id=${encodeURIComponent(branchId)}`,
      {
        method: "GET",
      }
    );
  },

  async get(
    expenseId: string
  ): Promise<Expense> {
    return api<Expense>(
      `/api/expenses/${encodeURIComponent(expenseId)}`,
      {
        method: "GET",
      }
    );
  },
};