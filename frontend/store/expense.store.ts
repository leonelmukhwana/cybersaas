"use client";

import { create } from "zustand";

import { expenseService } from "@/services/expense.service";

import type {
  CreateExpenseRequest,
  Expense,
} from "@/types/expense";

interface ExpenseStore {
  expenses: Expense[];
  total: number;

  loading: boolean;
  saving: boolean;
  error: string | null;

  fetchExpenses: (branchId: string) => Promise<void>;

  createExpense: (
    data: CreateExpenseRequest
  ) => Promise<Expense>;

  clearError: () => void;
}

export const useExpenseStore = create<ExpenseStore>(
  (set) => ({
    expenses: [],
    total: 0,

    loading: false,
    saving: false,
    error: null,

    fetchExpenses: async (branchId: string) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const response =
          await expenseService.listMyBranch(branchId);

        set({
          expenses: response.expenses,
          total: response.total,
          loading: false,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load expenses.",
        });
      }
    },

    createExpense: async (
      data: CreateExpenseRequest
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const expense =
          await expenseService.create(data);

        set((state) => ({
          expenses: [
            expense,
            ...state.expenses,
          ],
          total: state.total + 1,
          saving: false,
        }));

        return expense;
      } catch (error) {
        const message =
          error instanceof Error
            ? error.message
            : "Failed to create expense.";

        set({
          saving: false,
          error: message,
        });

        throw error;
      }
    },

    clearError: () => {
      set({
        error: null,
      });
    },
  })
);