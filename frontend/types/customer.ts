export type CustomerType = "adult" | "child";

export interface Customer {
  id: string;
  tenant_id: string;
  branch_id: string;

  customer_type: CustomerType;

  full_name: string;
  phone?: string | null;
  id_number?: string | null;

  parent_name?: string | null;
  parent_phone?: string | null;
  parent_id_number?: string | null;

  created_at: string;
  updated_at: string;
}

export interface CreateCustomerRequest {
  id: string;
  branch_id: string;
  customer_type: CustomerType;

  full_name: string;
  phone?: string | null;
  id_number?: string | null;

  parent_name?: string | null;
  parent_phone?: string | null;
  parent_id_number?: string | null;
}

export interface UpdateCustomerRequest {
  customer_type: CustomerType;

  full_name: string;
  phone?: string | null;
  id_number?: string | null;

  parent_name?: string | null;
  parent_phone?: string | null;
  parent_id_number?: string | null;
}

export interface CustomerListResponse {
  customers: Customer[];
  total: number;
}

export interface CustomerQueryParams {
  branch_id: string;
  search?: string;
  limit?: number;
  offset?: number;
}