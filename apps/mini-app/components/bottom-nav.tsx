"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Badge } from "@astryxdesign/core/Badge";
import { AppIcon, navIcons } from "@/components/icons";
import { cn } from "@/lib/utils";
import { useCartStore } from "@/stores/app";

const links = [
  { href: "/", label: "Главная", icon: navIcons.home },
  { href: "/catalog", label: "Каталог", icon: navIcons.catalog },
  { href: "/cart", label: "Корзина", icon: navIcons.cart },
  { href: "/profile", label: "Профиль", icon: navIcons.profile },
];

const idleStyles =
  "bg-[#e8e8e6] text-[#666666] hover:bg-[#d1d1cb] hover:text-[#1b1b1b] dark:bg-[#3f3f46] dark:text-[#a1a1aa] dark:hover:bg-[#52525b] dark:hover:text-[#fafafa]";

const activeStyles =
  "bg-[#007aff] text-white hover:bg-[#0066d6] hover:text-white dark:bg-[#0a84ff] dark:hover:bg-[#409cff] dark:hover:text-white";

export function BottomNav() {
  const pathname = usePathname();
  const count = useCartStore((s) => s.itemCount());

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 border-t border-[var(--color-border-default,rgba(0,0,0,.08))] bg-[var(--color-background-surface,#fff)] pb-[env(safe-area-inset-bottom)]">
      <div className="mx-auto grid max-w-lg grid-cols-4 gap-1 px-1 py-1.5">
        {links.map((link) => {
          const active = pathname === link.href || (link.href !== "/" && pathname.startsWith(link.href));
          return (
            <Link
              key={link.href}
              href={link.href}
              className={cn(
                "relative flex flex-col items-center gap-0.5 rounded-[var(--radius-container,10px)] px-1 py-1.5 text-xs font-medium no-underline transition-colors",
                active ? activeStyles : idleStyles
              )}
            >
              <AppIcon icon={link.icon} size={20} />
              <span className="leading-none">{link.label}</span>
              {link.href === "/cart" && count > 0 ? (
                <span className="absolute right-1 top-0.5">
                  <Badge variant="info" label={String(count)} />
                </span>
              ) : null}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
