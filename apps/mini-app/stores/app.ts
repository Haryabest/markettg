"use client";

import { create } from "zustand";
import { api, type Cart, type User } from "@/lib/api";
import { mapTelegramUser, parseUserFromInitData } from "@/lib/telegram";
import { mergeUsers } from "@/lib/user-display";
import { notifyError } from "@/stores/banners";

const GUEST_CART_KEY = "markettg-cart-guest";

let activeTelegramId: number | null = null;

function cartStorageKey(telegramId?: number | null) {
  if (telegramId) return `markettg-cart-${telegramId}`;
  return GUEST_CART_KEY;
}

function readLocalCartForKey(key: string): Cart {
  if (typeof window === "undefined") return { items: [] };
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return { items: [] };
    return JSON.parse(raw) as Cart;
  } catch {
    return { items: [] };
  }
}

function readLocalCart(): Cart {
  return readLocalCartForKey(cartStorageKey(activeTelegramId));
}

function writeLocalCart(cart: Cart) {
  if (typeof window === "undefined") return;
  localStorage.setItem(cartStorageKey(activeTelegramId), JSON.stringify(cart));
}

export function bindCartToUser(telegramId?: number | null) {
  const prevKey = cartStorageKey(activeTelegramId);
  activeTelegramId = telegramId ?? null;
  const nextKey = cartStorageKey(activeTelegramId);
  if (prevKey !== nextKey) {
    useCartStore.setState({
      cart: readLocalCart(),
      isLocal: true,
    });
  }
}

type AuthState = {
  initData: string;
  user: User | null;
  telegramUser: User | null;
  isReady: boolean;
  setInitData: (data: string) => void;
  setTelegramUser: (user: {
    id: number;
    first_name?: string;
    last_name?: string;
    username?: string;
    photo_url?: string;
  }) => void;
  authenticate: () => Promise<void>;
  displayUser: () => User | null;
};

export const useAuthStore = create<AuthState>((set, get) => ({
  initData: "",
  user: null,
  telegramUser: null,
  isReady: false,
  setInitData: (data) => {
    api.setInitData(data);
    set({ initData: data });
  },
  setTelegramUser: (user) => {
    set({ telegramUser: mapTelegramUser(user) });
  },
  displayUser: () => get().user ?? get().telegramUser,
  authenticate: async () => {
    const { initData, telegramUser } = get();
    if (!initData) {
      bindCartToUser(telegramUser?.telegram_id ?? null);
      set({ isReady: true });
      return;
    }
    try {
      const res = await api.authTelegram(initData);
      let user = mergeUsers(res.user, telegramUser);
      try {
        user = mergeUsers(await api.getMe(), user);
      } catch {
        // auth response is enough
      }
      bindCartToUser(user.telegram_id);
      set({ user, telegramUser: user, isReady: true });
    } catch {
      const parsed = parseUserFromInitData(initData);
      const fallback = telegramUser ?? (parsed ? mapTelegramUser(parsed) : null);
      if (fallback?.telegram_id) set({ telegramUser: fallback });
      bindCartToUser(fallback?.telegram_id ?? null);
      set({ isReady: true });
    }
  },
}));

type CartState = {
  cart: Cart | null;
  isLocal: boolean;
  isLoading: boolean;
  fetchCart: () => Promise<void>;
  addItem: (productId: string, quantity: number) => Promise<boolean>;
  removeItem: (productId: string) => Promise<void>;
  itemCount: () => number;
  initLocal: () => void;
};

function upsertLocalItem(cart: Cart, productId: string, quantity: number): Cart {
  const items = [...cart.items.filter((i) => i.product_id !== productId)];
  if (quantity > 0) items.push({ product_id: productId, quantity });
  return { items };
}

export const useCartStore = create<CartState>((set, get) => ({
  cart: null,
  isLocal: false,
  isLoading: false,
  fetchCart: async () => {
    set({ isLoading: true });
    try {
      const cart = await api.getCart();
      writeLocalCart(cart);
      set({ cart, isLocal: false });
    } catch {
      const local = readLocalCart();
      set({ cart: local, isLocal: true });
    } finally {
      set({ isLoading: false });
    }
  },
  addItem: async (productId, quantity) => {
    const prev = get().cart || readLocalCart();
    const optimistic = upsertLocalItem(prev, productId, quantity);
    set({ cart: optimistic, isLocal: get().isLocal });

    try {
      const cart = await api.updateCartItem(productId, quantity);
      writeLocalCart(cart);
      set({ cart, isLocal: false });
      return true;
    } catch (error) {
      writeLocalCart(optimistic);
      set({ cart: optimistic, isLocal: true });
      const message = error instanceof Error ? error.message : "Не удалось сохранить на сервере";
      if (message.includes("401") || message.toLowerCase().includes("unauthorized")) {
        notifyError(
          "Корзина только на этом устройстве",
          "Для оплаты откройте магазин из Telegram."
        );
      } else {
        notifyError("Не удалось обновить корзину", message);
      }
      return true;
    }
  },
  removeItem: async (productId) => {
    const prev = get().cart;
    const optimistic = upsertLocalItem(prev || { items: [] }, productId, 0);
    set({ cart: optimistic });

    try {
      const cart = await api.removeCartItem(productId);
      writeLocalCart(cart);
      set({ cart, isLocal: false });
    } catch (error) {
      writeLocalCart(optimistic);
      set({ cart: optimistic, isLocal: true });
      const message = error instanceof Error ? error.message : "Ошибка удаления";
      notifyError("Не удалось удалить товар", message);
    }
  },
  itemCount: () => get().cart?.items?.reduce((s, i) => s + i.quantity, 0) || 0,
  initLocal: () => {
    const local = readLocalCart();
    if (local.items.length) set({ cart: local, isLocal: true });
  },
}));
