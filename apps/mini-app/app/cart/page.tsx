"use client";
import { Banner, Button, EmptyState, Heading, HStack, List, ListItem, NumberInput, Text, VStack } from "@/components/ui";

import { useEffect } from "react";
import { useQueries } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";
import { effectivePrice, productShortHint } from "@/lib/product";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon, productTypeIcon } from "@/components/icons";
import { useCartStore } from "@/stores/app";
export default function CartPage() {
  const { cart, fetchCart, addItem, removeItem, isLocal } = useCartStore();

  useEffect(() => {
    fetchCart();
  }, [fetchCart]);

  const items = cart?.items || [];
  const products = useQueries({
    queries: items.map((item) => ({
      queryKey: ["product", item.product_id],
      queryFn: () => api.getProduct(item.product_id),
    })),
  });

  const total = items.reduce((sum, item, index) => {
    const product = products[index]?.data;
    if (!product) return sum;
    return sum + effectivePrice(product) * item.quantity;
  }, 0);
  const quantity = items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <VStack gap={4}>
      <VStack gap={1}>
        <Heading level={1}>Корзина</Heading>
        <Text type="supporting" color="secondary" display="block">
          {items.length
            ? `${quantity} шт. · к оплате ${formatPrice(total)}`
            : "Соберите заказ из Stars, Premium или подарков."}
        </Text>
      </VStack>
      {isLocal ? (
        <Banner
          status="warning"
          title="Локальная корзина"
          description="Для оплаты откройте магазин из Telegram — тогда корзина синхронизируется с аккаунтом."
          isDismissable
        />
      ) : null}
      {items.length === 0 ? (
        <EmptyState
          title="Корзина пуста"
          description="Добавьте Stars, Premium или подарок — доставим сразу после оплаты."
          actions={
            <Button
              label="В каталог"
              href="/catalog"
              variant="primary"
              icon={<ActionIcon name="shopping-bag" size={20} />}
            />
          }
        />
      ) : (
        <>
          <List hasDividers>
            {items.map((item, index) => {
              const product = products[index]?.data;
              const Icon = productTypeIcon(product?.product_type || "GIFT", product?.delivery_config?.gift_id);
              const unit = product ? effectivePrice(product) : 0;
              return (
                <ListItem
                  key={item.product_id}
                  href={product ? `/product/${product.id}` : undefined}
                  startContent={<AppIcon icon={Icon} size={20} />}
                  label={product?.name || `Товар ${item.product_id.slice(0, 8)}`}
                  description={
                    product
                      ? `${productShortHint(product)} · ${formatPrice(unit)} × ${item.quantity} = ${formatPrice(unit * item.quantity)}`
                      : "Загрузка товара…"
                  }
                  endContent={
                    <HStack gap={2} align="center">
                      <NumberInput
                        label="Количество"
                        isLabelHidden
                        value={item.quantity}
                        min={1}
                        max={99}
                        isIntegerOnly
                        hasNumberSteppers
                        width={96}
                        onChange={(value) => addItem(item.product_id, value)}
                      />
                      <Button
                        label="Удалить"
                        variant="ghost"
                        size="sm"
                        isIconOnly
                        icon={<ActionIcon name="trash-2" size={16} />}
                        clickAction={() => removeItem(item.product_id)}
                      />
                    </HStack>
                  }
                />
              );
            })}
          </List>
          <VStack gap={1}>
            <Text type="body" weight="medium" display="block">
              Итого: {formatPrice(total)}
            </Text>
            <Text type="supporting" color="secondary" display="block">
              Промокод WELCOME10 можно ввести на следующем шаге. Оплата Stars или СБП.
            </Text>
          </VStack>
          <Button
            label="Оформить заказ"
            href="/checkout"
            variant="primary"
            size="lg"
            width="100%"
            icon={<ActionIcon name="credit-card" size={20} />}
          />
        </>
      )}
    </VStack>
  );
}
