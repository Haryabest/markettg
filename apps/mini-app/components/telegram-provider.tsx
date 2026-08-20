"use client";

import { useEffect } from "react";
import { useAuthStore, useCartStore } from "@/stores/app";

export function TelegramProvider({ children }: { children: React.ReactNode }) {
  const { setInitData, authenticate } = useAuthStore();
  const fetchCart = useCartStore((s) => s.fetchCart);

  useEffect(() => {
    let cancelled = false;

    const boot = async () => {
      try {
        const { default: WebApp } = await import("@twa-dev/sdk");
        if (cancelled) return;
        WebApp.ready();
        WebApp.expand();
        const dark = WebApp.colorScheme === "dark";
        document.documentElement.classList.toggle("dark", dark);
        document.documentElement.style.colorScheme = dark ? "dark" : "light";
        if (WebApp.initData) {
          setInitData(WebApp.initData);
        }
      } catch {
        // browser preview without Telegram SDK
      }
      await authenticate();
      if (!cancelled) await fetchCart();
    };

    boot();
    return () => {
      cancelled = true;
    };
  }, [setInitData, authenticate, fetchCart]);

  return <>{children}</>;
}
