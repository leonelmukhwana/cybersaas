
import { api } from "@/lib/api";

import type {
  Attendant,
  AttendantListResponse,
  CreateAttendantRequest,
  UpdateAttendantRequest,
  ChangeAttendantStatusRequest,
  AssignAttendantBranchRequest,
} from "@/types/attendant";

const BASE_PATH = "/api/attendants";

export const attendantService = {
  async list(): Promise<AttendantListResponse> {
    return api<AttendantListResponse>(BASE_PATH, {
      method: "GET",
    });
  },

  async me(token?: string): Promise<Attendant> {
    return api<Attendant>(
      `${BASE_PATH}/me`,
      {
        method: "GET",
        ...(token ? { token } : {}),
      }
    );
  },

  async get(id: string): Promise<Attendant> {
    return api<Attendant>(
      `${BASE_PATH}/${id}`,
      {
        method: "GET",
      }
    );
  },

  async create(
    data: CreateAttendantRequest
  ): Promise<Attendant> {
    return api<Attendant>(BASE_PATH, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async update(
    id: string,
    data: UpdateAttendantRequest
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/${id}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      }
    );
  },

  async changeStatus(
    id: string,
    data: ChangeAttendantStatusRequest
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/${id}/status`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async assignBranch(
    id: string,
    data: AssignAttendantBranchRequest
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/${id}/branch`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async unassignBranch(
    id: string
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/${id}/branch`,
      {
        method: "DELETE",
      }
    );
  },
};
