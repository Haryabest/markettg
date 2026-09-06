import { Heading, Text, VStack } from "@/components/ui";

export default function PrivacyPage() {
  return (
    <VStack gap={4}>
      <Heading level={1}>Политика конфиденциальности</Heading>
      <Text type="body" display="block">
        MarketTG обрабатывает данные вашего Telegram-аккаунта (ID, имя, username, фото) только для
        авторизации, оформления заказов и доставки цифровых товаров.
      </Text>
      <Text type="body" display="block">
        Корзина, избранное и история заказов привязаны к вашему Telegram ID и не доступны другим
        пользователям. Данные передаются по защищённому соединению; подмена аккаунта блокируется
        проверкой подписи Telegram initData на сервере.
      </Text>
      <Text type="body" display="block">
        Мы не продаём персональные данные третьим лицам. Для удаления данных обратитесь в поддержку
        через профиль.
      </Text>
      <Text type="supporting" color="secondary" display="block">
        Обновлено: 20.08.2026
      </Text>
    </VStack>
  );
}
