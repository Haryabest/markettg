"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { HStack } from "@astryxdesign/core/HStack";
import { List, ListItem } from "@astryxdesign/core/List";
import { Text } from "@astryxdesign/core/Text";
import { TextInput } from "@astryxdesign/core/TextInput";
import { VStack } from "@astryxdesign/core/VStack";
import { api } from "@/lib/api";
import { applyPromoCode, promoSuccessMessage, type AppliedPromo } from "@/lib/promo";
import { formatPrice } from "@/lib/utils";
import { AppIcon } from "@/components/icons";
import { Banknote, ArrowRight, Star, Ticket, Wallet } from "lucide-react";
import { notifyError, notifySuccess } from "@/stores/banners";
import { useCartStore } from "@/stores/app";

const PAYMENT_METHODS = [
  {
    value: "STARS" as const,
    label: "Telegram Stars",
    description: "Оплата внутри Telegram, без банковской карты",
    icon: Star,
  },
  {
    value: "SBP" as const,
    label: "СБП",
    description: "Ссылка на оплату через систему быстрых платежей",
    icon: Banknote,
  },
];

function PaymentRadio({ selected }: { selected: boolean }) {
  return (
    <span
      className={`inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 transition-colors ${
        selected
          ? "border-[var(--color-accent,#2563eb)]"
          : "border-[var(--color-border-strong,#a1a1aa)]"
      }`}
      aria-hidden
    >
      {selected ? (
        <span className="h-2.5 w-2.5 rounded-full bg-[var(--color-accent,#2563eb)]" />
      ) : null}
    </span>
  );
}

export default function CheckoutPage() {
  const router = useRouter();
  const cart = useCartStore((s) => s.cart);
  const [promoInput, setPromoInput] = useState("");
  const [appliedPromo, setAppliedPromo] = useState<AppliedPromo | null>(null);
  const [method, setMethod] = useState<"STARS" | "SBP">("STARS");

  const items = cart?.items || [];
  const { data: referrals } = useQuery({
    queryKey: ["referrals"],
    queryFn: () => api.getReferrals(),
    retry: false,
  });
  const { data: products } = useQuery({
    queryKey: ["checkout-products", items.map((i) => i.product_id).join(",")],
    queryFn: async () => {
      const loaded = await Promise.all(items.map((item) => api.getProduct(item.product_id)));
      return loaded;
    },
    enabled: items.length > 0,
  });

  const subtotal = items.reduce((sum, item, index) => {
    const product = products?.[index];
    return sum + (product?.price_kopecks || 0) * item.quantity;
  }, 0);
  const discount = appliedPromo?.discountKopecks || 0;
  const total = Math.max(subtotal - discount, 0);

  const handleApplyPromo = () => {
    const promo = applyPromoCode(promoInput, subtotal, referrals?.rewards);
    if (!promo) {
      setAppliedPromo(null);
      notifyError("Промокод не найден", "Проверьте код, WELCOME10 или промокод из раздела «Пригласи друга»");
      return;
    }
    setAppliedPromo(promo);
    setPromoInput(promo.code);
    notifySuccess("Промокод применён", promoSuccessMessage(promo));
  };

  const handleCheckout = async () => {
    try {
      const order = await api.createOrder(appliedPromo?.code || promoInput.trim() || undefined);
      const payment = await api.createPayment(order.id, method);
      if (payment.payment_url) {
        window.open(payment.payment_url, "_blank");
      }
      notifySuccess("Заказ создан", "Переходим к статусу оплаты");
      router.push(`/orders/${order.id}`);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Ошибка оформления";
      notifyError("Не удалось оформить", message);
      throw e;
    }
  };

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>Оформление</Heading>
        <Text type="supporting" color="secondary" display="block">
          Проверьте состав, примените промокод и выберите оплату. Доставка начнётся сразу после успешного платежа.
        </Text>
      </VStack>

      {products && products.length > 0 && (
        <List hasDividers header="Ваш заказ">
          {items.map((item, index) => {
            const product = products[index];
            return (
              <ListItem
                key={item.product_id}
                label={product?.name || "Товар"}
                description={`${item.quantity} шт.${product?.description ? ` · ${product.description}` : ""}`}
                endContent={
                  <Text type="body" weight="medium">
                    {product ? formatPrice(product.price_kopecks * item.quantity) : "—"}
                  </Text>
                }
              />
            );
          })}
          <ListItem
            label="Сумма"
            description={appliedPromo ? `Промокод ${appliedPromo.code} применён` : "Без промокода"}
            endContent={
              <Text type="body" weight="medium">
                {formatPrice(subtotal)}
              </Text>
            }
          />
          {appliedPromo ? (
            <ListItem
              label="Скидка"
              description={
                appliedPromo.discountType === "PERCENT"
                  ? `${appliedPromo.discountValue}% по промокоду`
                  : "Фиксированная скидка"
              }
              endContent={
                <Text type="body" weight="medium">
                  −{formatPrice(discount)}
                </Text>
              }
            />
          ) : null}
          <ListItem
            label="К оплате"
            endContent={
              <Text type="large" weight="semibold">
                {formatPrice(total)}
              </Text>
            }
          />
        </List>
      )}

      <VStack gap={1} width="100%">
        <Text type="label" weight="medium" display="block">
          Промокод <span className="opacity-60">· необязательно</span>
        </Text>
        <HStack gap={2} align="center" width="100%">
          <TextInput
            label="Промокод"
            isLabelHidden
            value={promoInput}
            onChange={(value) => {
              setPromoInput(value);
              if (appliedPromo && value.trim().toUpperCase() !== appliedPromo.code) {
                setAppliedPromo(null);
              }
            }}
            placeholder="WELCOME10"
            width="100%"
            startIcon={<AppIcon icon={Ticket} size={16} />}
          />
          <button
            type="button"
            onClick={handleApplyPromo}
            disabled={!promoInput.trim()}
            aria-label="Применить промокод"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center self-center rounded-lg bg-[#007aff] text-white transition-colors hover:bg-[#0066d6] disabled:cursor-not-allowed disabled:opacity-40 dark:bg-[#0a84ff] dark:hover:bg-[#409cff]"
          >
            <AppIcon icon={ArrowRight} size={16} />
          </button>
        </HStack>
        <Text type="supporting" color="secondary" display="block">
          {appliedPromo
            ? `Применён ${appliedPromo.code}: −${formatPrice(appliedPromo.discountKopecks)}`
            : "WELCOME10 — 10% · реферальные FRIEND* и REF* — в профиле"}
        </Text>
      </VStack>
      <VStack gap={2}>
        <Text type="label" weight="semibold" display="block">
          Способ оплаты
        </Text>
        <List hasDividers>
          {PAYMENT_METHODS.map((option) => (
            <ListItem
              key={option.value}
              label={option.label}
              description={option.description}
              startContent={<AppIcon icon={option.icon} size={20} />}
              endContent={<PaymentRadio selected={method === option.value} />}
              onClick={() => setMethod(option.value)}
            />
          ))}
        </List>
      </VStack>
      <Button
        label="Оплатить"
        variant="primary"
        size="lg"
        width="100%"
        icon={<AppIcon icon={Wallet} />}
        clickAction={handleCheckout}
      />
    </VStack>
  );
}
