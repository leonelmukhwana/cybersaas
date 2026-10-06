"use client";

import { create } from "zustand";

import { sessionService } from "@/services/session.service";
import { attendantService } from "@/services/attendant.service";

import type {
  BillingPreview,
  CancelSessionRequest,
  EndSessionRequest,
  Session,
  SessionListResponse,
  SessionWithBilling,
  StartSessionRequest,
  TerminalBillingConfig,
  TerminalStartSessionRequest,
} from "@/types/session";

interface SessionStore {
  sessions: Session[];
  total: number;

  selectedSession: Session | null;
  billingPreview: BillingPreview | null;
  completedSession: SessionWithBilling | null;
  terminalBillingConfig: TerminalBillingConfig | null;

  loading: boolean;
  saving: boolean;
  error: string | null;

  fetchSessions: (
    limit?: number,
    offset?: number
  ) => Promise<void>;

  getSession: (
    sessionId: string
  ) => Promise<Session>;

  startSession: (
    data: StartSessionRequest
  ) => Promise<Session>;

  previewBilling: (
    sessionId: string
  ) => Promise<BillingPreview>;

  endSession: (
    sessionId: string,
    data?: EndSessionRequest
  ) => Promise<SessionWithBilling>;

  cancelSession: (
    sessionId: string,
    data: CancelSessionRequest
  ) => Promise<void>;

  pauseSession: (
    sessionId: string
  ) => Promise<Session>;

  resumeSession: (
    sessionId: string
  ) => Promise<Session>;

  terminalStartSession: (
    data: TerminalStartSessionRequest
  ) => Promise<Session>;

  fetchTerminalActiveSession: () => Promise<Session | null>;

  getTerminalSession: (
    sessionId: string
  ) => Promise<Session>;

  terminalPreviewBilling: (
    sessionId: string
  ) => Promise<BillingPreview>;

  pauseTerminalSession: (
    sessionId: string
  ) => Promise<void>;

  resumeTerminalSession: (
    sessionId: string
  ) => Promise<void>;

  endTerminalSession: (
    sessionId: string
  ) => Promise<SessionWithBilling>;

  fetchTerminalBillingConfig: () => Promise<TerminalBillingConfig>;

  clearSelectedSession: () => void;
  clearBillingPreview: () => void;
  clearCompletedSession: () => void;
  clearError: () => void;
}

