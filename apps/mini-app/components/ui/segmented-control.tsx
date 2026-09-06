"use client";

import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useActionIconTrigger } from "@/lib/action-icon-trigger";

export function SegmentedControl({
  label,
  value,
  onChange,
  layout,
  children,
}: {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  layout?: "fill";
  children: ReactNode;
}) {
  return (
    <div
      role="tablist"
      aria-label={label || "Выбор"}
      className={cn(
        "inline-flex rounded-full border border-[var(--tg-button-color,var(--color-border-default,rgba(0,0,0,.16)))] bg-[var(--tg-secondary-bg-color,var(--color-background-surface,#fff))] p-1",
        layout === "fill" && "flex w-full"
      )}
    >
      {children}
    </div>
  );
}

export function SegmentedControlItem({
  value,
  label,
  icon,
  activeValue,
  onSelect,
}: {
  value: string;
  label: string;
  icon?: ReactNode;
  activeValue?: string;
  onSelect?: (value: string) => void;
}) {
  const selected = activeValue === value;
  const { icon: wiredIcon, onParentPointerDown } = useActionIconTrigger(icon);

  return (
    <button
      type="button"
      role="tab"
      aria-selected={selected}
      onClick={() => onSelect?.(value)}
      onPointerDownCapture={onParentPointerDown}
      className={cn(
        "inline-flex flex-1 items-center justify-center gap-1.5 rounded-full px-3 py-2 text-sm font-medium transition-colors",
        selected
          ? "bg-[var(--tg-button-color,#007aff)] text-[var(--tg-button-text-color,#fff)]"
          : "text-[var(--foreground,#1b1b1b)] hover:bg-black/5 dark:hover:bg-white/10"
      )}
    >
      {wiredIcon}
      <span>{label}</span>
    </button>
  );
}
