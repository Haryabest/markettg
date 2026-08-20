"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Banner } from "@astryxdesign/core/Banner";
import { Heading } from "@astryxdesign/core/Heading";
import { HStack } from "@astryxdesign/core/HStack";
import { Link } from "@astryxdesign/core/Link";
import { Text } from "@astryxdesign/core/Text";
import { VStack } from "@astryxdesign/core/VStack";
import { Grid } from "@astryxdesign/core/Grid";
import { api } from "@/lib/api";
import { AppIcon } from "@/components/icons";
import { ProductGrid } from "@/components/product-grid";
import { Bolt, ShieldCheck, Truck } from "lucide-react";

export default function HomePage() {
  const queryClient = useQueryClient();

  const { data: promos } = useQuery({
    queryKey: ["promotions"],
    queryFn: () => api.getPromotions(),
  });

  const {
    data: popular,
    isLoading: popularLoading,
    isError: popularError,
  } = useQuery({
    queryKey: ["products", "popular"],
    queryFn: () => api.getProducts({ sort: "popularity", limit: "8" }),
  });

  const {
    data: newest,
    isLoading: newestLoading,
    isError: newestError,
  } = useQuery({
    queryKey: ["products", "newest"],
    queryFn: () => api.getProducts({ sort: "newest", limit: "6" }),
  });

  return (
    <VStack gap={6}>
      <VStack gap={1}>
        <Heading level={1}>MarketTG</Heading>
        <Text type="supporting" color="secondary" display="block">
          Магазин Telegram Stars, Premium и подарков. Оплата Stars или СБП, доставка на аккаунт сразу после оплаты.
        </Text>
      </VStack>

      {promos?.items?.map((promo) => (
        <Banner
          key={promo.id}
          status="info"
          title={promo.title}
          description={
            promo.description ||
            (promo.discount_type === "PERCENT"
              ? `Скидка ${promo.discount_value}% по промокоду на оформлении`
              : `Скидка ${promo.discount_value / 100} ₽`)
          }
        />
      ))}

      <Grid columns={3} gap={2}>
        <VStack gap={1} align="center">
          <AppIcon icon={Bolt} size={20} />
          <Text type="label" weight="medium" justify="center" display="block">
            Сразу
          </Text>
          <Text type="supporting" color="secondary" justify="center" display="block">
            После оплаты
          </Text>
        </VStack>
        <VStack gap={1} align="center">
          <AppIcon icon={ShieldCheck} size={20} />
          <Text type="label" weight="medium" justify="center" display="block">
            На аккаунт
          </Text>
          <Text type="supporting" color="secondary" justify="center" display="block">
            Официально
          </Text>
        </VStack>
        <VStack gap={1} align="center">
          <AppIcon icon={Truck} size={20} />
          <Text type="label" weight="medium" justify="center" display="block">
            Без очереди
          </Text>
          <Text type="supporting" color="secondary" justify="center" display="block">
            24/7
          </Text>
        </VStack>
      </Grid>

      <VStack gap={3}>
        <HStack justify="between" align="center">
          <VStack gap={0}>
            <Heading level={2}>Популярное</Heading>
            <Text type="supporting" color="secondary" display="block">
              {popular?.total ? `${popular.total} товаров в каталоге` : "Чаще всего покупают"}
            </Text>
          </VStack>
          <Link href="/catalog" isStandalone>
            Все
          </Link>
        </HStack>
        <ProductGrid
          products={popular?.items}
          isLoading={popularLoading}
          isError={popularError}
          onRetry={() => queryClient.invalidateQueries({ queryKey: ["products"] })}
          emptyTitle="Каталог пуст"
          emptyDescription="Товары появятся после загрузки каталога."
        />
      </VStack>

      <VStack gap={3}>
        <HStack justify="between" align="center">
          <Heading level={2}>Новинки</Heading>
          <Link href="/catalog?sort=newest" isStandalone>
            Каталог
          </Link>
        </HStack>
        <ProductGrid
          products={newest?.items}
          isLoading={newestLoading}
          isError={newestError}
          onRetry={() => queryClient.invalidateQueries({ queryKey: ["products", "newest"] })}
        />
      </VStack>

      <Banner
        status="info"
        title="Как это работает"
        description="Выберите товар → оплатите Stars или СБП → Stars, Premium или Gift придут в Telegram. Поддержка и статус заказа — в профиле."
      />
    </VStack>
  );
}
