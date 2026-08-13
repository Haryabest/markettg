"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

type AuthState = {
  token: string | null;
  refreshToken: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  apiFetch: (path: string, options?: RequestInit) => Promise<Response>;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      refreshToken: null,
      login: async (email, password) => {
        const res = await fetch(`${API_URL}/api/v1/admin/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });
        if (!res.ok) throw new Error("Invalid credentials");
        const data = await res.json();
        set({ token: data.access_token, refreshToken: data.refresh_token });
      },
      logout: () => set({ token: null, refreshToken: null }),
      apiFetch: async (path, options = {}) => {
        const { token } = get();
        return fetch(`${API_URL}${path}`, {
          ...options,
          headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
            ...(options.headers as Record<string, string>),
          },
        });
      },
    }),
    { name: "admin-auth" }
  )
);
