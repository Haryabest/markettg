"use client";

import type { ReactNode } from "react";
import { Alert } from "@heroui/react";

type BannerStatus = "info" | "warning" | "error" | "success";

const STATUS_MAP: Record<BannerStatus, "default" | "warning" | "danger" | "success"> = {
  info: "default",
  warning: "warning",
  error: "danger",
  success: "success",
};

export function Banner({
  status = "info",
  title,
  description,
}: {
  status?: BannerStatus;
  title: string;
  description?: string;
  isDismissable?: boolean;
}) {
  return (
    <Alert status={STATUS_MAP[status]}>
      <Alert.Indicator />
      <Alert.Content>
        <Alert.Title>{title}</Alert.Title>
        {description ? <Alert.Description>{description}</Alert.Description> : null}
      </Alert.Content>
    </Alert>
  );
}

export function EmptyState({
  title,
  description,
  actions,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-2xl border border-default-200 bg-content1 px-6 py-10 text-center">
      <p className="text-lg font-semibold">{title}</p>
      {description ? <p className="max-w-sm text-sm text-default-500">{description}</p> : null}
      {actions ? <div className="pt-2">{actions}</div> : null}
    </div>
  );
}
