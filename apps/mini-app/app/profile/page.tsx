"use client";
import { Avatar, Card, Heading, Item, Text, VStack } from "@/components/ui";

import { usePathname } from "next/navigation";
import { FilledButton } from "@/components/filled-button";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon } from "@/components/icons";
import { openSupport } from "@/lib/support";
import { useAuthStore, useCartStore } from "@/stores/app";
import { getDisplayDescription, getDisplayName } from "@/lib/user-display";
import { HelpCircle } from "lucide-react";

const purchaseLinks = [
  {
    href: "/orders",
    label: "Мои заказы",
    description: "Статус оплаты и доставки Stars / Premium / Gift",
    icon: "package",
  },
  {
    href: "/favorites",
    label: "Избранное",
    description: "Сохранённые пакеты и подарки",
    icon: "heart",
  },
  {
    href: "/cart",
    label: "Корзина",
    description: "Товары, ожидающие оплаты",
    icon: "shopping-bag",
  },
] as const;

export default function ProfilePage() {
  const pathname = usePathname();
  const user = useAuthStore((s) => s.user ?? s.telegramUser);
  const isReady = useAuthStore((s) => s.isReady);
  const initData = useAuthStore((s) => s.initData);
  const cartCount = useCartStore((s) => s.itemCount());
  const name = getDisplayName(user, isReady);
  const description = getDisplayDescription(user, isReady, Boolean(initData));

  const isActive = (href: string) => pathname === href || pathname.startsWith(`${href}/`);

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>Профиль</Heading>
        <Text type="supporting" color="secondary" display="block">
          Заказы, избранное и корзина привязаны к вашему Telegram. Данные других аккаунтов недоступны.
        </Text>
      </VStack>

      <Card padding={4} elevation="low">
        <Item
          startContent={<Avatar name={name} src={user?.photo_url} />}
          label={name}
          description={description}
        />
      </Card>

      <VStack gap={2}>
        <Text type="label" weight="semibold" display="block">
          Покупки
        </Text>
        {purchaseLinks.map((item) => (
          <FilledButton
            key={item.href}
            href={item.href}
            label={item.label}
            description={
              item.href === "/cart"
                ? cartCount
                  ? `${cartCount} товаров ждут оплаты`
                  : "Пока пусто — загляните в каталог"
                : item.description
            }
            icon={<ActionIcon name={item.icon} size={20} />}
            active={isActive(item.href)}
            variant="menu"
            size="md"
            fullWidth
          />
        ))}
      </VStack>

      <VStack gap={2}>
        <Text type="label" weight="semibold" display="block">
          Бонусы и помощь
        </Text>
        <FilledButton
          href="/profile/referral"
          label="Пригласи друга"
          description="5% другу на первый заказ, 100 ₽ вам после его покупки"
          icon={<ActionIcon name="users" size={20} />}
          active={isActive("/profile/referral")}
          variant="menu"
          size="md"
          fullWidth
        />
        <FilledButton
          label="Написать в поддержку"
          description="Помощь с заказом, оплатой или доставкой"
          icon={<ActionIcon name="message-circle" size={20} />}
          variant="menu"
          size="md"
          fullWidth
          onClick={openSupport}
        />
        <FilledButton
          href="/catalog"
          label="Каталог"
          description="Stars, Premium и подарки с мгновенной доставкой"
          icon={<ActionIcon name="star" size={20} />}
          active={isActive("/catalog")}
          variant="menu"
          size="md"
          fullWidth
        />
      </VStack>

      <VStack gap={2}>
        <Text type="label" weight="semibold" display="block">
          Правовая информация
        </Text>
        <FilledButton
          href="/legal/privacy"
          label="Политика конфиденциальности"
          description="Какие данные собираем и как защищаем"
          icon={<ActionIcon name="shield-user" size={20} />}
          active={isActive("/legal/privacy")}
          variant="menu"
          size="md"
          fullWidth
        />
        <FilledButton
          href="/legal/terms"
          label="Пользовательское соглашение"
          description="Условия покупки и реферальной программы"
          icon={<ActionIcon name="file-text" size={20} />}
          active={isActive("/legal/terms")}
          variant="menu"
          size="md"
          fullWidth
        />
        <FilledButton
          label="Безопасность аккаунта"
          description="Авторизация через подпись Telegram — подмена ID невозможна"
          icon={<AppIcon icon={HelpCircle} size={20} />}
          variant="menu"
          size="md"
          fullWidth
        />
      </VStack>
    </VStack>
  );
}
