import {
  processSyncQueue,
  type SyncEventRequest,
  type SyncEventResponse,
  type SyncQueueResult,
} from "@/lib/offline/sync-queue";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:8080/api";

export interface SyncPullResponse {
  events: SyncEvent[];
  has_more: boolean;
  server_time: string;
}

export interface SyncEvent {
  id: string;
  tenant_id: string;
  branch_id: string;
  terminal_id?: string | null;
  user_id?: string | null;
  event_type: string;
  entity_type: string;
  entity_id: string;
  payload: unknown;
  status: string;
  attempts: number;
  last_error?: string | null;
  created_at: string;
  processed_at?: string | null;
}

export interface SyncState {
  id?: string;
  terminal_id: string;
  last_pull_at?: string | null;
  last_push_at?: string | null;
  last_successful_sync_at?: string | null;
  last_sync_error?: string | null;
  updated_at?: string;
}

export interface SyncClientOptions {
  terminalToken: string;
}

class SyncService {
  private terminalToken: string;

  constructor(options: SyncClientOptions) {
    if (!options.terminalToken) {
      throw new Error(
        "Terminal authentication token is required for synchronization.",
      );
    }

    this.terminalToken =
      options.terminalToken;
  }

  private getHeaders(): HeadersInit {
    return {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.terminalToken}`,
    };
  }

  async pushEvent(
    event: SyncEventRequest,
  ): Promise<SyncEventResponse> {
    const response = await fetch(
      `${API_BASE_URL}/sync/events`,
      {
        method: "POST",
        headers: this.getHeaders(),
        body: JSON.stringify(event),
      },
    );

    const data = await this.parseResponse(
      response,
    );

    if (!response.ok) {
      throw new Error(
        this.getErrorMessage(
          data,
          response.status,
        ),
      );
    }

    return data as SyncEventResponse;
  }

  async syncQueue(): Promise<SyncQueueResult> {
    return processSyncQueue(
      (event) => this.pushEvent(event),
    );
  }

  async pullEvents(
    since?: string,
    limit = 100,
  ): Promise<SyncPullResponse> {
    const params = new URLSearchParams();

    if (since) {
      params.set("since", since);
    }

    params.set(
      "limit",
      String(Math.min(Math.max(limit, 1), 500)),
    );

    const response = await fetch(
      `${API_BASE_URL}/sync/events?${params.toString()}`,
      {
        method: "GET",
        headers: this.getHeaders(),
      },
    );

    const data = await this.parseResponse(
      response,
    );

    if (!response.ok) {
      throw new Error(
        this.getErrorMessage(
          data,
          response.status,
        ),
      );
    }

    return data as SyncPullResponse;
  }

  async getState(): Promise<SyncState> {
    const response = await fetch(
      `${API_BASE_URL}/sync/state`,
      {
        method: "GET",
        headers: this.getHeaders(),
      },
    );

    const data = await this.parseResponse(
      response,
    );

    if (!response.ok) {
      throw new Error(
        this.getErrorMessage(
          data,
          response.status,
        ),
      );
    }

    return data as SyncState;
  }

  async markSuccessful(): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/sync/successful`,
      {
        method: "POST",
        headers: this.getHeaders(),
      },
    );

    const data = await this.parseResponse(
      response,
    );

    if (!response.ok) {
      throw new Error(
        this.getErrorMessage(
          data,
          response.status,
        ),
      );
    }
  }

  async sync(): Promise<SyncQueueResult> {
    const result =
      await this.syncQueue();

    if (
      result.failed === 0 &&
      result.processed > 0 &&
      result.remaining === 0
    ) {
      await this.markSuccessful();
    }

    return result;
  }

  private async parseResponse(
    response: Response,
  ): Promise<unknown> {
    const text =
      await response.text();

    if (!text) {
      return null;
    }

    try {
      return JSON.parse(text);
    } catch {
      return text;
    }
  }

  private getErrorMessage(
    data: unknown,
    status: number,
  ): string {
    if (
      typeof data === "object" &&
      data !== null &&
      "error" in data &&
      typeof data.error === "string"
    ) {
      return data.error;
    }

    if (
      typeof data === "object" &&
      data !== null &&
      "message" in data &&
      typeof data.message === "string"
    ) {
      return data.message;
    }

    if (typeof data === "string" && data) {
      return data;
    }

    return `Synchronization request failed with status ${status}.`;
  }
}

export function createSyncService(
  terminalToken: string,
): SyncService {
  return new SyncService({
    terminalToken,
  });
}