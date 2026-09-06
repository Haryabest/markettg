"use client";

import { AppBanners } from "@/components/app-banners";
import { QueryProvider } from "@/components/providers";
import { TelegramProvider } from "@/components/telegram-provider";
import { AppFrame } from "@/components/app-frame";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <TelegramProvider>
        <AppBanners />
        <AppFrame>{children}</AppFrame>
      </TelegramProvider>
    </QueryProvider>
  );
}
