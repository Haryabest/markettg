"use client";

import { useEffect } from "react";
import WebApp from "@twa-dev/sdk";
import { useAuthStore } from "@/stores/app";

export function TelegramProvider({ children }: { children: React.ReactNode }) {
  const { setInitData, authenticate } = useAuthStore();

  useEffect(() => {
    try {
      WebApp.ready();
      WebApp.expand();
      if (WebApp.initData) {
        setInitData(WebApp.initData);
        authenticate();
      } else {
        authenticate();
      }
    } catch {
      authenticate();
    }
  }, [setInitData, authenticate]);

  return <>{children}</>;
}
