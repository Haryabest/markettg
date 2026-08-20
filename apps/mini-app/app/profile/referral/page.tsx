"use client";

import { useQuery } from "@tanstack/react-query";
import { Heading } from "@astryxdesign/core/Heading";
import { Text } from "@astryxdesign/core/Text";
import { VStack } from "@astryxdesign/core/VStack";
import { FilledButton } from "@/components/filled-button";
import { AppIcon } from "@/components/icons";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";
import { notifyError, notifySuccess } from "@/stores/banners";
import { Copy, Gift, Users } from "lucide-react";

export default function ReferralPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["referrals"],
    queryFn: () => api.getReferrals(),
    retry: false,
  });

  const copyLink = async () => {
    if (!data?.link) {
      notifyError("Нужен вход через Telegram", "Откройте магазин из бота, чтобы получить реферальную ссылку.");
      return;
    }
    try {
      await navigator.clipboard.writeText(data.link);
      notifySuccess("Ссылка скопирована", "Отправьте её друзьям в Telegram");
    } catch {
      notifyError("Не удалось скопировать", data.link);
    }
  };

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>Пригласи друга</Heading>
        <Text type="supporting" color="secondary" display="block">
          Делитесь ссылкой — вы оба получите бонусы после первой покупки друга.
        </Text>
      </VStack>

      <VStack gap={2}>
        <FilledButton
          label="Друг получает 5% на первый заказ"
          description="Промокод FRIEND* появится у друга после регистрации по вашей ссылке"
          icon={<AppIcon icon={Gift} size={20} />}
          variant="menu"
          size="md"
          fullWidth
        />
        <FilledButton
          label="Вы получаете 100 ₽"
          description="После первой оплаты друга — персональный промокод на следующий заказ"
          icon={<AppIcon icon={Users} size={20} />}
          variant="menu"
          size="md"
          fullWidth
        />
      </VStack>

      {isLoading ? (
        <Text type="supporting" color="secondary" display="block">
          Загрузка статистики…
        </Text>
      ) : isError || !data ? (
        <Text type="supporting" color="secondary" display="block">
          Откройте магазин из Telegram, чтобы активировать реферальную программу.
        </Text>
      ) : (
        <VStack gap={3}>
          <VStack gap={1}>
            <Text type="label" weight="semibold" display="block">
              Ваша ссылка
            </Text>
            <Text type="body" display="block">
              {data.link || `Код: ${data.code}`}
            </Text>
          </VStack>

          <div className="grid grid-cols-3 gap-2">
            <FilledButton label={`${data.invited_count}`} description="Приглашено" fullWidth />
            <FilledButton label={`${data.qualified_count}`} description="Оплатили" active fullWidth />
            <FilledButton
              label={formatPrice(data.total_bonus_kopecks)}
              description="Бонусов"
              fullWidth
            />
          </div>

          <FilledButton
            label="Скопировать ссылку"
            icon={<AppIcon icon={Copy} size={18} />}
            size="md"
            fullWidth
            onClick={copyLink}
          />

          {data.rewards?.length ? (
            <VStack gap={2}>
              <Text type="label" weight="semibold" display="block">
                Ваши промокоды
              </Text>
              {data.rewards.map((reward) => (
                <FilledButton
                  key={reward.id}
                  label={reward.promo_code}
                  description={
                    reward.is_used
                      ? "Использован"
                      : reward.discount_type === "PERCENT"
                        ? `Скидка ${reward.discount_value}%`
                        : `Скидка ${formatPrice(reward.discount_value)}`
                  }
                  active={!reward.is_used}
                  variant="menu"
                  size="md"
                  fullWidth
                />
              ))}
            </VStack>
          ) : null}
        </VStack>
      )}
    </VStack>
  );
}