export const useSessionStore = create<SessionStore>(
  (set) => ({
    sessions: [],
    total: 0,

    selectedSession: null,
    billingPreview: null,
    completedSession: null,
    terminalBillingConfig: null,

    loading: false,
    saving: false,
    error: null,

    // =========================================================
    // GET CURRENT ATTENDANT BRANCH
    // =========================================================

    getAttendantBranchId: undefined as never,

    // =========================================================
    // FETCH ONLINE ATTENDANT SESSIONS
    // =========================================================

    fetchSessions: async (
      limit = 50,
      offset = 0
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        const response: SessionListResponse =
          await sessionService.list(
            branchId,
            limit,
            offset
          );

        set({
          sessions: response.sessions,
          total: response.total,
          loading: false,
          error: null,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load sessions.",
        });

        throw error;
      }
    },

    // =========================================================
    // GET SESSION
    // =========================================================

    getSession: async (sessionId) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        const session =
          await sessionService.get(
            sessionId,
            branchId
          );

        set({
          selectedSession: session,
          loading: false,
        });

        return session;
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load session.",
        });

        throw error;
      }
    },

    // =========================================================
    // START SESSION
    // =========================================================

    startSession: async (data) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const session =
          await sessionService.start(data);

        set((state) => ({
          sessions: [
            session,
            ...state.sessions,
          ],
          total: state.total + 1,
          selectedSession: session,
          saving: false,
        }));

        return session;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to start session.",
        });

        throw error;
      }
    },

    // =========================================================
    // BILLING PREVIEW
    // =========================================================

    previewBilling: async (sessionId) => {
      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        const preview =
          await sessionService.billingPreview(
            sessionId,
            branchId
          );

        set({
          billingPreview: preview,
          error: null,
        });

        return preview;
      } catch (error) {
        set({
          error:
            error instanceof Error
              ? error.message
              : "Failed to calculate billing preview.",
        });

        throw error;
      }
    },

    // =========================================================
    // END SESSION
    // =========================================================

    endSession: async (
      sessionId,
      data = {}
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        const completed =
          await sessionService.end(
            sessionId,
            branchId,
            data
          );

        set((state) => ({
          sessions:
            state.sessions.map(
              (session) =>
                session.id === sessionId
                  ? completed
                  : session
            ),

          selectedSession:
            state.selectedSession?.id ===
            sessionId
              ? completed
              : state.selectedSession,

          completedSession:
            completed,

          billingPreview: null,

          saving: false,
        }));

        return completed;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to end session.",
        });

        throw error;
      }
    },

    // =========================================================
    // CANCEL SESSION
    // =========================================================

    cancelSession: async (
      sessionId,
      data
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        await sessionService.cancel(
          sessionId,
          branchId,
          data
        );

        const updated =
          await sessionService.get(
            sessionId,
            branchId
          );

        set((state) => ({
          sessions:
            state.sessions.map(
              (session) =>
                session.id === sessionId
                  ? updated
                  : session
            ),

          selectedSession:
            state.selectedSession?.id ===
            sessionId
              ? updated
              : state.selectedSession,

          billingPreview: null,

          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to cancel session.",
        });

        throw error;
      }
    },

    // =========================================================
    // PAUSE SESSION
    // =========================================================

    pauseSession: async (sessionId) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        const updated =
          await sessionService.pause(
            sessionId,
            branchId
          );

        set((state) => ({
          sessions:
            state.sessions.map(
              (session) =>
                session.id === sessionId
                  ? updated
                  : session
            ),

          selectedSession:
            state.selectedSession?.id ===
            sessionId
              ? updated
              : state.selectedSession,

          billingPreview: null,

          saving: false,
        }));

        return updated;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to pause session.",
        });

        throw error;
      }
    },

    // =========================================================
    // RESUME SESSION
    // =========================================================

    resumeSession: async (sessionId) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.me();

        const branchId =
          attendant.branch_id;

        if (!branchId) {
          throw new Error(
            "No branch is assigned to this attendant."
          );
        }

        const updated =
          await sessionService.resume(
            sessionId,
            branchId
          );

        set((state) => ({
          sessions:
            state.sessions.map(
              (session) =>
                session.id === sessionId
                  ? updated
                  : session
            ),

          selectedSession:
            state.selectedSession?.id ===
            sessionId
              ? updated
              : state.selectedSession,

          billingPreview: null,

          saving: false,
        }));

        return updated;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to resume session.",
        });

        throw error;
      }
    },

    // =========================================================
    // TERMINAL SESSION METHODS
    // =========================================================

    terminalStartSession: async (data) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const session =
          await sessionService.terminalStart(
            data
          );

        set({
          selectedSession: session,
          saving: false,
        });

        return session;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to start terminal session.",
        });

        throw error;
      }
    },

    fetchTerminalActiveSession:
      async () => {
        set({
          loading: true,
          error: null,
        });

        try {
          const session =
            await sessionService.terminalActive();

          set({
            selectedSession: session,
            loading: false,
          });

          return session;
        } catch (error) {
          set({
            loading: false,
            error:
              error instanceof Error
                ? error.message
                : "Failed to load active terminal session.",
          });

          throw error;
        }
      },

    getTerminalSession: async (
      sessionId
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const session =
          await sessionService.terminalGet(
            sessionId
          );

        set({
          selectedSession: session,
          loading: false,
        });

        return session;
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load terminal session.",
        });

        throw error;
      }
    },

    terminalPreviewBilling: async (
      sessionId
    ) => {
      try {
        const preview =
          await sessionService.terminalBillingPreview(
            sessionId
          );

        set({
          billingPreview: preview,
          error: null,
        });

        return preview;
      } catch (error) {
        set({
          error:
            error instanceof Error
              ? error.message
              : "Failed to calculate terminal billing preview.",
        });

        throw error;
      }
    },

    pauseTerminalSession: async (
      sessionId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await sessionService.terminalPause(
          sessionId
        );

        const updated =
          await sessionService.terminalGet(
            sessionId
          );

        set({
          selectedSession: updated,
          saving: false,
        });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to pause terminal session.",
        });

        throw error;
      }
    },

    resumeTerminalSession: async (
      sessionId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await sessionService.terminalResume(
          sessionId
        );

        const updated =
          await sessionService.terminalGet(
            sessionId
          );

        set({
          selectedSession: updated,
          saving: false,
        });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to resume terminal session.",
        });

        throw error;
      }
    },

    endTerminalSession: async (
      sessionId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const completed =
          await sessionService.terminalEnd(
            sessionId
          );

        set({
          selectedSession: completed,
          completedSession: completed,
          billingPreview: null,
          saving: false,
        });

        return completed;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to end terminal session.",
        });

        throw error;
      }
    },

    fetchTerminalBillingConfig:
      async () => {
        try {
          const config =
            await sessionService.terminalBillingConfig();

          set({
            terminalBillingConfig: config,
            error: null,
          });

          return config;
        } catch (error) {
          set({
            error:
              error instanceof Error
                ? error.message
                : "Failed to load terminal billing configuration.",
          });

          throw error;
        }
      },

    // =========================================================
    // CLEAR
    // =========================================================

    clearSelectedSession: () => {
      set({
        selectedSession: null,
      });
    },

    clearBillingPreview: () => {
      set({
        billingPreview: null,
      });
    },

    clearCompletedSession: () => {
      set({
        completedSession: null,
      });
    },

    clearError: () => {
      set({
        error: null,
      });
    },
  })
);