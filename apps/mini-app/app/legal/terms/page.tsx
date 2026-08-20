import { Heading } from "@astryxdesign/core/Heading";
import { Text } from "@astryxdesign/core/Text";
import { VStack } from "@astryxdesign/core/VStack";

export default function TermsPage() {
  return (
    <VStack gap={4}>
      <Heading level={1}>Пользовательское соглашение</Heading>
      <Text type="body" display="block">
        Используя MarketTG, вы соглашаетесь покупать цифровые товары (Telegram Stars, Premium,
        подарки) для своего аккаунта Telegram. Оплата производится через Telegram Stars или СБП.
      </Text>
      <Text type="body" display="block">
        Доставка выполняется автоматически после успешной оплаты. Возврат возможен только при
        технической ошибке со стороны сервиса — обратитесь в поддержку с номером заказа.
      </Text>
      <Text type="body" display="block">
        Реферальная программа: бонусы начисляются после первой оплаченной покупки приглашённого
        пользователя. Запрещены самоприглашения и накрутка. Администрация может отменить бонусы при
        нарушении правил.
      </Text>
      <Text type="supporting" color="secondary" display="block">
        Обновлено: 20.08.2026
      </Text>
    </VStack>
  );
}
