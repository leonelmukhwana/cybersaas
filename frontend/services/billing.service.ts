
import type {
  BillingConfig,
  BillingConfigInput,
  BillingConfigResponse,
} from "@/types/billing";

const BACKEND_URL = (
  process.env.NEXT_PUBLIC_API_URL ??
  "https://cybersaas.onrender.com"
).replace(/\/+$/, "");

const API_BASE_URL = BACKEND_URL.endsWith("/api")
  ? BACKEND_URL
  : `${BACKEND_URL}/api`;

async function request<T>(
  token: string,
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  const response = await fetch(`${API_BASE_URL}${normalizedPath}`, {
    ...options,
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
      ...options.headers,
    },
    cache: "no-store",
  });

  const result = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error(
      result.error ??
        result.message ??
        `Request failed with status ${response.status}.`,
    );
  }

  return result as T;
}

export const billingService = {
  async get(
    token: string,
    branchId: string,
  ): Promise<BillingConfig> {
    const result = await request<BillingConfigResponse>(
      token,
      `/branches/${encodeURIComponent(branchId)}/billing-config`,
    );

    return result.data;
  },

  async save(
    token: string,
    branchId: string,
    input: BillingConfigInput,
  ): Promise<BillingConfig> {
    const result = await request<BillingConfigResponse>(
      token,
      `/branches/${encodeURIComponent(branchId)}/billing-config`,
      {
        method: "PUT",
        body: JSON.stringify(input),
      },
    );

    return result.data;
  },
};
