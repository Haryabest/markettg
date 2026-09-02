"use client";

import { useEffect, useRef, useState } from "react";
import { Text } from "@astryxdesign/core/Text";
import { AppIcon } from "@/components/icons";
import { ArrowUpDown, Check } from "lucide-react";

export type SortOption = {
  value: string;
  label: string;
};

export function SortMenu({
  value,
  options,
  onChange,
}: {
  value: string;
  options: readonly SortOption[];
  onChange: (value: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const active = options.find((option) => option.value === value);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [open]);

  return (
    <div ref={rootRef} className="relative shrink-0">
      <button
        type="button"
        aria-label="Сортировка"
        aria-expanded={open}
        onClick={() => setOpen((prev) => !prev)}
        className="inline-flex h-9 items-center gap-1.5 rounded-full border border-[var(--color-border-default,rgba(0,0,0,.08))] bg-[var(--color-background-surface,#fff)] px-3 text-sm text-[var(--color-text-primary,#1b1b1b)] shadow-sm"
      >
        <AppIcon icon={ArrowUpDown} size={16} />
        <Text type="label" weight="medium" display="inline">
          {active?.label ?? "Сортировка"}
        </Text>
      </button>
      {open ? (
        <div className="absolute right-0 top-[calc(100%+6px)] z-20 min-w-[180px] overflow-hidden rounded-[12px] border border-[var(--color-border-default,rgba(0,0,0,.08))] bg-[var(--color-background-surface,#fff)] py-1 shadow-lg">
          {options.map((option) => {
            const selected = option.value === value;
            return (
              <button
                key={option.value}
                type="button"
                onClick={() => {
                  onChange(option.value);
                  setOpen(false);
                }}
                className="flex w-full items-center justify-between gap-3 px-3 py-2 text-left text-sm hover:bg-black/5 dark:hover:bg-white/5"
              >
                <span>{option.label}</span>
                {selected ? <AppIcon icon={Check} size={16} /> : null}
              </button>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}
