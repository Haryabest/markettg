"use client";

import { create } from "zustand";

export type BannerStatus = "info" | "success" | "warning" | "error";

export type AppBanner = {
  id: string;
  status: BannerStatus;
  title: string;
  description?: string;
  sticky?: boolean;
};

type BannerState = {
  items: AppBanner[];
  push: (banner: Omit<AppBanner, "id">) => string;
  dismiss: (id: string) => void;
  dismissAll: () => void;
};

export const useBannerStore = create<BannerState>((set) => ({
  items: [],
  push: (banner) => {
    const id = crypto.randomUUID();
    set((state) => ({ items: [{ ...banner, id }, ...state.items].slice(0, 3) }));
    if (!banner.sticky) {
      window.setTimeout(() => {
        set((state) => ({ items: state.items.filter((item) => item.id !== id) }));
      }, 6000);
    }
    return id;
  },
  dismiss: (id) => set((state) => ({ items: state.items.filter((item) => item.id !== id) })),
  dismissAll: () => set({ items: [] }),
}));

export function notifyError(title: string, description?: string) {
  return useBannerStore.getState().push({ status: "error", title, description });
}

export function notifySuccess(title: string, description?: string) {
  return useBannerStore.getState().push({ status: "success", title, description });
}

export function notifySlow(title: string, description?: string) {
  return useBannerStore.getState().push({ status: "info", title, description, sticky: true });
}

export function dismissBanner(id: string) {
  useBannerStore.getState().dismiss(id);
}
