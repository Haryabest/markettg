"use client";
import { Button, EmptyState, Grid, Skeleton } from "@/components/ui";

import { ProductCard } from "@/components/product-card";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon } from "@/components/icons";
import type { Product } from "@/lib/api";

export function ProductGrid({
  products,
  isLoading,
  isError,
  onRetry,
  emptyTitle = "Товаров пока нет",
  emptyDescription = "Загляните позже или откройте другую категорию.",
}: {
  products?: Product[];
  isLoading?: boolean;
  isError?: boolean;
  onRetry?: () => void;
  emptyTitle?: string;
  emptyDescription?: string;
}) {
  if (isLoading) {
    return (
      <Grid columns={2} gap={3}>
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="overflow-hidden rounded-[var(--radius-container,12px)] border border-white/5">
            <Skeleton height={140} index={i} />
            <div className="space-y-2 p-3">
              <Skeleton height={16} width="85%" index={i + 1} />
              <Skeleton height={12} width="65%" index={i + 2} />
              <Skeleton height={20} width="45%" index={i + 3} />
            </div>
          </div>
        ))}
      </Grid>
    );
  }

  if (isError) {
    return (
      <EmptyState
        title="Не удалось загрузить товары"
        description="Проверьте сеть. Если открываете из Telegram, API должен идти через тот же адрес, что и магазин."
        actions={
          onRetry ? (
            <Button
              label="Повторить"
              variant="primary"
              icon={<ActionIcon name="refresh-cw" size={18} />}
              clickAction={onRetry}
            />
          ) : undefined
        }
      />
    );
  }

  if (!products?.length) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  return (
    <Grid columns={2} gap={3}>
      {products.map((product) => (
        <ProductCard key={product.id} product={product} />
      ))}
    </Grid>
  );
}
