export interface Attendant {
  id: string;
  tenant_id: string;
  full_name: string;
  email?: string | null;
  phone?: string | null;
  status: "active" | "inactive" | "suspended";
  branch_id?: string | null;
  branch_name?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateAttendantRequest {
  full_name: string;
  email?: string | null;
  phone?: string | null;
  password: string;
  branch_id: string;
}

export interface UpdateAttendantRequest {
  full_name: string;
  email?: string | null;
  phone?: string | null;
}

export interface ChangeAttendantStatusRequest {
  status: "active" | "inactive" | "suspended";
}

export interface AssignAttendantBranchRequest {
  branch_id: string;
}

export interface AttendantListResponse {
  attendants: Attendant[];
  total: number;
}