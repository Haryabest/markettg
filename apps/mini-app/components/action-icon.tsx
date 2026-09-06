"use client";

import {
  forwardRef,
  useCallback,
  useImperativeHandle,
  useRef,
  type ComponentType,
  type Ref,
} from "react";
import { ArrowLeftIcon } from "@animateicons/react/lucide/arrow-left-icon";
import { ArrowRightIcon } from "@animateicons/react/lucide/arrow-right-icon";
import { ArrowUpDownIcon } from "@animateicons/react/lucide/arrow-up-down-icon";
import { BanknoteIcon } from "@animateicons/react/lucide/banknote-icon";
import { CopyIcon } from "@animateicons/react/lucide/copy-icon";
import { CreditCardIcon } from "@animateicons/react/lucide/credit-card-icon";
import { FileTextIcon } from "@animateicons/react/lucide/file-text-icon";
import { GiftIcon } from "@animateicons/react/lucide/gift-icon";
import { HeartIcon } from "@animateicons/react/lucide/heart-icon";
import { HouseIcon } from "@animateicons/react/lucide/house-icon";
import { LayoutGridIcon } from "@animateicons/react/lucide/layout-grid-icon";
import { MessageCircleIcon } from "@animateicons/react/lucide/message-circle-icon";
import { MessageCircleMoreIcon } from "@animateicons/react/lucide/message-circle-more-icon";
import { PackageIcon } from "@animateicons/react/lucide/package-icon";
import { RefreshCwIcon } from "@animateicons/react/lucide/refresh-cw-icon";
import { SearchIcon } from "@animateicons/react/lucide/search-icon";
import { SendIcon } from "@animateicons/react/lucide/send-icon";
import { ShieldUserIcon } from "@animateicons/react/lucide/shield-user-icon";
import { ShoppingBagIcon } from "@animateicons/react/lucide/shopping-bag-icon";
import { SparklesIcon } from "@animateicons/react/lucide/sparkles-icon";
import { StarIcon } from "@animateicons/react/lucide/star-icon";
import { UserStarIcon } from "@animateicons/react/lucide/user-star-icon";
import { Trash2Icon } from "@animateicons/react/lucide/trash-2-icon";
import { UserRoundIcon } from "@animateicons/react/lucide/user-round-icon";
import { UsersIcon } from "@animateicons/react/lucide/users-icon";
import { WalletIcon } from "@animateicons/react/lucide/wallet-icon";
import { cn } from "@/lib/utils";

type IconHandle = { startAnimation: () => void; stopAnimation: () => void };
type AnimatedIcon = ComponentType<{
  size?: number;
  color?: string;
  duration?: number;
  isAnimated?: boolean;
  className?: string;
  ref?: Ref<IconHandle>;
}>;

/** Animated action icons from Animate UI / @animateicons/react (Lucide + Motion). */
export type ActionIconName =
  | "arrow-left"
  | "arrow-right"
  | "arrow-up-down"
  | "banknote"
  | "copy"
  | "credit-card"
  | "file-text"
  | "gift"
  | "heart"
  | "home"
  | "layout-grid"
  | "message-circle"
  | "message-circle-more"
  | "package"
  | "refresh-cw"
  | "search"
  | "send"
  | "shield-user"
  | "shopping-bag"
  | "sparkles"
  | "star"
  | "trash-2"
  | "user-round"
  | "user-star"
  | "users"
  | "wallet";

const ICONS: Record<ActionIconName, AnimatedIcon> = {
  "arrow-left": ArrowLeftIcon,
  "arrow-right": ArrowRightIcon,
  "arrow-up-down": ArrowUpDownIcon,
  banknote: BanknoteIcon,
  copy: CopyIcon,
  "credit-card": CreditCardIcon,
  "file-text": FileTextIcon,
  gift: GiftIcon,
  heart: HeartIcon,
  home: HouseIcon,
  "layout-grid": LayoutGridIcon,
  "message-circle": MessageCircleIcon,
  "message-circle-more": MessageCircleMoreIcon,
  package: PackageIcon,
  "refresh-cw": RefreshCwIcon,
  search: SearchIcon,
  send: SendIcon,
  "shield-user": ShieldUserIcon,
  "shopping-bag": ShoppingBagIcon,
  sparkles: SparklesIcon,
  star: StarIcon,
  "trash-2": Trash2Icon,
  "user-round": UserRoundIcon,
  "user-star": UserStarIcon,
  users: UsersIcon,
  wallet: WalletIcon,
};

export type ActionIconHandle = IconHandle;

export const ActionIcon = forwardRef<
  ActionIconHandle,
  {
    name: ActionIconName;
    size?: number;
    className?: string;
  }
>(function ActionIcon({ name, size = 18, className }, forwardedRef) {
  const iconRef = useRef<IconHandle | null>(null);

  useImperativeHandle(
    forwardedRef,
    () => ({
      startAnimation: () => iconRef.current?.startAnimation(),
      stopAnimation: () => iconRef.current?.stopAnimation(),
    }),
    []
  );

  const attachIconRef = useCallback((instance: IconHandle | null) => {
    iconRef.current = instance;
  }, []);

  const Icon = ICONS[name];

  return (
    <span
      className={cn("inline-flex shrink-0 select-none", className)}
      data-action-icon=""
      aria-hidden
      style={{ pointerEvents: "none" }}
    >
      <Icon
        ref={attachIconRef}
        size={size}
        color="currentColor"
        isAnimated={false}
        className="pointer-events-none"
      />
    </span>
  );
});

ActionIcon.displayName = "ActionIcon";

export function playActionIcon(ref: ActionIconHandle | null | undefined) {
  if (!ref) return;
  ref.stopAnimation();
  ref.startAnimation();
}
