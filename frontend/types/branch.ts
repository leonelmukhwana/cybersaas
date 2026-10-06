export type BranchStatus = "active" | "inactive";

export interface Branch {
  id: string;
  tenant_id: string;
  name: string;
  address?: string | null;
  phone?: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBranchRequest {
  business_name: string;
  name: string;
  address?: string;
  phone?: string;
}

export interface UpdateBranchRequest {
  name: string;
  address?: string;
  phone?: string;
}

export interface ChangeBranchStatusRequest {
  status: string;
}

export interface BranchListResponse {
  branches: Branch[];
  total: number;
}