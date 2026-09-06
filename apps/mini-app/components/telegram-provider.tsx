"use client";

import { useEffect } from "react";
import { useAuthStore, useCartStore } from "@/stores/app";
import { collectTelegramSession, isTelegramWebApp } from "@/lib/telegram";
import { applyTelegramTheme, bindTelegramThemeListener } from "@/lib/telegram-theme";

export function TelegramProvider({ children }: { children: React.ReactNode }) {
  const { setInitData, setTelegramUser, authenticate } = useAuthStore();
  const fetchCart = useCartStore((s) => s.fetchCart);

  useEffect(() => {
    let cancelled = false;
    let unbindTheme: (() => void) | undefined;

    const applySession = async () => {
      const { initData, user } = await collectTelegramSession();
      if (cancelled) return;

      if (initData) setInitData(initData);
      if (user) setTelegramUser(user);

      const native = window.Telegram?.WebApp;
      if (native) {
        native.ready?.();
        native.expand?.();
        applyTelegramTheme(native);
      }

      try {
        const { default: WebApp } = await import("@twa-dev/sdk");
        if (cancelled) return;

        WebApp.ready();
        WebApp.expand();

        applyTelegramTheme(WebApp);
        unbindTheme?.();
        unbindTheme = bindTelegramThemeListener(WebApp);

        const sdkInitData = WebApp.initData?.trim() || "";
        const sdkUser = WebApp.initDataUnsafe?.user;
        if (sdkInitData) setInitData(sdkInitData);
        if (sdkUser?.id) setTelegramUser(sdkUser);
      } catch {
        // Native Telegram WebApp script is enough.
      }
    };

    const boot = async () => {
      await applySession();
      if (cancelled) return;

      await authenticate();

      if (!cancelled && isTelegramWebApp() && !useAuthStore.getState().user) {
        await applySession();
        await authenticate();
      }

      if (!cancelled) await fetchCart();
    };

    boot();
    return () => {
      cancelled = true;
      unbindTheme?.();
    };
  }, [setInitData, setTelegramUser, authenticate, fetchCart]);

  return <>{children}</>;
}
