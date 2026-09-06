"use client";
import { Badge, EmptyState, Heading, List, ListItem, Spinner, Text, VStack, Button } from "@/components/ui";

import { useQuery } from "@tanstack/react-query";
import { BackHeader } from "@/components/back-header";
import { api } from "@/lib/api";
import { formatDate, formatPrice } from "@/lib/utils";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon } from "@/components/icons";

import { Package } from "lucide-react";

const STATUS_LABELS: Record<string, string> = {
  CREATED: "Создан",
  PAYMENT_PENDING: "Ожидает оплаты",
  PAID: "Оплачен",
  DELIVERING: "Доставляется",
  COMPLETED: "Выполнен",
  CANCELLED: "Отменён",
};

const STATUS_VARIANT: Record<string, "neutral" | "info" | "success" | "warning" | "error"> = {
  CREATED: "neutral",
  PAYMENT_PENDING: "warning",
  PAID: "info",
  DELIVERING: "info",
  COMPLETED: "success",
  CANCELLED: "error",
};

export default function OrdersPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["orders"],
    queryFn: () => api.getOrders(),
  });

  return (
    <VStack gap={4}>
      <BackHeader fallbackHref="/profile" />
      <VStack gap={1}>
        <Heading level={1}>Заказы</Heading>
        <Text type="supporting" color="secondary" display="block">
          История покупок Stars, Premium и подарков. Нажмите на заказ, чтобы увидеть состав и статус доставки.
        </Text>
      </VStack>
      {isLoading && <Spinner label="Загрузка заказов" />}
      {isError && (
        <EmptyState
          title="Не удалось загрузить заказы"
          description="Нужен вход через Telegram Mini App."
        />
      )}
      {!isLoading && !isError && !data?.orders?.length && (
        <EmptyState
          title="Заказов пока нет"
          description="Оформите первый заказ в каталоге — Stars и Premium придут сразу после оплаты."
          actions={
            <Button
              label="В каталог"
              href="/catalog"
              variant="primary"
              icon={<ActionIcon name="layout-grid" size={20} />}
            />
          }
        />
      )}
      <List hasDividers>
        {data?.orders?.map((order) => (
          <ListItem
            key={order.id}
            startContent={<AppIcon icon={Package} />}
            label={`Заказ #${order.id.slice(0, 8)}`}
            description={`${formatPrice(order.total_kopecks)}${order.discount_kopecks ? ` · скидка ${formatPrice(order.discount_kopecks)}` : ""} · ${formatDate(order.created_at)}${order.items?.length ? ` · ${order.items.length} позиций` : ""}`}
            href={`/orders/${order.id}`}
            endContent={
              <Badge
                variant={STATUS_VARIANT[order.status] || "neutral"}
                label={STATUS_LABELS[order.status] || order.status}
              />
            }
          />
        ))}
      </List>
    </VStack>
  );
}
