"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { Badge } from "@astryxdesign/core/Badge";
import { Button } from "@astryxdesign/core/Button";
import { Card } from "@astryxdesign/core/Card";
import { Heading } from "@astryxdesign/core/Heading";
import { Item } from "@astryxdesign/core/Item";
import { List, ListItem } from "@astryxdesign/core/List";
import { Spinner } from "@astryxdesign/core/Spinner";
import { StatusDot } from "@astryxdesign/core/StatusDot";
import { Text } from "@astryxdesign/core/Text";
import { VStack } from "@astryxdesign/core/VStack";
import { api } from "@/lib/api";
import { formatDate, formatPrice } from "@/lib/utils";
import { AppIcon, productTypeIcon } from "@/components/icons";
import { Clock, Wallet } from "lucide-react";

const STATUS_LABELS: Record<string, string> = {
  CREATED: "Создан",
  PAYMENT_PENDING: "Ожидает оплаты",
  PAID: "Оплачен, готовим доставку",
  DELIVERING: "Доставляем на аккаунт",
  COMPLETED: "Доставлен",
  CANCELLED: "Отменён",
};

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();

  const { data: order, isError } = useQuery({
    queryKey: ["order", id],
    queryFn: () => api.getOrder(id),
    enabled: !!id,
    refetchInterval: 10_000,
  });

  if (!order && !isError) return <Spinner label="Загрузка заказа" />;
  if (!order) {
    return (
      <VStack gap={4}>
        <Heading level={1}>Заказ</Heading>
        <Text type="body" color="secondary" display="block">
          Не удалось открыть заказ. Возможно, нужна авторизация через Telegram.
        </Text>
        <Button label="К заказам" href="/orders" variant="primary" />
      </VStack>
    );
  }

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>Заказ #{order.id.slice(0, 8)}</Heading>
        <Text type="supporting" color="secondary" display="block">
          Создан {formatDate(order.created_at)}. Статус обновляется автоматически.
        </Text>
      </VStack>
      <Card padding={4}>
        <VStack gap={3}>
          <Item
            startContent={<StatusDot variant="accent" label={order.status} isPulsing={order.status !== "COMPLETED" && order.status !== "CANCELLED"} />}
            label={STATUS_LABELS[order.status] || order.status}
            description="Stars и Premium обычно приходят в течение минуты после оплаты"
          />
          <Text type="large" weight="semibold" display="block">
            {formatPrice(order.total_kopecks)}
          </Text>
          {order.discount_kopecks ? (
            <Badge variant="green" label={`Скидка ${formatPrice(order.discount_kopecks)}`} />
          ) : null}
        </VStack>
      </Card>
      {order.items && (
        <List hasDividers header="Состав заказа">
          {order.items.map((item) => (
            <ListItem
              key={item.id}
              startContent={<AppIcon icon={productTypeIcon(item.product_type)} />}
              label={item.name}
              description={`${item.product_type} · ${item.quantity} шт. × ${formatPrice(item.price_kopecks)}`}
              endContent={
                <Text type="body" weight="medium">
                  {formatPrice(item.price_kopecks * item.quantity)}
                </Text>
              }
            />
          ))}
        </List>
      )}
      <List hasDividers header="Детали">
        <ListItem
          startContent={<AppIcon icon={Clock} size={16} />}
          label="Оформлен"
          description={formatDate(order.created_at)}
        />
        <ListItem
          startContent={<AppIcon icon={Wallet} size={16} />}
          label="К оплате"
          description={order.discount_kopecks ? "Уже с учётом промокода" : "Без скидки"}
          endContent={<Text type="body" weight="medium">{formatPrice(order.total_kopecks)}</Text>}
        />
      </List>
    </VStack>
  );
}
