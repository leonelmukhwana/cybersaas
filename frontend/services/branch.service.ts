import { api } from "@/lib/api";

import type {
  Branch,
  BranchListResponse,
  ChangeBranchStatusRequest,
  CreateBranchRequest,
  UpdateBranchRequest,
} from "@/types/branch";

export const branchService = {
  async list(token: string) {
    return api<BranchListResponse>("/api/branches", {
      method: "GET",
      token,
    });
  },

  async get(id: string, token: string) {
    return api<Branch>(`/api/branches/${id}`, {
      method: "GET",
      token,
    });
  },

  async create(
    data: CreateBranchRequest,
    token: string
  ) {
    return api<Branch>("/api/branches", {
      method: "POST",
      token,
      body: JSON.stringify(data),
    });
  },

  async update(
    id: string,
    data: UpdateBranchRequest,
    token: string
  ) {
    return api<Branch>(`/api/branches/${id}`, {
      method: "PUT",
      token,
      body: JSON.stringify(data),
    });
  },

  async changeStatus(
    id: string,
    data: ChangeBranchStatusRequest,
    token: string
  ) {
    return api<Branch>(`/api/branches/${id}/status`, {
      method: "PATCH",
      token,
      body: JSON.stringify(data),
    });
  },
};