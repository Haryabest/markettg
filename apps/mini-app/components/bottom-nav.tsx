"use client";

import Link from "next/link";
import { useCallback, useRef } from "react";
import { usePathname } from "next/navigation";
import { ActionIcon, playActionIcon, type ActionIconHandle, type ActionIconName } from "@/components/action-icon";
import { navActionIcons } from "@/components/icons";
import { cn } from "@/lib/utils";
import { useCartStore } from "@/stores/app";
import { Badge } from "@/components/ui";

const links = [
  { href: "/", label: "Главная", icon: navActionIcons.home },
  { href: "/catalog", label: "Каталог", icon: navActionIcons.catalog },
  { href: "/cart", label: "Корзина", icon: navActionIcons.cart },
  { href: "/profile", label: "Профиль", icon: navActionIcons.profile },
] as const;

function NavItem({
  href,
  label,
  icon,
  active,
  cartCount,
}: {
  href: string;
  label: string;
  icon: ActionIconName;
  active: boolean;
  cartCount: number;
}) {
  const iconRef = useRef<ActionIconHandle | null>(null);

  const onPress = useCallback(() => {
    playActionIcon(iconRef.current);
  }, []);

  return (
    <Link
      href={href}
      onPointerDownCapture={onPress}
      className={cn(
        "relative flex flex-col items-center gap-0.5 rounded-[var(--radius-container,10px)] px-1 py-1.5 text-xs font-medium no-underline transition-colors",
        active
          ? "bg-[var(--tg-button-color,#007aff)] text-[var(--tg-button-text-color,#fff)]"
          : "bg-[var(--color-background-surface,var(--tg-secondary-bg-color,#e8e8e6))] text-[var(--tg-hint-color,var(--color-text-secondary,#666))]"
      )}
    >
      <ActionIcon ref={iconRef} name={icon} size={20} />
      <span className="pointer-events-none leading-none">{label}</span>
      {href === "/cart" && cartCount > 0 ? (
        <span className="pointer-events-none absolute right-1 top-0.5">
          <Badge variant="info" label={String(cartCount)} />
        </span>
      ) : null}
    </Link>
  );
}

export function BottomNav() {
  const pathname = usePathname();
  const count = useCartStore((s) => s.itemCount());

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 border-t border-[var(--color-border-default,rgba(0,0,0,.08))] bg-[var(--tg-secondary-bg-color,var(--color-background-surface,var(--tg-bg-color,#fff)))] pb-[env(safe-area-inset-bottom)]">
      <div className="mx-auto grid max-w-lg grid-cols-4 gap-1 px-1 py-1.5">
        {links.map((link) => {
          const active = pathname === link.href || (link.href !== "/" && pathname.startsWith(link.href));
          return (
            <NavItem
              key={link.href}
              href={link.href}
              label={link.label}
              icon={link.icon}
              active={active}
              cartCount={count}
            />
          );
        })}
      </div>
    </nav>
  );
}
