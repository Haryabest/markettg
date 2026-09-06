"use client";

import { BottomNav } from "@/components/bottom-nav";

export function AppFrame({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-dvh bg-[var(--background,var(--tg-bg-color,#fff))] text-[var(--foreground,var(--tg-text-color,#1b1b1b))]">
      <div className="mx-auto min-h-full max-w-lg px-4 pb-28 pt-5">{children}</div>
      <BottomNav />
    </div>
  );
}
