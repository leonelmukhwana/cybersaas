"use client";

import { create } from "zustand";

import { attendantService } from "@/services/attendant.service";

import type {
  Attendant,
  CreateAttendantRequest,
  UpdateAttendantRequest,
  ChangeAttendantStatusRequest,
} from "@/types/attendant";

interface AttendantStore {
  attendants: Attendant[];
  total: number;

  loading: boolean;
  saving: boolean;
  error: string | null;

  fetchAttendants: () => Promise<void>;

  createAttendant: (
    data: CreateAttendantRequest
  ) => Promise<Attendant>;

  updateAttendant: (
    id: string,
    data: UpdateAttendantRequest
  ) => Promise<void>;

  changeStatus: (
    id: string,
    data: ChangeAttendantStatusRequest
  ) => Promise<void>;

  assignBranch: (
    id: string,
    branchId: string
  ) => Promise<void>;

  unassignBranch: (
    id: string
  ) => Promise<void>;

  clearError: () => void;
}

export const useAttendantStore =
  create<AttendantStore>((set) => ({
    attendants: [],
    total: 0,

    loading: false,
    saving: false,
    error: null,

    clearError: () => {
      set({ error: null });
    },

    fetchAttendants: async () => {
      set({
        loading: true,
        error: null,
      });

      try {
        const response =
          await attendantService.list();

        set({
          attendants: response.attendants,
          total: response.total,
          loading: false,
        });
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load attendants.",
        });

        throw error;
      }
    },

    createAttendant: async (data) => {
      set({
        saving: true,
        error: null,
      });

      try {
        const attendant =
          await attendantService.create(data);

        set((state) => ({
          attendants: [
            attendant,
            ...state.attendants,
          ],
          total: state.total + 1,
          saving: false,
        }));

        return attendant;
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to create attendant.",
        });

        throw error;
      }
    },

    updateAttendant: async (id, data) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await attendantService.update(
          id,
          data
        );

        const updated =
          await attendantService.get(id);

        set((state) => ({
          attendants: state.attendants.map(
            (attendant) =>
              attendant.id === id
                ? updated
                : attendant
          ),
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to update attendant.",
        });

        throw error;
      }
    },

    changeStatus: async (id, data) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await attendantService.changeStatus(
          id,
          data
        );

        set((state) => ({
          attendants: state.attendants.map(
            (attendant) =>
              attendant.id === id
                ? {
                    ...attendant,
                    status: data.status,
                  }
                : attendant
          ),
          saving: false,
        }));
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to change attendant status.",
        });

        throw error;
      }
    },

    assignBranch: async (id, branchId) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await attendantService.assignBranch(
          id,
          {
            branch_id: branchId,
          }
        );

        await useAttendantStore
          .getState()
          .fetchAttendants();

        set({ saving: false });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to assign branch.",
        });

        throw error;
      }
    },

    unassignBranch: async (id) => {
      set({
        saving: true,
        error: null,
      });

      try {
        await attendantService.unassignBranch(
          id
        );

        await useAttendantStore
          .getState()
          .fetchAttendants();

        set({ saving: false });
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to unassign branch.",
        });

        throw error;
      }
    },
  }));