
import { api } from "@/lib/api";

import type {
  BillingPreview,
  CancelSessionRequest,
  EndSessionRequest,
  Session,
  SessionListResponse,
  StartSessionRequest,
  TerminalBillingConfig,
  TerminalStartSessionRequest,
  SessionWithBilling,
} from "@/types/session";

const BASE_PATH = "/api/sessions";

export const sessionService = {
  // =========================================================
  // ONLINE DASHBOARD
  // =========================================================

  async list(
    branchId: string,
    limit = 50,
    offset = 0
  ): Promise<SessionListResponse> {
    if (!branchId) {
      throw new Error("branch_id is required");
    }

    return api<SessionListResponse>(
      `${BASE_PATH}?branch_id=${encodeURIComponent(
        branchId
      )}&limit=${limit}&offset=${offset}`,
      {
        method: "GET",
      }
    );
  },

  async get(
    sessionId: string,
    branchId: string
  ): Promise<Session> {
    if (!branchId) {
      throw new Error("branch_id is required");
    }

    return api<Session>(
      `${BASE_PATH}/${sessionId}?branch_id=${encodeURIComponent(
        branchId
      )}`,
      {
        method: "GET",
      }
    );
  },

  async start(
    data: StartSessionRequest
  ): Promise<Session> {
    return api<Session>(BASE_PATH, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async billingPreview(
    sessionId: string,
    branchId: string
  ): Promise<BillingPreview> {
    return api<BillingPreview>(
      `${BASE_PATH}/${sessionId}/billing-preview?branch_id=${encodeURIComponent(
        branchId
      )}`,
      {
        method: "GET",
      }
    );
  },

  async end(
    sessionId: string,
    branchId: string,
    data: EndSessionRequest = {}
  ): Promise<SessionWithBilling> {
    return api<SessionWithBilling>(
      `${BASE_PATH}/${sessionId}/end?branch_id=${encodeURIComponent(
        branchId
      )}`,
      {
        method: "POST",
        body: JSON.stringify(data),
      }
    );
  },

  async cancel(
    sessionId: string,
    branchId: string,
    data: CancelSessionRequest
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/${sessionId}/cancel?branch_id=${encodeURIComponent(
        branchId
      )}`,
      {
        method: "POST",
        body: JSON.stringify(data),
      }
    );
  },

  async pause(
    sessionId: string,
    branchId: string
  ): Promise<Session> {
    return api<Session>(
      `${BASE_PATH}/${sessionId}/pause?branch_id=${encodeURIComponent(
        branchId
      )}`,
      {
        method: "POST",
      }
    );
  },

  async resume(
    sessionId: string,
    branchId: string
  ): Promise<Session> {
    return api<Session>(
      `${BASE_PATH}/${sessionId}/resume?branch_id=${encodeURIComponent(
        branchId
      )}`,
      {
        method: "POST",
      }
    );
  },

  // =========================================================
  // TERMINAL / PC CLIENT
  // =========================================================

  async terminalStart(
    data: TerminalStartSessionRequest
  ): Promise<Session> {
    return api<Session>(
      `${BASE_PATH}/terminal/start`,
      {
        method: "POST",
        body: JSON.stringify(data),
      }
    );
  },

  async terminalActive(): Promise<Session | null> {
    return api<Session | null>(
      `${BASE_PATH}/terminal/active`,
      {
        method: "GET",
      }
    );
  },

  async terminalGet(
    sessionId: string
  ): Promise<Session> {
    return api<Session>(
      `${BASE_PATH}/terminal/${sessionId}`,
      {
        method: "GET",
      }
    );
  },

  async terminalBillingPreview(
    sessionId: string
  ): Promise<BillingPreview> {
    return api<BillingPreview>(
      `${BASE_PATH}/terminal/${sessionId}/billing-preview`,
      {
        method: "GET",
      }
    );
  },

  async terminalPause(
    sessionId: string
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/terminal/${sessionId}/pause`,
      {
        method: "POST",
      }
    );
  },

  async terminalResume(
    sessionId: string
  ): Promise<void> {
    await api<void>(
      `${BASE_PATH}/terminal/${sessionId}/resume`,
      {
        method: "POST",
      }
    );
  },

  async terminalEnd(
    sessionId: string
  ): Promise<SessionWithBilling> {
    return api<SessionWithBilling>(
      `${BASE_PATH}/terminal/${sessionId}/end`,
      {
        method: "POST",
      }
    );
  },

  async terminalBillingConfig(): Promise<TerminalBillingConfig> {
    return api<TerminalBillingConfig>(
      `${BASE_PATH}/terminal/billing-config`,
      {
        method: "GET",
      }
    );
  },
};
