
import { create } from "zustand";

import { terminalService } from "@/services/terminal.service";

import type {
  ChangeTerminalLockStateRequest,
  ChangeTerminalStatusRequest,
  GenerateLicenceKeyRequest,
  GenerateLicenceKeyResponse,
  LicenceKey,
  MoveTerminalRequest,
  RenameTerminalRequest,
  Terminal,
} from "@/types/terminal";

interface TerminalStore {
  terminals: Terminal[];
  total: number;

  licenceKeys: LicenceKey[];
  licenceKeysTotal: number;

  loading: boolean;
  saving: boolean;
  error: string | null;

  selectedTerminal: Terminal | null;
  generatedLicenceKey: GenerateLicenceKeyResponse | null;

  // Owner
  fetchTerminals: (
    limit?: number,
    offset?: number
  ) => Promise<void>;

  getTerminal: (
    terminalId: string
  ) => Promise<Terminal>;

  renameTerminal: (
    terminalId: string,
    data: RenameTerminalRequest
  ) => Promise<void>;

  changeTerminalStatus: (
    terminalId: string,
    data: ChangeTerminalStatusRequest
  ) => Promise<void>;

  moveTerminal: (
    terminalId: string,
    data: MoveTerminalRequest
  ) => Promise<void>;

  changeTerminalLockState: (
    terminalId: string,
    data: ChangeTerminalLockStateRequest
  ) => Promise<void>;

  deregisterTerminal: (
    terminalId: string
  ) => Promise<void>;

  // Attendant
  fetchAttendantTerminals: (
    limit?: number,
    offset?: number
  ) => Promise<void>;

  getAttendantTerminal: (
    terminalId: string
  ) => Promise<Terminal>;

  changeAttendantLockState: (
    terminalId: string,
    data: ChangeTerminalLockStateRequest
  ) => Promise<void>;

  restartAttendantTerminal: (
    terminalId: string
  ) => Promise<void>;

  shutdownAttendantTerminal: (
    terminalId: string
  ) => Promise<void>;

  // Licence keys
  fetchLicenceKeys: (
    limit?: number,
    offset?: number
  ) => Promise<void>;

  generateLicenceKey: (
    data: GenerateLicenceKeyRequest
  ) => Promise<GenerateLicenceKeyResponse>;

  revokeLicenceKey: (
    licenceKeyId: string
  ) => Promise<void>;

  clearGeneratedLicenceKey: () => void;
  clearError: () => void;
  clearSelectedTerminal: () => void;
}

