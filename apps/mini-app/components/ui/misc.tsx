"use client";

import NextLink from "next/link";
import type { ReactNode } from "react";
import { Avatar as HeroAvatar, Card, Skeleton as HeroSkeleton, Spinner as HeroSpinner } from "@heroui/react";
import { cn } from "@/lib/utils";

export function Skeleton({
  height,
  width,
  className,
}: {
  height?: number | string;
  width?: number | string;
  index?: number;
  className?: string;
}) {
  const style: React.CSSProperties = {};
  if (height) style.height = typeof height === "number" ? `${height}px` : height;
  if (width) style.width = typeof width === "number" ? `${width}px` : width;
  if (width === "85%") style.width = "85%";
  if (width === "65%") style.width = "65%";
  if (width === "45%") style.width = "45%";
  return <HeroSkeleton className={cn("rounded-lg", className)} style={style} />;
}

export function Spinner({ label }: { label?: string }) {
  return (
    <div className="flex flex-col items-center gap-2 py-8">
      <HeroSpinner />
      {label ? <span className="text-sm text-default-500">{label}</span> : null}
    </div>
  );
}

export function Avatar({ name, src }: { name: string; src?: string }) {
  const initials = name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
  return (
    <HeroAvatar>
      {src ? <HeroAvatar.Image src={src} alt={name} /> : null}
      <HeroAvatar.Fallback>{initials}</HeroAvatar.Fallback>
    </HeroAvatar>
  );
}

export function CardWrapper({
  children,
  padding,
  elevation,
  className,
}: {
  children: ReactNode;
  padding?: number;
  elevation?: string;
  className?: string;
}) {
  return (
    <Card className={cn(padding === 4 && "p-4", elevation === "low" && "shadow-sm", className)}>
      {children}
    </Card>
  );
}

export { CardWrapper as Card };

export function Item({
  startContent,
  label,
  description,
}: {
  startContent?: ReactNode;
  label: string;
  description?: string;
}) {
  return (
    <div className="flex items-start gap-3">
      {startContent}
      <div className="min-w-0">
        <div className="font-medium">{label}</div>
        {description ? <div className="text-sm text-default-500">{description}</div> : null}
      </div>
    </div>
  );
}

export function Link({
  href,
  children,
  isStandalone,
}: {
  href: string;
  children: ReactNode;
  isStandalone?: boolean;
}) {
  return (
    <NextLink
      href={href}
      className={cn("text-primary hover:underline", isStandalone && "text-sm font-medium")}
    >
      {children}
    </NextLink>
  );
}

export function ClickableCard({
  label,
  href,
  padding = 0,
  elevation,
  className,
  children,
}: {
  label?: string;
  href: string;
  padding?: number;
  elevation?: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <NextLink
      href={href}
      aria-label={label}
      className={cn(
        "block overflow-hidden rounded-2xl transition-transform hover:scale-[1.01]",
        elevation === "low" && "shadow-sm",
        padding === 0 && "p-0",
        className
      )}
    >
      {children}
    </NextLink>
  );
}

export function StatusDot({
  variant,
  label,
  isPulsing,
}: {
  variant?: string;
  label?: string;
  isPulsing?: boolean;
}) {
  return (
    <span
      className={cn(
        "inline-block h-2.5 w-2.5 rounded-full bg-primary",
        isPulsing && "animate-pulse",
        variant === "accent" && "bg-primary"
      )}
      title={label}
    />
  );
}
