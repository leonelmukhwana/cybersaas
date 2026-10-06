import type {
  CreatePaymentInput,
  CreateSaleInput,
  PaymentResponse,
  ReceiptResponse,
  SaleResponse,
} from "@/types/sale";

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function getToken(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  return localStorage.getItem("access_token");
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken();

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token
        ? {
            Authorization: `Bearer ${token}`,
          }
        : {}),
      ...(options.headers ?? {}),
    },
  });

  if (!response.ok) {
    const text = await response.text();

    throw new Error(
      text || `Request failed: ${response.status}`,
    );
  }

  return response.json();
}

export async function createSaleOnline(
  input: CreateSaleInput,
  clientOperationId: string,
): Promise<SaleResponse> {
  return apiFetch<SaleResponse>("/sales", {
    method: "POST",
    body: JSON.stringify({
      ...input,
      client_operation_id: clientOperationId,
    }),
  });
}

export async function createPaymentOnline(
  input: CreatePaymentInput,
  clientOperationId: string,
): Promise<PaymentResponse> {
  return apiFetch<PaymentResponse>("/payments", {
    method: "POST",
    body: JSON.stringify({
      ...input,
      client_operation_id: clientOperationId,
    }),
  });
}

export async function confirmPaymentOnline(
  paymentId: string,
): Promise<PaymentResponse> {
  return apiFetch<PaymentResponse>(
    `/payments/${paymentId}/confirm`,
    {
      method: "PATCH",
    },
  );
}

export async function createReceiptOnline(
  saleId: string,
  paymentId: string,
): Promise<ReceiptResponse> {
  return apiFetch<ReceiptResponse>("/receipts", {
    method: "POST",
    body: JSON.stringify({
      sale_id: saleId,
      payment_id: paymentId,
    }),
  });
}

export async function printReceiptOnline(
  receiptId: string,
): Promise<ReceiptResponse> {
  return apiFetch<ReceiptResponse>(
    `/receipts/${receiptId}/print`,
    {
      method: "POST",
    },
  );
}