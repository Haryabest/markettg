"use client";
import { Heading, Text, VStack } from "@/components/ui";

import { useQuery } from "@tanstack/react-query";
import { BackHeader } from "@/components/back-header";
import { FilledButton } from "@/components/filled-button";
import { ActionIcon } from "@/components/action-icon";
import { AppIcon } from "@/components/icons";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";
import { getDisplayName } from "@/lib/user-display";
import { notifyError, notifySuccess } from "@/stores/banners";
import { useAuthStore } from "@/stores/app";
import { Gift } from "lucide-react";

export default function ReferralPage() {
  const initData = useAuthStore((s) => s.initData);
  const isReady = useAuthStore((s) => s.isReady);
  const user = useAuthStore((s) => s.user ?? s.telegramUser);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["referrals"],
    queryFn: () => api.getReferrals(),
    enabled: isReady && Boolean(initData),
    retry: 1,
  });

  const copyText = async (text: string, successTitle: string) => {
    try {
      await navigator.clipboard.writeText(text);
      notifySuccess(successTitle, text);
    } catch {
      notifyError("Не удалось скопировать", text);
    }
  };

  const copyLink = async () => {
    if (!data?.link) {
      notifyError("Нужен вход через Telegram", "Откройте магазин из бота, чтобы получить реферальную ссылку.");
      return;
    }
    await copyText(data.link, "Ссылка скопирована");
  };

  const shareLink = async () => {
    if (!data?.link) {
      notifyError("Нужен вход через Telegram", "Откройте магазин из бота, чтобы поделиться ссылкой.");
      return;
    }
    const text = "Скидка 5% на первый заказ в MarketTG";
    try {
      const { default: WebApp } = await import("@twa-dev/sdk");
      if (WebApp.openTelegramLink) {
        WebApp.openTelegramLink(
          `https://t.me/share/url?url=${encodeURIComponent(data.link)}&text=${encodeURIComponent(text)}`
        );
        return;
      }
    } catch {
      // fallback to clipboard
    }
    await copyLink();
  };

  const needsTelegram = isReady && !initData;
  const waitingAuth = !isReady || (Boolean(initData) && isLoading);
  const welcomeBonus = data?.bonus_info?.referred_welcome ?? "5% на первый заказ друга";
  const referrerBonus = data?.bonus_info?.referrer_reward ?? "100 ₽ после первой покупки друга";

  return (
    <VStack gap={5}>
      <BackHeader fallbackHref="/profile" />
      <VStack gap={1}>
        <Heading level={1}>Пригласи друга</Heading>
        <Text type="supporting" color="secondary" display="block">
          Делитесь ссылкой — вы оба получите бонусы после первой покупки друга.
        </Text>
        {user?.telegram_id ? (
          <Text type="supporting" color="secondary" display="block">
            {getDisplayName(user, true)}
          </Text>
        ) : null}
      </VStack>

      <VStack gap={2}>
        <FilledButton
          label={`Друг получает ${welcomeBonus}`}
          description="Промокод FRIEND* появится у друга после регистрации по вашей ссылке"
          icon={<AppIcon icon={Gift} size={20} />}
          variant="menu"
          size="md"
          fullWidth
        />
        <FilledButton
          label={`Вы получаете ${referrerBonus}`}
          description="Персональный промокод на следующий заказ"
          icon={<ActionIcon name="users" size={20} />}
          variant="menu"
          size="md"
          fullWidth
        />
      </VStack>

      {waitingAuth ? (
        <Text type="supporting" color="secondary" display="block">
          Загрузка статистики…
        </Text>
      ) : needsTelegram ? (
        <Text type="supporting" color="secondary" display="block">
          Откройте магазин из Telegram (кнопка «Открыть магазин» в боте), а не в обычном браузере.
        </Text>
      ) : isError || !data ? (
        <VStack gap={2}>
          <Text type="supporting" color="secondary" display="block">
            Не удалось загрузить реферальную ссылку. Проверьте интернет и попробуйте снова.
          </Text>
          <FilledButton label="Повторить" size="md" fullWidth icon={<ActionIcon name="refresh-cw" size={18} />} onClick={() => refetch()} />
        </VStack>
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
            label="Поделиться в Telegram"
            icon={<ActionIcon name="send" size={18} />}
            size="md"
            fullWidth
            onClick={shareLink}
          />
          <FilledButton
            label="Скопировать ссылку"
            icon={<ActionIcon name="copy" size={18} />}
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
                        ? `Скидка ${reward.discount_value}% · нажмите, чтобы скопировать`
                        : `Скидка ${formatPrice(reward.discount_value)} · нажмите, чтобы скопировать`
                  }
                  active={!reward.is_used}
                  variant="menu"
                  size="md"
                  fullWidth
                  onClick={
                    reward.is_used ? undefined : () => copyText(reward.promo_code, "Промокод скопирован")
                  }
                  icon={reward.is_used ? undefined : <ActionIcon name="copy" size={18} />}
                />
              ))}
            </VStack>
          ) : null}
        </VStack>
      )}
    </VStack>
  );
}
