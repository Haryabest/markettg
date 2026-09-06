"use client";

import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function TextInput({
  label,
  value,
  onChange,
  placeholder,
  startIcon,
  hasClear,
  width,
  isLabelHidden,
}: {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  startIcon?: ReactNode;
  hasClear?: boolean;
  width?: string;
  isLabelHidden?: boolean;
}) {
  return (
    <label className={cn("flex flex-col gap-1", width === "100%" && "w-full")}>
      {label && !isLabelHidden ? (
        <span className="text-sm text-[var(--tg-hint-color,var(--color-text-secondary,#8a8a8e))]">{label}</span>
      ) : null}
      <span
        className={cn(
          "flex items-center gap-2 rounded-xl border bg-[var(--tg-secondary-bg-color,var(--color-background-surface,#fff))] px-3 py-2",
          "border-[var(--tg-button-color,var(--color-border-default,rgba(0,0,0,.16)))]"
        )}
      >
        {startIcon ? (
          <span className="shrink-0 text-[var(--tg-hint-color,var(--color-text-secondary,#8a8a8e))]">
            {startIcon}
          </span>
        ) : null}
        <input
          aria-label={isLabelHidden ? label || placeholder || "Поиск" : label}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          placeholder={placeholder}
          className="min-w-0 flex-1 bg-transparent text-[var(--tg-text-color,var(--foreground,#1b1b1b))] outline-none placeholder:text-[var(--tg-hint-color,#8a8a8e)]"
        />
        {hasClear && value ? (
          <button
            type="button"
            aria-label="Очистить"
            onClick={() => onChange("")}
            className="text-sm text-[var(--tg-hint-color,#8a8a8e)]"
          >
            ×
          </button>
        ) : null}
      </span>
    </label>
  );
}
