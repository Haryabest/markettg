"use client";
import { Badge, Button, Card, Heading, Item, List, ListItem, Spinner, StatusDot, Text, VStack } from "@/components/ui";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { BackHeader } from "@/components/back-header";
import { api } from "@/lib/api";
import { formatDate, formatPrice } from "@/lib/utils";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon, productTypeIcon } from "@/components/icons";
import { Clock } from "lucide-react";
import { openTelegramInvoice } from "@/lib/telegram";
import { notifyError, notifySuccess } from "@/stores/banners";

const STATUS_LABELS: Record<string, string> = {
  CREATED: "Создан",
  PAYMENT_PENDING: "Ожидает оплаты",
  PAID: "Оплачен, готовим доставку",
  DELIVERY_PENDING: "Оплачен, готовим доставку",
  DELIVERING: "Доставляем на аккаунт",
  COMPLETED: "Доставлен",
  CANCELLED: "Отменён",
};

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [paying, setPaying] = useState(false);

  const { data: order, isError, refetch } = useQuery({
    queryKey: ["order", id],
    queryFn: () => api.getOrder(id),
    enabled: !!id,
    refetchInterval: 10_000,
  });

  const payWithStars = async () => {
    if (!order) return;
    setPaying(true);
    try {
      const payment = await api.createPayment(order.id, "STARS");
      if (!payment.payment_url) {
        notifyError("Оплата недоступна", "Не удалось создать счёт Stars");
        return;
      }
      const status = await openTelegramInvoice(payment.payment_url);
      if (status === null) {
        window.open(payment.payment_url, "_blank");
        notifySuccess("Счёт открыт", "Завершите оплату в Telegram");
        return;
      }
      if (status === "paid") {
        notifySuccess("Оплата получена", "Готовим доставку подарка");
        await refetch();
        return;
      }
      if (status === "cancelled") {
        notifyError("Оплата отменена", "Можно попробовать снова");
        return;
      }
      notifyError("Оплата не прошла", "Попробуйте ещё раз");
    } catch (e) {
      notifyError("Ошибка оплаты", e instanceof Error ? e.message : "Не удалось оплатить");
    } finally {
      setPaying(false);
    }
  };

  const canPay = order?.status === "PAYMENT_PENDING" || order?.status === "CREATED";

  if (!order && !isError) return <Spinner label="Загрузка заказа" />;
  if (!order) {
    return (
      <VStack gap={4}>
        <BackHeader fallbackHref="/orders" />
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
      <BackHeader fallbackHref="/orders" />
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
          startContent={<ActionIcon name="wallet" size={16} />}
          label="К оплате"
          description={order.discount_kopecks ? "Уже с учётом промокода" : "Без скидки"}
          endContent={<Text type="body" weight="medium">{formatPrice(order.total_kopecks)}</Text>}
        />
      </List>
      {canPay ? (
        <Button
          label={paying ? "Открываем оплату…" : "Оплатить Stars"}
          icon={<ActionIcon name="star" size={18} />}
          variant="primary"
          fullWidth
          isDisabled={paying}
          clickAction={payWithStars}
        />
      ) : null}
    </VStack>
  );
}
