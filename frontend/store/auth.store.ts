
"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

import type { AuthUser } from "@/types/auth";

interface AuthState {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  isOffline: boolean;

  offlineBranchId: string | null;
  offlineBranchName: string | null;

  setAuth: (user: AuthUser, token: string) => void;

  setOfflineAuth: (data: {
    user: AuthUser;
    branchId: string;
    branchName: string;
  }) => void;

  updateUser: (user: AuthUser) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isOffline: false,

      offlineBranchId: null,
      offlineBranchName: null,

      setAuth: (user, token) =>
        set({
          user,
          token,
          isAuthenticated: true,
          isOffline: false,
          offlineBranchId: null,
          offlineBranchName: null,
        }),

      setOfflineAuth: ({
        user,
        branchId,
        branchName,
      }) =>
        set({
          user,
          token: null,
          isAuthenticated: true,
          isOffline: true,
          offlineBranchId: branchId,
          offlineBranchName: branchName,
        }),

      updateUser: (user) =>
        set({
          user,
        }),

      clearAuth: () =>
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          isOffline: false,
          offlineBranchId: null,
          offlineBranchName: null,
        }),
    }),
    {
      name: "cybersaas-auth",
    }
  )
);
