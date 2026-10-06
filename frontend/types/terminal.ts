export type TerminalStatus =
  | "pending"
  | "active"
  | "disabled";

export type TerminalLockState =
  | "locked"
  | "unlocked";

export interface Terminal {
  id: string;
  tenant_id: string;
  branch_id: string;
  branch_name?: string | null;
  terminal_code: string;
  machine_name?: string | null;
  device_identifier?: string | null;
  status: TerminalStatus;
  registered_at?: string | null;
  last_seen_at?: string | null;
  disabled_at?: string | null;
  created_at: string;
  updated_at: string;
}

/* =========================
   Terminal Registration
========================= */

export interface RegisterTerminalRequest {
  licence_key: string;
  machine_name: string;
  device_identifier: string;
}

export interface RegisterTerminalResponse {
  terminal_id: string;
  tenant_id: string;
  branch_id: string;
  terminal_code: string;
  credential: string;
}

/* =========================
   Terminal Authentication
========================= */

export interface TerminalAuthRequest {
  credential: string;
}

export interface TerminalAuthResponse {
  terminal_id: string;
  tenant_id: string;
  branch_id: string;
}

/* =========================
   Terminal List
========================= */

export interface TerminalListResponse {
  terminals: Terminal[];
  total: number;
}

/* =========================
   Terminal Management
========================= */

export interface RenameTerminalRequest {
  machine_name: string;
}

export interface ChangeTerminalStatusRequest {
  status: TerminalStatus;
}

export interface MoveTerminalRequest {
  branch_id: string;
}

export interface ChangeTerminalLockStateRequest {
  state: TerminalLockState;
}

/* =========================
   Terminal Control
========================= */

export interface TerminalControlState {
  terminal_id: string;
  desired_state: TerminalLockState;
  command_version: number;
  updated_at: string;
}

export interface TerminalLockStateResponse {
  message: string;
  data: TerminalControlState;
}

/* =========================
   Licence Keys
========================= */

export interface LicenceKey {
  id: string;
  tenant_id?: string | null;
  branch_id: string;
  branch_name?: string | null;
  expires_at?: string | null;
  used_at?: string | null;
  revoked_at?: string | null;
  created_by?: string | null;
  created_at: string;
}

export interface LicenceKeyListResponse {
  licence_keys: LicenceKey[];
  total: number;
}

export interface GenerateLicenceKeyRequest {
  branch_id: string;
  expires_in_hours?: number;
}

export interface GenerateLicenceKeyResponse {
  id: string;
  branch_id: string;
  branch_name?: string | null;
  licence_key: string;
  expires_at: string;
  created_at: string;
}

/* =========================
   Generic API
========================= */

export interface ApiMessageResponse {
  message: string;
}