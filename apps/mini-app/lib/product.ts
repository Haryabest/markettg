import type { BadgeVariant } from "@/components/ui";
import type { LucideIcon } from "lucide-react";
import type { Product } from "@/lib/api";
import { formatPrice } from "@/lib/utils";
import { productTypeIcon } from "@/components/icons";

export function productGlyph(type: string): string {
  if (type === "STARS") return "★";
  if (type === "PREMIUM") return "✦";
  return "❀";
}

export function productBadge(type: string): { label: string; variant: BadgeVariant } {
  if (type === "STARS") return { label: "Stars", variant: "yellow" };
  if (type === "PREMIUM") return { label: "Premium", variant: "purple" };
  if (type === "NFT") return { label: "NFT", variant: "blue" };
  return { label: "Gift", variant: "pink" };
}

export function productEmoji(product: Product): string {
  return productGlyph(product.product_type);
}

export function collectibleStarLabel(product: Product): string | null {
  const stars = product.delivery_config?.star_count;
  if ((product.product_type === "GIFT" || product.product_type === "NFT") && stars) {
    return `${stars} ⭐`;
  }
  return null;
}

export function giftStarLabel(product: Product): string | null {
  return collectibleStarLabel(product);
}

export function effectivePrice(product: Product): number {
  if (product.on_sale && product.sale_price_kopecks) {
    return product.sale_price_kopecks;
  }
  return product.price_kopecks;
}

export function productIcon(product: Product): LucideIcon {
  return productTypeIcon(product.product_type, product.delivery_config?.gift_id);
}

export function productFacts(product: Product): string[] {
  const facts: string[] = [];
  const amount = product.delivery_config?.amount;
  const days = product.delivery_config?.duration_days;
  const price = effectivePrice(product);

  if (product.product_type === "STARS" && amount) {
    facts.push(`${amount} Stars на ваш Telegram`);
    facts.push(`${Math.max(1, Math.round(price / amount))} ₽ за звезду`);
  }
  if (product.product_type === "PREMIUM" && days) {
    const months = Math.round(days / 30);
    facts.push(months === 1 ? "Подписка на 1 месяц" : `Подписка на ${months} месяцев`);
    facts.push(`${formatPrice(Math.round(price / Math.max(1, months)))} в месяц`);
  }
  if (product.product_type === "GIFT" || product.product_type === "NFT") {
    const stars = product.delivery_config?.star_count;
    if (stars) {
      const label = product.product_type === "NFT" ? "Коллекционный NFT Gift" : "Официальный Telegram Gift";
      facts.push(`${label} · ${stars} Stars`);
      if (product.delivery_config?.gift_num) {
        facts.push(`Уникальный номер #${product.delivery_config.gift_num}`);
      }
      if (product.delivery_config?.remaining_count != null && product.delivery_config?.total_count != null) {
        facts.push(
          `Лимит: ${product.delivery_config.remaining_count} из ${product.delivery_config.total_count}`
        );
      }
    } else {
      facts.push(product.product_type === "NFT" ? "Коллекционный NFT Gift" : "Анимированный Telegram Gift");
    }
    facts.push("Можно отправить в любой чат");
  }
  facts.push("Доставка сразу после оплаты");
  facts.push("Официальные Stars и Premium");
  return facts;
}

export function productShortHint(product: Product): string {
  const amount = product.delivery_config?.amount;
  const days = product.delivery_config?.duration_days;
  if (product.product_type === "STARS" && amount) return `${amount} Stars · мгновенно`;
  if (product.product_type === "PREMIUM" && days) {
    const months = Math.round(days / 30);
    return months === 1 ? "1 месяц Premium" : `${months} месяцев Premium`;
  }
  if (product.product_type === "GIFT" || product.product_type === "NFT") {
    const stars = product.delivery_config?.star_count;
    if (stars) {
      return product.product_type === "NFT"
        ? `${stars} ⭐ · NFT #${product.delivery_config?.gift_num || "—"}`
        : `${stars} ⭐ · Telegram Gift`;
    }
  }
  return "Telegram Gift · в чат сразу";
}
