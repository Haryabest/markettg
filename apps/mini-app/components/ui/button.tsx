"use client";

import NextLink from "next/link";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useActionIconTrigger } from "@/lib/action-icon-trigger";

type ButtonVariant = "primary" | "secondary" | "ghost";

const VARIANT_CLASS: Record<ButtonVariant, string> = {
  primary:
    "bg-[var(--tg-button-color,#007aff)] text-[var(--tg-button-text-color,#fff)]",
  secondary:
    "border border-[var(--color-border-default,var(--tg-hint-color,rgba(0,0,0,.16)))] bg-[var(--color-background-surface,var(--tg-secondary-bg-color,transparent))] text-[var(--tg-text-color,var(--foreground,#1b1b1b))]",
  ghost: "bg-transparent text-[var(--tg-link-color,var(--tg-button-color,#007aff))]",
};

const SIZE_CLASS = {
  sm: "h-8 px-3 text-sm",
  md: "h-10 px-4 text-sm",
  lg: "h-12 px-5 text-base",
};

export function Button({
  label,
  href,
  clickAction,
  onPress,
  icon,
  variant = "primary",
  size = "md",
  width,
  fullWidth,
  isIconOnly,
  isDisabled,
  className,
}: {
  label?: string;
  href?: string;
  clickAction?: () => void;
  onPress?: () => void;
  icon?: ReactNode;
  variant?: ButtonVariant;
  size?: "sm" | "md" | "lg";
  width?: string;
  fullWidth?: boolean;
  isIconOnly?: boolean;
  isDisabled?: boolean;
  className?: string;
}) {
  const action = clickAction ?? onPress;
  const { icon: wiredIcon, onParentPointerDown } = useActionIconTrigger(icon);
  const content = (
    <>
      {wiredIcon}
      {!isIconOnly && label ? <span>{label}</span> : null}
    </>
  );

  const classes = cn(
    "inline-flex items-center justify-center gap-2 rounded-xl font-medium no-underline transition-opacity disabled:cursor-not-allowed disabled:opacity-40",
    VARIANT_CLASS[variant],
    isIconOnly ? "h-10 w-10 px-0" : SIZE_CLASS[size],
    (fullWidth || width === "100%") && "w-full",
    className
  );

  if (href) {
    return (
      <NextLink
        href={href}
        className={classes}
        aria-disabled={isDisabled}
        onPointerDownCapture={onParentPointerDown}
      >
        {content}
      </NextLink>
    );
  }

  return (
    <button
      type="button"
      onClick={action}
      onPointerDownCapture={onParentPointerDown}
      disabled={isDisabled}
      className={classes}
    >
      {content}
    </button>
  );
}
