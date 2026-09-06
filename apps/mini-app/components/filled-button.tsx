"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useActionIconTrigger } from "@/lib/action-icon-trigger";

type FilledButtonVariant = "chip" | "menu";

type FilledButtonProps = {
  label: string;
  description?: string;
  active?: boolean;
  icon?: ReactNode;
  onClick?: () => void;
  href?: string;
  className?: string;
  size?: "sm" | "md";
  fullWidth?: boolean;
  variant?: FilledButtonVariant;
};

const idleStyles =
  "bg-[var(--color-background-surface,var(--tg-secondary-bg-color,#e8e8e6))] text-[var(--tg-text-color,var(--foreground,#1b1b1b))]";

const activeStyles =
  "bg-[var(--tg-button-color,#007aff)] text-[var(--tg-button-text-color,#fff)]";

const menuIdleStyles =
  "bg-[var(--color-background-surface,var(--tg-secondary-bg-color,#ececea))] text-[var(--tg-text-color,var(--foreground,#1b1b1b))]";

function FilledButtonContent({
  label,
  description,
  icon,
}: {
  label: string;
  description?: string;
  icon?: ReactNode;
}) {
  return (
    <>
      {icon ? (
        <span className="pointer-events-none shrink-0 text-inherit [&_svg]:text-inherit">{icon}</span>
      ) : null}
      <span className={cn(description ? "flex min-w-0 flex-1 flex-col items-start gap-0.5 text-left" : "")}>
        <span className="leading-tight">{label}</span>
        {description ? (
          <span className="text-xs font-normal leading-snug text-inherit opacity-80">{description}</span>
        ) : null}
      </span>
    </>
  );
}

export function FilledButton({
  label,
  description,
  active = false,
  icon,
  onClick,
  href,
  className,
  size = "sm",
  fullWidth = false,
  variant = "chip",
}: FilledButtonProps) {
  const isMenu = variant === "menu" || Boolean(description);
  const { icon: wiredIcon, onParentPointerDown } = useActionIconTrigger(icon);

  const classes = cn(
    "inline-flex items-center gap-2 rounded-[var(--radius-container,12px)] border-0 font-medium no-underline transition-colors",
    size === "sm" ? "px-3 py-2 text-sm" : "px-4 py-3.5 text-base",
    isMenu ? "items-start justify-start" : "items-center justify-center",
    active ? activeStyles : isMenu ? menuIdleStyles : idleStyles,
    fullWidth && "w-full",
    className
  );

  const content = <FilledButtonContent label={label} description={description} icon={wiredIcon} />;

  if (href) {
    return (
      <Link href={href} className={classes} onPointerDownCapture={onParentPointerDown}>
        {content}
      </Link>
    );
  }

  return (
    <button
      type="button"
      onPointerDownCapture={onParentPointerDown}
      onClick={onClick}
      className={classes}
    >
      {content}
    </button>
  );
}

type FilledButtonGroupProps = {
  children: ReactNode;
  className?: string;
};

export function FilledButtonGroup({ children, className }: FilledButtonGroupProps) {
  return <div className={cn("flex flex-wrap gap-2", className)}>{children}</div>;
}
