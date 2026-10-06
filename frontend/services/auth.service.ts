import { api } from "@/lib/api";

import type {
  AuthResponse,
  ForgotPasswordRequest,
  LoginRequest,
  RegisterRequest,
  ResetPasswordRequest,
} from "@/types/auth";

export const authService = {
  async login(data: LoginRequest) {
    return api<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async register(data: RegisterRequest) {
    return api<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async forgotPassword(data: ForgotPasswordRequest) {
    return api<{ message: string }>(
      "/api/auth/forgot-password",
      {
        method: "POST",
        body: JSON.stringify(data),
      }
    );
  },

  async resetPassword(data: ResetPasswordRequest) {
    return api<{ message: string }>(
      "/api/auth/reset-password",
      {
        method: "POST",
        body: JSON.stringify({
          token: data.token,
          password: data.password,
        }),
      }
    );
  },

  async me(token: string) {
    return api<AuthResponse>("/api/auth/me", {
      method: "GET",
      token,
    });
  },
};