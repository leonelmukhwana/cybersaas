
import { create } from "zustand";

import type {
  Customer,
  CreateCustomerRequest,
  UpdateCustomerRequest,
} from "@/types/customer";

import { customerService } from "@/services/customer.service";

interface CustomerStore {
  customers: Customer[];
  total: number;

  loading: boolean;
  saving: boolean;

  error: string | null;

  selectedCustomer: Customer | null;

  fetchCustomers: (
    token: string,
    branchId: string,
    search?: string,
  ) => Promise<void>;

  createCustomer: (
    token: string,
    data: CreateCustomerRequest,
  ) => Promise<Customer>;

  updateCustomer: (
    token: string,
    customerId: string,
    branchId: string,
    data: UpdateCustomerRequest,
  ) => Promise<Customer>;

  selectCustomer: (
    customer: Customer | null,
  ) => void;

  clearError: () => void;
}

export const useCustomerStore =
  create<CustomerStore>((set) => ({
    customers: [],
    total: 0,

    loading: false,
    saving: false,

    error: null,

    selectedCustomer: null,

    fetchCustomers: async (
      token,
      branchId,
      search = "",
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const response =
          await customerService.list(
            token,
            {
              branch_id: branchId,
              search:
                search.trim() || undefined,
              limit: 100,
              offset: 0,
            },
          );

        set({
          customers:
            response.customers ?? [],
          total: response.total ?? 0,
          loading: false,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load customers.",
        });

        throw error;
      }
    },

    createCustomer: async (
      token,
      data,
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const request: CreateCustomerRequest = {
          ...data,
          id: data.id || crypto.randomUUID(),
        };

        console.log(
          "[Customer] Creating customer:",
          request,
        );

        const customer =
          await customerService.create(
            token,
            request,
          );

        set((state) => ({
          customers: [
            customer,
            ...state.customers,
          ],
          total: state.total + 1,
          saving: false,
        }));

        return customer;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to register customer.",
        });

        throw error;
      }
    },

    updateCustomer: async (
      token,
      customerId,
      branchId,
      data,
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const updated =
          await customerService.update(
            token,
            customerId,
            branchId,
            data,
          );

        set((state) => ({
          customers:
            state.customers.map(
              (customer) =>
                customer.id === customerId
                  ? updated
                  : customer,
            ),
          selectedCustomer: null,
          saving: false,
        }));

        return updated;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to update customer.",
        });

        throw error;
      }
    },

    selectCustomer: (
      customer,
    ) => {
      set({
        selectedCustomer:
          customer,
      });
    },

    clearError: () => {
      set({
        error: null,
      });
    },
  }));
