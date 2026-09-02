"use client";



import { useEffect } from "react";

import { useAuthStore, useCartStore } from "@/stores/app";

import { collectTelegramSession, isTelegramWebApp } from "@/lib/telegram";



export function TelegramProvider({ children }: { children: React.ReactNode }) {

  const { setInitData, setTelegramUser, authenticate } = useAuthStore();

  const fetchCart = useCartStore((s) => s.fetchCart);



  useEffect(() => {

    let cancelled = false;



    const applySession = async () => {

      const { initData, user } = await collectTelegramSession();

      if (cancelled) return;



      if (initData) setInitData(initData);

      if (user) setTelegramUser(user);



      try {

        const { default: WebApp } = await import("@twa-dev/sdk");

        if (cancelled) return;



        WebApp.ready();

        WebApp.expand();



        const dark = WebApp.colorScheme === "dark";

        document.documentElement.classList.toggle("dark", dark);

        document.documentElement.style.colorScheme = dark ? "dark" : "light";



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

    };

  }, [setInitData, setTelegramUser, authenticate, fetchCart]);



  return <>{children}</>;

}


