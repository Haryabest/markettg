"use client";

import { AppShell } from "@astryxdesign/core/AppShell";
import { BottomNav } from "@/components/bottom-nav";

export function AppFrame({ children }: { children: React.ReactNode }) {
  return (
    <AppShell variant="elevated" height="fill" contentPadding={0} mobileNav={false}>
      <div className="mx-auto min-h-full max-w-lg px-4 pb-28 pt-5">{children}</div>
      <BottomNav />
    </AppShell>
  );
}
