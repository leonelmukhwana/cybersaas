
import { api } from "@/lib/api";

import type {
  CreatePaymentInput,
  PaymentResponse,
  PaymentStatus,
} from "@/types/sale";

export interface PaymentListResponse {
  payments: PaymentResponse[];
  total: number;
}

export interface PaymentQueryParams {
  branch_id: string;
  status?: PaymentStatus;
  limit?: number;
  offset?: number;
}

export interface ConfirmPaymentInput {
  external_reference?: string | null;
  mpesa_receipt_number?: string | null;
  provider_request_id?: string | null;
  provider_transaction_id?: string | null;
}

export interface FailPaymentInput {
  reason: string;
}

export interface RefundPaymentInput {
  reason: string;
}

export const paymentService = {
  async create(
    data: CreatePaymentInput,
  ): Promise<PaymentResponse> {
    return api<PaymentResponse>("/api/payments", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async list(
    params: PaymentQueryParams,
  ): Promise<PaymentListResponse> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      params.branch_id,
    );

    if (params.status) {
      searchParams.set(
        "status",
        params.status,
      );
    }

    searchParams.set(
      "limit",
      String(params.limit ?? 50),
    );

    searchParams.set(
      "offset",
      String(params.offset ?? 0),
    );

    return api<PaymentListResponse>(
      `/api/payments?${searchParams.toString()}`,
      {
        method: "GET",
      },
    );
  },

  async get(
    paymentId: string,
    branchId: string,
  ): Promise<PaymentResponse> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      branchId,
    );

    return api<PaymentResponse>(
      `/api/payments/${paymentId}?${searchParams.toString()}`,
      {
        method: "GET",
      },
    );
  },

  async confirm(
    paymentId: string,
    branchId: string,
    data: ConfirmPaymentInput = {},
  ): Promise<PaymentResponse> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      branchId,
    );

    return api<PaymentResponse>(
      `/api/payments/${paymentId}/confirm?${searchParams.toString()}`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      },
    );
  },

  async fail(
    paymentId: string,
    branchId: string,
    data: FailPaymentInput,
  ): Promise<{ message: string }> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      branchId,
    );

    return api<{ message: string }>(
      `/api/payments/${paymentId}/fail?${searchParams.toString()}`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      },
    );
  },

  async refund(
    paymentId: string,
    branchId: string,
    data: RefundPaymentInput,
  ): Promise<{ message: string }> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      branchId,
    );

    return api<{ message: string }>(
      `/api/payments/${paymentId}/refund?${searchParams.toString()}`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      },
    );
  },
};
