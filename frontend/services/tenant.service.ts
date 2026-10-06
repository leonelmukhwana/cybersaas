import { api } from "@/lib/api";

export interface Tenant {
  id: string;
  business_name: string;
  owner_name: string;
  phone: string;
  email?: string | null;
  status: string;
  created_at: string;
}

export interface CreateTenantRequest {
  business_name: string;
}

export interface CreateTenantResponse {
  message: string;
  tenant: Tenant;
}

export interface GetTenantResponse {
  tenant: Tenant;
}

export const tenantService = {
  async create(
    data: CreateTenantRequest,
    token: string
  ): Promise<CreateTenantResponse> {
    return api<CreateTenantResponse>("/api/tenant", {
      method: "POST",
      token,
      body: JSON.stringify(data),
    });
  },

  async me(token: string): Promise<GetTenantResponse> {
    return api<GetTenantResponse>("/api/tenant/me", {
      method: "GET",
      token,
    });
  },
};