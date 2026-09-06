"use client";

import NextLink from "next/link";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useActionIconTrigger } from "@/lib/action-icon-trigger";

export function List({
  children,
  hasDividers,
  header,
}: {
  children: ReactNode;
  hasDividers?: boolean;
  header?: string;
}) {
  return (
    <div className="overflow-hidden rounded-2xl border border-default-200 bg-content1">
      {header ? (
        <div className="border-b border-default-200 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-default-500">
          {header}
        </div>
      ) : null}
      <div className={cn(hasDividers && "divide-y divide-default-200")}>{children}</div>
    </div>
  );
}

export function ListItem({
  label,
  description,
  startContent,
  endContent,
  href,
  onClick,
}: {
  label: string;
  description?: string;
  startContent?: ReactNode;
  endContent?: ReactNode;
  href?: string;
  onClick?: () => void;
}) {
  const { icon: wiredStart, onParentPointerDown } = useActionIconTrigger(startContent);

  const inner = (
    <>
      {wiredStart ? <span className="shrink-0">{wiredStart}</span> : null}
      <span className="min-w-0 flex-1">
        <span className="block font-medium">{label}</span>
        {description ? <span className="mt-0.5 block text-sm text-default-500">{description}</span> : null}
      </span>
      {endContent ? <span className="shrink-0">{endContent}</span> : null}
    </>
  );

  const className =
    "flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-default-100";

  if (href) {
    return (
      <NextLink href={href} className={className} onPointerDownCapture={onParentPointerDown}>
        {inner}
      </NextLink>
    );
  }

  if (onClick) {
    return (
      <button type="button" onClick={onClick} onPointerDownCapture={onParentPointerDown} className={className}>
        {inner}
      </button>
    );
  }

  return <div className={className}>{inner}</div>;
}
