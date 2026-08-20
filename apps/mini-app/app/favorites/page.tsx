"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { EmptyState } from "@astryxdesign/core/EmptyState";
import { Heading } from "@astryxdesign/core/Heading";
import { Text } from "@astryxdesign/core/Text";
import { VStack } from "@astryxdesign/core/VStack";
import { Button } from "@astryxdesign/core/Button";
import { api } from "@/lib/api";
import { ProductGrid } from "@/components/product-grid";
import { AppIcon } from "@/components/icons";
import { ShoppingBag } from "lucide-react";

export default function FavoritesPage() {
  const queryClient = useQueryClient();
  const { data: favs, isLoading: favsLoading, isError } = useQuery({
    queryKey: ["favorites"],
    queryFn: () => api.getFavorites(),
  });

  const { data: products, isLoading: productsLoading } = useQuery({
    queryKey: ["fav-products", favs?.product_ids?.join(",") || ""],
    queryFn: async () => {
      if (!favs?.product_ids?.length) return { items: [] };
      const results = await Promise.all(favs.product_ids.map((id) => api.getProduct(id)));
      return { items: results };
    },
    enabled: !!favs,
  });

  const empty = !favsLoading && !productsLoading && !products?.items?.length;

  return (
    <VStack gap={4}>
      <VStack gap={1}>
        <Heading level={1}>Избранное</Heading>
        <Text type="supporting" color="secondary" display="block">
          Сохраняйте пакеты Stars, Premium и подарки, чтобы быстро купить их позже.
        </Text>
      </VStack>
      {empty ? (
        <EmptyState
          title="Пока пусто"
          description="Нажмите «В избранное» на карточке товара в каталоге."
          actions={
            <Button
              label="В каталог"
              href="/catalog"
              variant="primary"
              icon={<AppIcon icon={ShoppingBag} />}
            />
          }
        />
      ) : (
        <ProductGrid
          products={products?.items}
          isLoading={favsLoading || productsLoading}
          isError={isError}
          onRetry={() => queryClient.invalidateQueries({ queryKey: ["favorites"] })}
          emptyTitle="Пока пусто"
          emptyDescription="Добавьте товары сердцем на странице товара."
        />
      )}
    </VStack>
  );
}
