"use client";

import { useEffect, useRef, useState } from "react";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon } from "@/components/icons";
import { useActionIconTrigger } from "@/lib/action-icon-trigger";
import { Check } from "lucide-react";

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
  const { icon: wiredIcon, onParentPointerDown } = useActionIconTrigger(
    <ActionIcon name="arrow-up-down" size={16} />
  );

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
        onPointerDownCapture={onParentPointerDown}
        className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[var(--tg-button-color,var(--color-border-default,rgba(0,0,0,.16)))] bg-[var(--tg-secondary-bg-color,var(--color-background-surface,#fff))] text-[var(--tg-button-color,var(--color-text-primary,#1b1b1b))]"
        title={active?.label ?? "Сортировка"}
      >
        {wiredIcon}
      </button>
      {open ? (
        <div className="absolute right-0 top-[calc(100%+6px)] z-20 min-w-[180px] overflow-hidden rounded-[12px] border border-[var(--color-border-default,rgba(0,0,0,.08))] bg-[var(--color-background-surface,var(--tg-secondary-bg-color,#fff))] py-1 shadow-lg">
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
