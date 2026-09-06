"use client";
import { Badge, ClickableCard, Text, VStack } from "@/components/ui";

import { formatPrice } from "@/lib/utils";
import { collectibleStarLabel, effectivePrice, productBadge } from "@/lib/product";
import { AppIcon, productTypeIcon } from "@/components/icons";
import { GiftMedia } from "@/components/gift-media";
import type { Product } from "@/lib/api";
import { Zap } from "lucide-react";
import { cn } from "@/lib/utils";

function cardVisuals(type: string) {
  if (type === "STARS") {
    return {
      shell: "border-amber-500/20 bg-gradient-to-br from-amber-500/15 via-yellow-500/10 to-orange-600/5",
      media: "bg-gradient-to-br from-amber-400/35 via-yellow-500/20 to-orange-500/10",
      iconWrap: "bg-amber-500/25 text-amber-200 shadow-[0_0_24px_rgba(245,158,11,0.25)]",
      price: "text-amber-300",
    };
  }
  if (type === "PREMIUM") {
    return {
      shell: "border-violet-500/20 bg-gradient-to-br from-violet-500/15 via-purple-500/10 to-fuchsia-600/5",
      media: "bg-gradient-to-br from-violet-500/35 via-purple-500/20 to-fuchsia-500/10",
      iconWrap: "bg-violet-500/25 text-violet-100 shadow-[0_0_24px_rgba(139,92,246,0.25)]",
      price: "text-violet-200",
    };
  }
  if (type === "NFT") {
    return {
      shell: "border-cyan-500/20 bg-[#101014]",
      media: "bg-[radial-gradient(circle_at_center,#1a2238_0%,#0d0f14_70%)]",
      iconWrap: "bg-cyan-500/20 text-cyan-100",
      price: "text-cyan-200",
    };
  }
  return {
    shell: "border-pink-500/20 bg-[#101014]",
    media: "bg-[radial-gradient(circle_at_center,#2a1420_0%,#0d0f14_70%)]",
    iconWrap: "bg-pink-500/20 text-pink-100",
    price: "text-pink-200",
  };
}

export function ProductCard({ product }: { product: Product }) {
  const badge = productBadge(product.product_type);
  const Icon = productTypeIcon(product.product_type, product.delivery_config?.gift_id);
  const price = effectivePrice(product);
  const starLabel = collectibleStarLabel(product);
  const isCollectible = product.product_type === "GIFT" || product.product_type === "NFT";
  const hasPreview = Boolean(product.image_url || product.delivery_config?.sticker_url);
  const hasSale = Boolean(
    product.on_sale && product.sale_price_kopecks && product.sale_price_kopecks < product.price_kopecks
  );
  const visuals = cardVisuals(product.product_type);

  return (
    <ClickableCard
      label={product.name}
      href={`/product/${product.id}`}
      padding={0}
      elevation="low"
      className={cn("overflow-hidden border", visuals.shell)}
    >
      <VStack gap={0}>
        <div
          className={cn(
            "relative flex aspect-square items-center justify-center p-3",
            isCollectible ? visuals.media : cn("aspect-[4/3] p-4", visuals.media)
          )}
        >
          <div className="absolute left-3 top-3">
            <Badge variant={badge.variant} label={badge.label} />
          </div>
          {hasSale ? (
            <div className="absolute right-3 top-3">
              <Badge variant="red" label="−%" />
            </div>
          ) : null}
          {hasPreview ? (
            <GiftMedia
              name={product.name}
              imageUrl={product.image_url}
              stickerUrl={product.delivery_config?.sticker_url}
              fallbackIcon={Icon}
              imageClassName="h-[88%] w-[88%]"
            />
          ) : (
            <div
              className={cn(
                "flex h-24 w-24 items-center justify-center rounded-2xl backdrop-blur-sm",
                visuals.iconWrap
              )}
            >
              <AppIcon icon={Icon} size={44} />
            </div>
          )}
        </div>

        <VStack gap={1.5} className="p-3">
          <Text type="body" weight="semibold" maxLines={2} display="block">
            {product.name}
          </Text>
          <div className="flex items-end justify-between gap-2 pt-1">
            <div>
              <Text className={cn("text-lg font-bold leading-none", visuals.price)} display="block">
                {starLabel || formatPrice(price)}
              </Text>
              {starLabel ? (
                <Text type="supporting" color="secondary" display="block">
                  ≈ {formatPrice(price)}
                </Text>
              ) : null}
              {hasSale ? (
                <Text type="supporting" color="secondary" display="block">
                  <s>{formatPrice(product.price_kopecks)}</s>
                </Text>
              ) : null}
            </div>
            <div className="flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-[11px] text-white/80">
              <AppIcon icon={Zap} size={12} />
              <span>сразу</span>
            </div>
          </div>
        </VStack>
      </VStack>
    </ClickableCard>
  );
}
