"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { Badge } from "@astryxdesign/core/Badge";
import { Button } from "@astryxdesign/core/Button";
import { EmptyState } from "@astryxdesign/core/EmptyState";
import { Heading } from "@astryxdesign/core/Heading";
import { HStack } from "@astryxdesign/core/HStack";
import { List, ListItem } from "@astryxdesign/core/List";
import { Skeleton } from "@astryxdesign/core/Skeleton";
import { Text } from "@astryxdesign/core/Text";
import { VStack } from "@astryxdesign/core/VStack";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";
import { effectivePrice, giftStarLabel, productBadge, productFacts, productShortHint } from "@/lib/product";
import { AppIcon, productTypeIcon } from "@/components/icons";
import { GiftMedia } from "@/components/gift-media";
import { FilledButton } from "@/components/filled-button";
import { notifyError } from "@/stores/banners";
import { useCartStore } from "@/stores/app";
import { Check, CreditCard, Heart, ShoppingBag, Zap, ArrowLeft } from "lucide-react";

export default function ProductPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const addItem = useCartStore((s) => s.addItem);
  const queryClient = useQueryClient();

  const { data: product, isLoading, isError } = useQuery({
    queryKey: ["product", id],
    queryFn: () => api.getProduct(id),
    enabled: !!id,
  });

  const { data: favs } = useQuery({
    queryKey: ["favorites"],
    queryFn: () => api.getFavorites(),
    retry: false,
  });

  const favorite = Boolean(product && favs?.product_ids?.includes(product.id));

  const toggleFav = useMutation({
    mutationFn: async () => {
      if (!product) return;
      if (favorite) await api.removeFavorite(product.id);
      else await api.addFavorite(product.id);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["favorites"] }),
    onError: (error) => {
      notifyError("Избранное недоступно", error instanceof Error ? error.message : "Нужен вход через Telegram");
    },
  });

  const handleBack = () => {
    if (typeof window !== "undefined" && window.history.length > 1) {
      router.back();
      return;
    }
    router.push("/catalog");
  };

  if (isLoading) return <Skeleton height={420} />;
  if (isError || !product) {
    return (
      <VStack gap={4}>
        <FilledButton
          label="Назад"
          icon={<AppIcon icon={ArrowLeft} size={18} />}
          size="md"
          onClick={handleBack}
        />
        <EmptyState
          title="Товар не найден"
          description="Вернитесь в каталог и выберите другой пакет Stars, Premium или подарок."
          actions={<Button label="В каталог" href="/catalog" variant="primary" />}
        />
      </VStack>
    );
  }

  const badge = productBadge(product.product_type);
  const Icon = productTypeIcon(product.product_type, product.delivery_config?.gift_id);
  const price = effectivePrice(product);
  const hasSale = Boolean(product.on_sale && product.sale_price_kopecks && product.sale_price_kopecks < product.price_kopecks);
  const facts = productFacts(product);

  const handleAddToCart = async () => {
    await addItem(product.id, 1);
  };

  const handleBuyNow = async () => {
    const ok = await addItem(product.id, 1);
    if (!ok) {
      notifyError("Не удалось оформить", "Попробуйте ещё раз или откройте магазин из Telegram");
      return;
    }
    router.push("/checkout");
  };

  return (
    <VStack gap={5}>
      <FilledButton
        label="Назад"
        icon={<AppIcon icon={ArrowLeft} size={18} />}
        size="md"
        onClick={handleBack}
      />
      <div className="flex aspect-square items-center justify-center rounded-[var(--radius-container,16px)] bg-[radial-gradient(circle_at_center,#1a2238_0%,#0d0f14_70%)] p-4">
        <GiftMedia
          name={product.name}
          imageUrl={product.image_url}
          stickerUrl={product.delivery_config?.sticker_url}
          fallbackIcon={Icon}
          animated={product.product_type === "GIFT" || product.product_type === "NFT"}
          className="h-full w-full"
          imageClassName="h-full w-full max-h-[min(78vw,420px)] max-w-[min(78vw,420px)]"
        />
      </div>
      <VStack gap={2}>
        <HStack gap={2} align="center">
          <Badge variant={badge.variant} label={badge.label} />
          {product.categories?.map((cat) => (
            <Badge key={cat.id} variant="neutral" label={cat.name} />
          ))}
          {hasSale ? <Badge variant="red" label="Скидка" /> : null}
        </HStack>
        <Heading level={1}>{product.name}</Heading>
        <Text type="supporting" color="secondary" display="block">
          {productShortHint(product)}
        </Text>
        <HStack gap={2} align="center">
          <Text type="large" weight="semibold" color="accent" display="block">
            {giftStarLabel(product) || formatPrice(price)}
          </Text>
          {giftStarLabel(product) ? (
            <Text type="body" color="secondary" display="block">
              ≈ {formatPrice(price)}
            </Text>
          ) : null}
          {hasSale ? (
            <Text type="body" color="secondary" display="block">
              <s>{formatPrice(product.price_kopecks)}</s>
            </Text>
          ) : null}
        </HStack>
        {product.description && (
          <Text type="body" color="secondary" display="block">
            {product.description}
          </Text>
        )}
      </VStack>
      <List hasDividers header="Что входит">
        {facts.map((fact) => (
          <ListItem
            key={fact}
            startContent={<AppIcon icon={fact.includes("сразу") ? Zap : Check} size={16} />}
            label={fact}
          />
        ))}
      </List>
      <VStack gap={2}>
        <Button
          label={`Купить сейчас · ${formatPrice(price)}`}
          variant="primary"
          size="lg"
          width="100%"
          icon={<AppIcon icon={CreditCard} />}
          clickAction={handleBuyNow}
        />
        <HStack gap={2}>
          <FilledButton
            label={favorite ? "В избранном" : "В избранное"}
            active={favorite}
            size="md"
            icon={<AppIcon icon={Heart} />}
            onClick={() => toggleFav.mutateAsync()}
            className="flex-1"
          />
          <FilledButton
            label="В корзину"
            size="md"
            icon={<AppIcon icon={ShoppingBag} />}
            onClick={handleAddToCart}
            className="flex-1"
          />
        </HStack>
      </VStack>
    </VStack>
  );
}
