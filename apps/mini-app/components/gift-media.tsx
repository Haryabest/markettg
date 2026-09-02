"use client";

import { useEffect, useRef, useState } from "react";
import lottie, { type AnimationItem } from "lottie-web";
import { inflate } from "pako";
import { AppIcon } from "@/components/icons";
import { cn } from "@/lib/utils";
import type { LucideIcon } from "lucide-react";

type GiftMediaProps = {
  name: string;
  imageUrl?: string;
  stickerUrl?: string;
  fallbackIcon: LucideIcon;
  className?: string;
  imageClassName?: string;
  animated?: boolean;
};

async function loadTgsAnimation(url: string) {
  const res = await fetch(url);
  if (!res.ok) throw new Error("sticker fetch failed");
  const buffer = await res.arrayBuffer();
  const json = inflate(new Uint8Array(buffer), { to: "string" });
  return JSON.parse(json);
}

export function GiftMedia({
  name,
  imageUrl,
  stickerUrl,
  fallbackIcon,
  className,
  imageClassName,
  animated = false,
}: GiftMediaProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const animationRef = useRef<AnimationItem | null>(null);
  const [showFallback, setShowFallback] = useState(!animated || !stickerUrl);

  useEffect(() => {
    if (!animated || !stickerUrl || !containerRef.current) {
      setShowFallback(true);
      return;
    }

    let cancelled = false;
    const container = containerRef.current;

    loadTgsAnimation(stickerUrl)
      .then((data) => {
        if (cancelled || !container) return;
        animationRef.current?.destroy();
        animationRef.current = lottie.loadAnimation({
          container,
          renderer: "svg",
          loop: true,
          autoplay: true,
          animationData: data,
          rendererSettings: {
            preserveAspectRatio: "xMidYMid meet",
            progressiveLoad: true,
          },
        });
        setShowFallback(false);
      })
      .catch(() => {
        if (!cancelled) setShowFallback(true);
      });

    return () => {
      cancelled = true;
      animationRef.current?.destroy();
      animationRef.current = null;
    };
  }, [animated, stickerUrl]);

  if (animated && stickerUrl && !showFallback) {
    return (
      <div
        ref={containerRef}
        className={cn(
          "drop-shadow-[0_16px_36px_rgba(0,0,0,0.45)]",
          className,
          imageClassName
        )}
        aria-label={name}
      />
    );
  }

  if (imageUrl) {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={imageUrl}
        alt={name}
        loading="lazy"
        decoding="async"
        className={cn(
          "object-contain drop-shadow-[0_12px_28px_rgba(0,0,0,0.45)]",
          imageClassName
        )}
      />
    );
  }

  return (
    <div className={cn("flex items-center justify-center", className)}>
      <AppIcon icon={fallbackIcon} size={44} />
    </div>
  );
}
