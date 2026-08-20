import type { LucideIcon } from "lucide-react";
import {
  Cake,
  CreditCard,
  Crown,
  Flower2,
  Gem,
  Gift,
  Heart,
  Home,
  LayoutGrid,
  Package,
  Rocket,
  Search,
  ShoppingBag,
  Sparkles,
  Star,
  Trophy,
  UserRound,
  Wine,
} from "lucide-react";

export function AppIcon({
  icon: Icon,
  size = 18,
}: {
  icon: LucideIcon;
  size?: number;
}) {
  return <Icon size={size} strokeWidth={2} aria-hidden />;
}

export const navIcons = {
  home: Home,
  catalog: LayoutGrid,
  search: Search,
  cart: ShoppingBag,
  profile: UserRound,
};

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
