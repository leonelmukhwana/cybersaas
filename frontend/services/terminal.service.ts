
import { api } from "@/lib/api";
import type {
  ApiMessageResponse,
  ChangeTerminalLockStateRequest,
  ChangeTerminalStatusRequest,
  GenerateLicenceKeyRequest,
  GenerateLicenceKeyResponse,
  LicenceKeyListResponse,
  MoveTerminalRequest,
  RegisterTerminalRequest,
  RenameTerminalRequest,
  Terminal,
  TerminalControlState,
  TerminalListResponse,
  TerminalLockStateResponse,
} from "@/types/terminal";

export const terminalService = {
  // =========================
  // Terminal registration
  // =========================

  async register(
    data: RegisterTerminalRequest
  ): Promise<{
    terminal_id: string;
    tenant_id: string;
    branch_id: string;
    terminal_code: string;
    credential: string;
  }> {
    return api("/api/terminals/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  // =========================
  // Terminal authentication
  // =========================

  async authenticate(
    credential: string
  ): Promise<{
    terminal_id: string;
    tenant_id: string;
    branch_id: string;
  }> {
    return api("/api/terminals/auth", {
      method: "POST",
      body: JSON.stringify({
        credential,
      }),
    });
  },

  // =========================
  // Owner terminal management
  // =========================

  async list(
    limit = 50,
    offset = 0
  ): Promise<TerminalListResponse> {
    return api(
      `/api/terminals?limit=${limit}&offset=${offset}`,
      {
        method: "GET",
      }
    );
  },

  async get(
    terminalId: string
  ): Promise<Terminal> {
    return api(
      `/api/terminals/${terminalId}`,
      {
        method: "GET",
      }
    );
  },

  async rename(
    terminalId: string,
    data: RenameTerminalRequest
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/terminals/${terminalId}/name`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async changeStatus(
    terminalId: string,
    data: ChangeTerminalStatusRequest
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/terminals/${terminalId}/status`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async move(
    terminalId: string,
    data: MoveTerminalRequest
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/terminals/${terminalId}/branch`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async changeLockState(
    terminalId: string,
    data: ChangeTerminalLockStateRequest
  ): Promise<TerminalLockStateResponse> {
    return api(
      `/api/terminals/${terminalId}/lock-state`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async deregister(
    terminalId: string
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/terminals/${terminalId}`,
      {
        method: "DELETE",
      }
    );
  },

  // =========================
  // Attendant terminal management
  // =========================

  async listAttendantTerminals(
    limit = 50,
    offset = 0
  ): Promise<TerminalListResponse> {
    return api(
      `/api/attendant/terminals?limit=${limit}&offset=${offset}`,
      {
        method: "GET",
      }
    );
  },

  async getAttendantTerminal(
    terminalId: string
  ): Promise<Terminal> {
    return api(
      `/api/attendant/terminals/${terminalId}`,
      {
        method: "GET",
      }
    );
  },

  async changeAttendantLockState(
    terminalId: string,
    data: ChangeTerminalLockStateRequest
  ): Promise<TerminalLockStateResponse> {
    return api(
      `/api/attendant/terminals/${terminalId}/lock-state`,
      {
        method: "PATCH",
        body: JSON.stringify(data),
      }
    );
  },

  async restartAttendantTerminal(
    terminalId: string
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/attendant/terminals/${terminalId}/restart`,
      {
        method: "POST",
      }
    );
  },

  async shutdownAttendantTerminal(
    terminalId: string
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/attendant/terminals/${terminalId}/shutdown`,
      {
        method: "POST",
      }
    );
  },

  // =========================
  // Licence keys
  // =========================

  async generateLicenceKey(
  data: GenerateLicenceKeyRequest
): Promise<GenerateLicenceKeyResponse> {
  const response = await api<
    GenerateLicenceKeyResponse | {
      data: GenerateLicenceKeyResponse;
    }
  >(
    "/api/terminals/licence-keys",
    {
      method: "POST",
      body: JSON.stringify(data),
    }
  );

  // Backend returns { data: {...} }
  if (
    response &&
    typeof response === "object" &&
    "data" in response &&
    response.data
  ) {
    return response.data;
  }

  // Also support a direct {...} response
  return response as GenerateLicenceKeyResponse;
},
  async listLicenceKeys(
    limit = 50,
    offset = 0
  ): Promise<LicenceKeyListResponse> {
    return api(
      `/api/terminals/licence-keys?limit=${limit}&offset=${offset}`,
      {
        method: "GET",
      }
    );
  },

  async revokeLicenceKey(
    licenceKeyId: string
  ): Promise<ApiMessageResponse> {
    return api(
      `/api/terminals/licence-keys/${licenceKeyId}`,
      {
        method: "DELETE",
      }
    );
  },

  // =========================
  // Terminal heartbeat
  // =========================

  async heartbeat(
    machineName?: string,
    deviceIdentifier?: string,
    token?: string
  ): Promise<{
    terminal_id: string;
    server_time: string;
    desired_state: "locked" | "unlocked";
    command_version: number;
  }> {
    return api(
      "/api/terminals/heartbeat",
      {
        method: "POST",
        token,
        body: JSON.stringify({
          machine_name: machineName ?? null,
          device_identifier: deviceIdentifier ?? null,
        }),
      }
    );
  },

  // =========================
  // Terminal control state
  // =========================

  async getControlState(
    token?: string
  ): Promise<TerminalControlState> {
    return api(
      "/api/terminals/control-state",
      {
        method: "GET",
        token,
      }
    );
  },
};
