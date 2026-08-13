"use client";

import { create } from "zustand";
import { api, type Cart, type User } from "@/lib/api";

type AuthState = {
  initData: string;
  user: User | null;
  isReady: boolean;
  setInitData: (data: string) => void;
  authenticate: () => Promise<void>;
};

export const useAuthStore = create<AuthState>((set, get) => ({
  initData: "",
  user: null,
  isReady: false,
  setInitData: (data) => {
    api.setInitData(data);
    set({ initData: data });
  },
  authenticate: async () => {
    const { initData } = get();
    if (!initData) {
      set({ isReady: true });
      return;
    }
    try {
      const res = await api.authTelegram(initData);
      set({ user: res.user, isReady: true });
    } catch {
      set({ isReady: true });
    }
  },
}));

type CartState = {
  cart: Cart | null;
  isLoading: boolean;
  fetchCart: () => Promise<void>;
  addItem: (productId: string, quantity: number) => Promise<void>;
  removeItem: (productId: string) => Promise<void>;
  itemCount: () => number;
};

export const useCartStore = create<CartState>((set, get) => ({
  cart: null,
  isLoading: false,
  fetchCart: async () => {
    set({ isLoading: true });
    try {
      const cart = await api.getCart();
      set({ cart });
    } catch {
      set({ cart: { items: [] } });
    } finally {
      set({ isLoading: false });
    }
  },
  addItem: async (productId, quantity) => {
    const prev = get().cart;
    set({
      cart: {
        items: [...(prev?.items || []).filter((i) => i.product_id !== productId), { product_id: productId, quantity }],
      },
    });
    try {
      const cart = await api.updateCartItem(productId, quantity);
      set({ cart });
    } catch {
      if (prev) set({ cart: prev });
    }
  },
  removeItem: async (productId) => {
    const cart = await api.removeCartItem(productId);
    set({ cart });
  },
  itemCount: () => get().cart?.items?.reduce((s, i) => s + i.quantity, 0) || 0,
}));
