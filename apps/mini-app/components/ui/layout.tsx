import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

const GAP: Record<number, string> = {
  0: "gap-0",
  1: "gap-1",
  1.5: "gap-1.5",
  2: "gap-2",
  3: "gap-3",
  4: "gap-4",
  5: "gap-5",
  6: "gap-6",
};

function gapClass(gap?: number) {
  return GAP[gap ?? 0] ?? `gap-${gap}`;
}

export function VStack({
  children,
  gap,
  align,
  className,
  width,
}: {
  children: ReactNode;
  gap?: number;
  align?: "start" | "center" | "end";
  className?: string;
  width?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-col",
        gapClass(gap),
        align === "center" && "items-center",
        align === "end" && "items-end",
        align === "start" && "items-start",
        width === "100%" && "w-full",
        className
      )}
    >
      {children}
    </div>
  );
}

export function HStack({
  children,
  gap,
  align,
  justify,
  width,
  className,
}: {
  children: ReactNode;
  gap?: number;
  align?: "start" | "center" | "end";
  justify?: "start" | "center" | "end" | "between";
  width?: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-row",
        gapClass(gap),
        align === "center" && "items-center",
        align === "end" && "items-end",
        align === "start" && "items-start",
        justify === "between" && "justify-between",
        justify === "center" && "justify-center",
        justify === "end" && "justify-end",
        width === "100%" && "w-full",
        className
      )}
    >
      {children}
    </div>
  );
}

export function Grid({
  children,
  columns = 2,
  gap = 3,
  className,
}: {
  children: ReactNode;
  columns?: number;
  gap?: number;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "grid",
        columns === 2 && "grid-cols-2",
        columns === 3 && "grid-cols-3",
        columns === 4 && "grid-cols-4",
        gapClass(gap),
        className
      )}
    >
      {children}
    </div>
  );
}

export function Heading({ level = 1, children }: { level?: 1 | 2 | 3 | 4; children: ReactNode }) {
  const Tag = (`h${level}` as "h1") || "h1";
  const sizes = {
    1: "text-2xl font-bold tracking-tight",
    2: "text-xl font-semibold",
    3: "text-lg font-semibold",
    4: "text-base font-semibold",
  };
  return <Tag className={sizes[level]}>{children}</Tag>;
}

export function Text({
  children,
  type = "body",
  color,
  weight,
  display,
  justify,
  maxLines,
  className,
}: {
  children: ReactNode;
  type?: "body" | "supporting" | "label" | "large";
  color?: "primary" | "secondary" | "accent";
  weight?: "normal" | "medium" | "semibold";
  display?: "block" | "inline";
  justify?: "center";
  maxLines?: number;
  className?: string;
}) {
  return (
    <span
      className={cn(
        display === "block" && "block",
        display === "inline" && "inline",
        type === "supporting" && "text-sm text-default-500",
        type === "label" && "text-sm font-medium",
        type === "large" && "text-lg",
        type === "body" && "text-base",
        color === "secondary" && "text-default-500",
        color === "accent" && "text-primary",
        weight === "medium" && "font-medium",
        weight === "semibold" && "font-semibold",
        justify === "center" && "text-center",
        maxLines === 2 && "line-clamp-2",
        className
      )}
    >
      {children}
    </span>
  );
}