export const useTerminalStore =
  create<TerminalStore>((set) => ({
    terminals: [],
    total: 0,

    licenceKeys: [],
    licenceKeysTotal: 0,

    loading: false,
    saving: false,
    error: null,

    selectedTerminal: null,
    generatedLicenceKey: null,

    // =========================================================
    // OWNER TERMINALS
    // =========================================================

    // =========================
    // Fetch terminals
    // =========================

    fetchTerminals: async (
      limit = 50,
      offset = 0
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const response =
          await terminalService.list(
            limit,
            offset
          );

        set({
          terminals: response.terminals,
          total: response.total,
          loading: false,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load terminals.",
        });

        throw error;
      }
    },

    // =========================
    // Get terminal
    // =========================

    getTerminal: async (
      terminalId: string
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const terminal =
          await terminalService.get(
            terminalId
          );

        set({
          selectedTerminal: terminal,
          loading: false,
        });

        return terminal;
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load terminal.",
        });

        throw error;
      }
    },

    // =========================
    // Rename terminal
    // =========================

    renameTerminal: async (
      terminalId,
      data
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.rename(
          terminalId,
          data
        );

        const updated =
          await terminalService.get(
            terminalId
          );

        set((state) => ({
          terminals: state.terminals.map(
            (terminal) =>
              terminal.id === terminalId
                ? updated
                : terminal
          ),
          selectedTerminal:
            state.selectedTerminal?.id ===
            terminalId
              ? updated
              : state.selectedTerminal,
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to rename terminal.",
        });

        throw error;
      }
    },

    // =========================
    // Change terminal status
    // =========================

    changeTerminalStatus: async (
      terminalId,
      data
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.changeStatus(
          terminalId,
          data
        );

        const updated =
          await terminalService.get(
            terminalId
          );

        set((state) => ({
          terminals: state.terminals.map(
            (terminal) =>
              terminal.id === terminalId
                ? updated
                : terminal
          ),
          selectedTerminal:
            state.selectedTerminal?.id ===
            terminalId
              ? updated
              : state.selectedTerminal,
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to change terminal status.",
        });

        throw error;
      }
    },

    // =========================
    // Move terminal
    // =========================

    moveTerminal: async (
      terminalId,
      data
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.move(
          terminalId,
          data
        );

        const updated =
          await terminalService.get(
            terminalId
          );

        set((state) => ({
          terminals: state.terminals.map(
            (terminal) =>
              terminal.id === terminalId
                ? updated
                : terminal
          ),
          selectedTerminal:
            state.selectedTerminal?.id ===
            terminalId
              ? updated
              : state.selectedTerminal,
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to move terminal.",
        });

        throw error;
      }
    },

    // =========================
    // Owner lock / unlock
    // =========================

    changeTerminalLockState: async (
      terminalId,
      data
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.changeLockState(
          terminalId,
          data
        );

        set({
          saving: false,
        });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to change terminal lock state.",
        });

        throw error;
      }
    },

    // =========================
    // Deregister terminal
    // =========================

    deregisterTerminal: async (
      terminalId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.deregister(
          terminalId
        );

        set((state) => ({
          terminals:
            state.terminals.filter(
              (terminal) =>
                terminal.id !== terminalId
            ),
          total: Math.max(
            0,
            state.total - 1
          ),
          selectedTerminal:
            state.selectedTerminal?.id ===
            terminalId
              ? null
              : state.selectedTerminal,
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to deregister terminal.",
        });

        throw error;
      }
    },

    // =========================================================
    // ATTENDANT TERMINALS
    // =========================================================

    // =========================
    // Fetch assigned terminals
    // =========================

    fetchAttendantTerminals: async (
      limit = 50,
      offset = 0
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const response =
          await terminalService.listAttendantTerminals(
            limit,
            offset
          );

        set({
          terminals: response.terminals,
          total: response.total,
          loading: false,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load assigned terminals.",
        });

        throw error;
      }
    },

    // =========================
    // Get assigned terminal
    // =========================

    getAttendantTerminal: async (
      terminalId: string
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const terminal =
          await terminalService.getAttendantTerminal(
            terminalId
          );

        set({
          selectedTerminal: terminal,
          loading: false,
        });

        return terminal;
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load terminal.",
        });

        throw error;
      }
    },

    // =========================
    // Attendant lock / unlock
    // =========================

    changeAttendantLockState: async (
      terminalId,
      data
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.changeAttendantLockState(
          terminalId,
          data
        );

        set({
          saving: false,
        });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to change terminal lock state.",
        });

        throw error;
      }
    },

    // =========================
    // Attendant restart
    // =========================

    restartAttendantTerminal: async (
      terminalId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.restartAttendantTerminal(
          terminalId
        );

        set({
          saving: false,
        });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to restart terminal.",
        });

        throw error;
      }
    },

    // =========================
    // Attendant shutdown
    // =========================

    shutdownAttendantTerminal: async (
      terminalId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.shutdownAttendantTerminal(
          terminalId
        );

        set({
          saving: false,
        });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to shut down terminal.",
        });

        throw error;
      }
    },

    // =========================================================
    // LICENCE KEYS
    // =========================================================

    // =========================
    // Fetch licence keys
    // =========================

    fetchLicenceKeys: async (
      limit = 50,
      offset = 0
    ) => {
      set({
        loading: true,
        error: null,
      });

      try {
        const response =
          await terminalService.listLicenceKeys(
            limit,
            offset
          );

        set({
          licenceKeys:
            response.licence_keys,
          licenceKeysTotal:
            response.total,
          loading: false,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load licence keys.",
        });

        throw error;
      }
    },

    // =========================
    // Generate licence key
    // =========================

    generateLicenceKey: async (
      data
    ) => {
      set({
        saving: true,
        error: null,
        generatedLicenceKey: null,
      });

      try {
        const response =
          await terminalService.generateLicenceKey(
            data
          );

        const licenceKey: LicenceKey = {
          id: response.id,
          branch_id: response.branch_id,
          branch_name:
            response.branch_name ?? null,
          expires_at:
            response.expires_at,
          used_at: null,
          revoked_at: null,
          created_by: null,
          created_at:
            response.created_at,
        };

        set((state) => ({
          generatedLicenceKey: response,

          licenceKeys: [
            licenceKey,
            ...state.licenceKeys,
          ],

          licenceKeysTotal:
            state.licenceKeysTotal + 1,

          saving: false,
        }));

        return response;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to generate licence key.",
        });

        throw error;
      }
    },

    // =========================
    // Revoke licence key
    // =========================

    revokeLicenceKey: async (
      licenceKeyId
    ) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await terminalService.revokeLicenceKey(
          licenceKeyId
        );

        set((state) => ({
          licenceKeys:
            state.licenceKeys.map(
              (licenceKey) =>
                licenceKey.id ===
                licenceKeyId
                  ? {
                      ...licenceKey,
                      revoked_at:
                        new Date().toISOString(),
                    }
                  : licenceKey
            ),
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to revoke licence key.",
        });

        throw error;
      }
    },

    // =========================================================
    // CLEAR ACTIONS
    // =========================================================

    clearGeneratedLicenceKey: () => {
      set({
        generatedLicenceKey: null,
      });
    },

    clearError: () => {
      set({
        error: null,
      });
    },

    clearSelectedTerminal: () => {
      set({
        selectedTerminal: null,
      });
    },
  }));
