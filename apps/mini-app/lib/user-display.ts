import type { User } from "@/lib/api";

export function mergeUsers(apiUser: User, telegramUser: User | null): User {
  return {
    id: apiUser.id || telegramUser?.id || "",
    telegram_id: apiUser.telegram_id || telegramUser?.telegram_id || 0,
    username: apiUser.username || telegramUser?.username,
    first_name: apiUser.first_name || telegramUser?.first_name,
    last_name: apiUser.last_name || telegramUser?.last_name,
    photo_url: apiUser.photo_url || telegramUser?.photo_url,
  };
}

export function getDisplayName(user: User | null | undefined, isReady: boolean): string {
  if (!isReady) return "Загрузка…";
  if (!user?.telegram_id) return "Гость";

  const fullName = [user.first_name, user.last_name].filter(Boolean).join(" ");
  if (fullName) return fullName;
  if (user.username) return `@${user.username}`;
  return `Пользователь ${user.telegram_id}`;
}

export function getDisplayDescription(
  user: User | null | undefined,
  isReady: boolean,
  hasInitData: boolean
): string {
  if (!isReady) return "Подключаем профиль Telegram…";
  if (!user?.telegram_id) {
    return hasInitData
      ? "Не удалось загрузить профиль. Перезапустите Mini App из бота."
      : "Откройте магазин из Telegram, а не в обычном браузере.";
  }
  if (user.username) return `@${user.username} · ID ${user.telegram_id}`;
  return `ID ${user.telegram_id}`;
}
