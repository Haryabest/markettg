"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { VStack } from "@astryxdesign/core/VStack";
import { AppIcon } from "@/components/icons";
import { cn } from "@/lib/utils";
import { useBannerStore, type AppBanner, type BannerStatus } from "@/stores/banners";
import { AlertTriangle, CheckCircle2, CircleAlert, Info, X } from "lucide-react";
import type { LucideIcon } from "lucide-react";

const TOAST_STYLES: Record<BannerStatus, { className: string; icon: LucideIcon }> = {
  error: {
    className: "bg-[#4a1818] border border-[#7a2b2b]",
    icon: CircleAlert,
  },
  success: {
    className: "bg-[#0f3d24] border border-[#1f6b42]",
    icon: CheckCircle2,
  },
  info: {
    className: "bg-[#0f2d4a] border border-[#1e4a7a]",
    icon: Info,
  },
  warning: {
    className: "bg-[#4a3a10] border border-[#7a5c18]",
    icon: AlertTriangle,
  },
};

function AppToast({ item, onDismiss }: { item: AppBanner; onDismiss: () => void }) {
  const toast = TOAST_STYLES[item.status];
  const role = item.status === "error" || item.status === "warning" ? "alert" : "status";

  return (
    <div
      role={role}
      className={cn(
        "pointer-events-auto flex items-start gap-3 rounded-xl p-4 shadow-lg",
        toast.className
      )}
    >
      <span className="mt-0.5 shrink-0 text-white">
        <AppIcon icon={toast.icon} size={20} />
      </span>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-semibold leading-snug text-white">{item.title}</div>
        {item.description ? (
          <div className="mt-1 text-sm leading-snug text-white/85">{item.description}</div>
        ) : null}
      </div>
      <button
        type="button"
        onClick={onDismiss}
        aria-label="Закрыть"
        className="shrink-0 rounded-md p-0.5 text-white/80 transition-colors hover:bg-white/10 hover:text-white"
      >
        <X size={18} />
      </button>
    </div>
  );
}

export function AppBanners() {
  const items = useBannerStore((s) => s.items);
  const dismiss = useBannerStore((s) => s.dismiss);
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  if (!mounted || !items.length) return null;

  return createPortal(
    <div
      className="pointer-events-none fixed inset-x-0 top-0 z-[10000] mx-auto max-w-lg px-3 pt-[max(0.5rem,env(safe-area-inset-top))]"
      aria-live="polite"
    >
      <VStack gap={2}>
        {items.map((item) => (
          <AppToast key={item.id} item={item} onDismiss={() => dismiss(item.id)} />
        ))}
      </VStack>
    </div>,
    document.body
  );
}
