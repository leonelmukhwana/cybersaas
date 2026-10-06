export interface Tenant {
  id: string;
  business_name: string;
  owner_name: string;
  phone: string;
  email?: string;
  status: string;
  created_at: string;
}

export interface CreateTenantRequest {
  business_name: string;
}