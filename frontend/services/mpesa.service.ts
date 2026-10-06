
import { api } from "@/lib/api";

export interface MpesaConfiguration {
  id: string;
  tenant_id?: string;
  branch_id?: string;
  provider: string;
  environment: "sandbox" | "production";
  business_short_code?: string | null;
  till_number?: string | null;
  paybill_number?: string | null;
  account_reference?: string | null;
  callback_url?: string | null;
  active: boolean;
  consumer_key_configured: boolean;
  consumer_secret_configured: boolean;
  passkey_configured: boolean;
  created_at: string;
  updated_at: string;
}

export interface SaveMpesaConfigurationRequest {
  branch_id?: string;
  provider: string;
  environment: "sandbox" | "production";
  business_short_code?: string;
  till_number?: string;
  paybill_number?: string;

  // Only send these when the owner enters new values.
  consumer_key?: string;
  consumer_secret?: string;
  passkey?: string;

  account_reference?: string;
  callback_url?: string;
  active?: boolean;
}



export interface InitiateSTKRequest {
  branch_id: string;
  sale_id: string;
  payment_id?: string | null;
  phone_number: string;
  amount: string;
  account_reference: string;
  transaction_description: string;
}

export interface STKRequestResponse {
  id: string;
  tenant_id: string;
  branch_id: string;
  sale_id: string;
  payment_id?: string | null;
  phone_number: string;
  amount: string;
  account_reference: string;
  transaction_description: string;
  merchant_request_id?: string | null;
  checkout_request_id?: string | null;
  response_code?: string | null;
  response_description?: string | null;
  customer_message?: string | null;
  status: string;
  result_code?: string | null;
  result_description?: string | null;
  mpesa_receipt_number?: string | null;
  transaction_date?: string | null;
  callback_received_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface InitiateSTKResponse {
  message: string;
  stk_request: STKRequestResponse;
}

export interface InitiateSTKRequest {
  branch_id: string;
  sale_id: string;
  payment_id?: string | null;
  phone_number: string;
  amount: string;
  account_reference: string;
  transaction_description: string;
}

export interface InitiateSTKResponse {
  message: string;
  stk_request: {
    id: string;
    tenant_id: string;
    branch_id: string;
    sale_id: string;
    payment_id?: string | null;
    phone_number: string;
    amount: string;
    account_reference: string;
    transaction_description: string;
    merchant_request_id?: string | null;
    checkout_request_id?: string | null;
    response_code?: string | null;
    response_description?: string | null;
    customer_message?: string | null;
    status: string;
    result_code?: string | null;
    result_description?: string | null;
    mpesa_receipt_number?: string | null;
    transaction_date?: string | null;
    callback_received_at?: string | null;
    created_at: string;
    updated_at: string;
  };
}

export const mpesaService = {
  async getBranch(
    branchId: string,
  ): Promise<MpesaConfiguration> {
    return api<MpesaConfiguration>(
      `/api/mpesa/branches/${branchId}/config`,
      {
        method: "GET",
      },
    );
  },

  async saveBranch(
    branchId: string,
    data: SaveMpesaConfigurationRequest,
  ): Promise<MpesaConfiguration> {
    return api<MpesaConfiguration>(
      `/api/mpesa/branches/${branchId}/config`,
      {
        method: "PUT",
        body: JSON.stringify({
          ...data,
          branch_id: branchId,
        }),
      },
    );
  },

  async deleteBranch(
    branchId: string,
  ): Promise<void> {
    await api<void>(
      `/api/mpesa/branches/${branchId}/config`,
      {
        method: "DELETE",
      },
    );
  },


  async initiateSTK(
  data: InitiateSTKRequest,
): Promise<InitiateSTKResponse> {
  return api<InitiateSTKResponse>(
    "/api/mpesa/branches/stk",
    {
      method: "POST",
      body: JSON.stringify(data),
    },
  );
},
};
