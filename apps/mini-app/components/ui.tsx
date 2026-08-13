import Link from "next/link";
import { cn, formatPrice } from "@/lib/utils";
import type { Product } from "@/lib/api";
import Image from "next/image";

export function ProductCard({ product, className }: { product: Product; className?: string }) {
  return (
    <Link
      href={`/product/${product.id}`}
      className={cn(
        "flex flex-col rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#f4f4f5)] p-3 transition active:scale-[0.98]",
        className
      )}
    >
      <div className="relative mb-2 aspect-square overflow-hidden rounded-xl bg-zinc-200">
        {product.image_url ? (
          <Image src={product.image_url} alt={product.name} fill className="object-cover" unoptimized />
        ) : (
          <div className="flex h-full items-center justify-center text-3xl">
            {product.product_type === "STARS" ? "⭐" : product.product_type === "PREMIUM" ? "👑" : "🎁"}
          </div>
        )}
      </div>
      <h3 className="line-clamp-2 text-sm font-medium">{product.name}</h3>
      <p className="mt-1 text-base font-semibold text-[var(--tg-theme-link-color,#2481cc)]">
        {formatPrice(product.price_kopecks)}
      </p>
    </Link>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn("animate-pulse rounded-xl bg-zinc-200", className)} />;
}

export function BottomNav() {
  const links = [
    { href: "/", label: "Главная", icon: "🏠" },
    { href: "/catalog", label: "Каталог", icon: "📦" },
    { href: "/search", label: "Поиск", icon: "🔍" },
    { href: "/cart", label: "Корзина", icon: "🛒" },
    { href: "/profile", label: "Профиль", icon: "👤" },
  ];
  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-zinc-200 bg-[var(--tg-theme-bg-color,#fff)] pb-safe">
      <div className="mx-auto flex max-w-lg justify-around px-2 py-2">
        {links.map((l) => (
          <Link key={l.href} href={l.href} className="flex flex-col items-center gap-0.5 px-2 py-1 text-xs text-zinc-600">
            <span className="text-lg">{l.icon}</span>
            {l.label}
          </Link>
        ))}
      </div>
    </nav>
  );
}
