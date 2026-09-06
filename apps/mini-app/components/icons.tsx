import type { LucideIcon } from "lucide-react";
import {
  Cake,
  CreditCard,
  Crown,
  Flower2,
  Gem,
  Gift,
  Heart,
  Package,
  Rocket,
  ShoppingBag,
  Sparkles,
  Star,
  Trophy,
  Wine,
} from "lucide-react";
import type { ActionIconName } from "@/components/action-icon";

export const navActionIcons = {
  home: "home",
  catalog: "layout-grid",
  search: "search",
  cart: "shopping-bag",
  profile: "user-round",
} as const satisfies Record<string, ActionIconName>;

export function AppIcon({
  icon: Icon,
  size = 18,
}: {
  icon: LucideIcon;
  size?: number;
}) {
  return <Icon size={size} strokeWidth={2} aria-hidden />;
}

export function categoryIcon(slug: string): LucideIcon {
  if (slug === "stars") return Star;
  if (slug === "premium") return Crown;
  if (slug === "gifts") return Gift;
  return Sparkles;
}

export function giftIcon(giftId?: string): LucideIcon {
  switch (giftId) {
    case "rose":
    case "bouquet":
      return Flower2;
    case "heart":
      return Heart;
    case "rocket":
      return Rocket;
    case "trophy":
      return Trophy;
    case "cake":
      return Cake;
    case "champagne":
      return Wine;
    case "gem":
    case "ring":
      return Gem;
    default:
      return Gift;
  }
}

export function productTypeIcon(type: string, giftId?: string): LucideIcon {
  if (type === "STARS") return Star;
  if (type === "PREMIUM") return Crown;
  return giftIcon(giftId);
}

export { CreditCard, Heart, Package, ShoppingBag };
