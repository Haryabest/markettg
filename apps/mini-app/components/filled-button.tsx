"use client";

import Link from "next/link";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

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
  "bg-[#e8e8e6] text-[#1b1b1b] hover:bg-[#d1d1cb] hover:text-[#1b1b1b] dark:bg-[#52525b] dark:text-[#fafafa] dark:hover:bg-[#71717a] dark:hover:text-[#ffffff]";

const activeStyles =
  "bg-[#007aff] text-white hover:bg-[#0066d6] hover:text-white dark:bg-[#0a84ff] dark:hover:bg-[#409cff] dark:hover:text-white";

const menuIdleStyles =
  "bg-[#ececea] text-[#1b1b1b] hover:bg-[#deded9] hover:text-[#1b1b1b] dark:bg-[#52525b] dark:text-[#fafafa] dark:hover:bg-[#71717a] dark:hover:text-[#ffffff]";

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
      {icon ? <span className="shrink-0 text-inherit [&_svg]:text-inherit">{icon}</span> : null}
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

  const classes = cn(
    "inline-flex items-center gap-2 rounded-[var(--radius-container,12px)] border-0 font-medium no-underline transition-colors",
    size === "sm" ? "px-3 py-2 text-sm" : "px-4 py-3.5 text-base",
    isMenu ? "items-start justify-start" : "items-center justify-center",
    active ? activeStyles : isMenu ? menuIdleStyles : idleStyles,
    fullWidth && "w-full",
    className
  );

  const content = <FilledButtonContent label={label} description={description} icon={icon} />;

  if (href) {
    return (
      <Link href={href} className={classes}>
        {content}
      </Link>
    );
  }

  return (
    <button type="button" onClick={onClick} className={classes}>
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
