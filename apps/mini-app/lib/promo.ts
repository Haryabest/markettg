import type { ReferralReward } from "@/lib/api";
import { formatPrice } from "@/lib/utils";

export type AppliedPromo = {
  code: string;
  discountType: "PERCENT" | "FIXED";
  discountValue: number;
  discountKopecks: number;
};

const KNOWN_PROMOS: Record<string, { discountType: "PERCENT" | "FIXED"; discountValue: number }> = {
  WELCOME10: { discountType: "PERCENT", discountValue: 10 },
};

function calcDiscount(
  discountType: "PERCENT" | "FIXED",
  discountValue: number,
  subtotalKopecks: number
): number {
  let discountKopecks = 0;
  if (discountType === "PERCENT") {
    discountKopecks = Math.floor((subtotalKopecks * discountValue) / 100);
  } else {
    discountKopecks = discountValue;
  }
  if (discountKopecks > subtotalKopecks) {
    discountKopecks = subtotalKopecks;
  }
  return discountKopecks;
}

export function applyPromoCode(
  code: string,
  subtotalKopecks: number,
  referralRewards?: ReferralReward[]
): AppliedPromo | null {
  const normalized = code.trim().toUpperCase();
  if (!normalized) return null;

  const promo = KNOWN_PROMOS[normalized];
  if (promo) {
    return {
      code: normalized,
      discountType: promo.discountType,
      discountValue: promo.discountValue,
      discountKopecks: calcDiscount(promo.discountType, promo.discountValue, subtotalKopecks),
    };
  }

  const reward = referralRewards?.find(
    (item) => item.promo_code.toUpperCase() === normalized && !item.is_used
  );
  if (!reward) return null;

  const discountType = reward.discount_type === "FIXED" ? "FIXED" : "PERCENT";
  return {
    code: reward.promo_code.toUpperCase(),
    discountType,
    discountValue: reward.discount_value,
    discountKopecks: calcDiscount(discountType, reward.discount_value, subtotalKopecks),
  };
}

export function promoSuccessMessage(promo: AppliedPromo): string {
  if (promo.discountType === "PERCENT") {
    return `Промокод ${promo.code}: скидка ${promo.discountValue}% (−${formatPrice(promo.discountKopecks)})`;
  }
  return `Промокод ${promo.code}: скидка ${formatPrice(promo.discountKopecks)}`;
}
