"use client";

import { useCallback, useEffect } from "react";
import { useRouter } from "next/navigation";
import { FilledButton } from "@/components/filled-button";
import { ActionIcon } from "@/components/action-icon";

type BackHeaderProps = {
  fallbackHref?: string;
  label?: string;
};

export function BackHeader({ fallbackHref = "/profile", label = "Назад" }: BackHeaderProps) {
  const router = useRouter();

  const goBack = useCallback(() => {
    if (typeof window !== "undefined" && window.history.length > 1) {
      router.back();
      return;
    }
    router.push(fallbackHref);
  }, [router, fallbackHref]);

  useEffect(() => {
    let cleanup: (() => void) | undefined;

    import("@twa-dev/sdk")
      .then(({ default: WebApp }) => {
        if (!WebApp.BackButton) return;
        WebApp.BackButton.show();
        WebApp.BackButton.onClick(goBack);
        cleanup = () => {
          WebApp.BackButton.offClick(goBack);
          WebApp.BackButton.hide();
        };
      })
      .catch(() => {});

    return () => cleanup?.();
  }, [goBack]);

  return (
    <FilledButton
      label={label}
      icon={<ActionIcon name="arrow-left" size={18} />}
      size="md"
      onClick={goBack}
    />
  );
}
