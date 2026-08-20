export function supportUrl(): string {
  return process.env.NEXT_PUBLIC_SUPPORT_URL || "https://t.me/markettg_support";
}

export function openSupport() {
  if (typeof window !== "undefined") {
    window.open(supportUrl(), "_blank", "noopener,noreferrer");
  }
}
