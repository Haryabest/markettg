"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { api } from "@/lib/api";
import { ProductCard, Skeleton } from "@/components/ui";

export default function HomePage() {
  const { data: promos } = useQuery({
    queryKey: ["promotions"],
    queryFn: () => api.getPromotions(),
  });

  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => api.getCategories(),
  });

  const { data: popular } = useQuery({
    queryKey: ["products", "popular"],
    queryFn: () => api.getProducts({ sort: "popularity", limit: "6" }),
  });

  const { data: newest } = useQuery({
    queryKey: ["products", "newest"],
    queryFn: () => api.getProducts({ sort: "newest", limit: "6" }),
  });

  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-2xl font-bold">MarketTG</h1>
        <p className="text-sm text-zinc-500">Stars · Premium · Gifts</p>
      </header>

      {promos?.items && promos.items.length > 0 && (
        <section>
          <h2 className="mb-3 text-lg font-semibold">Акции</h2>
          <div className="space-y-2">
            {promos.items.map((p) => (
              <div key={p.id} className="rounded-2xl bg-gradient-to-r from-blue-500 to-purple-600 p-4 text-white">
                <h3 className="font-semibold">{p.title}</h3>
                {p.description && <p className="text-sm opacity-90">{p.description}</p>}
              </div>
            ))}
          </div>
        </section>
      )}

      <section>
        <h2 className="mb-3 text-lg font-semibold">Категории</h2>
        <div className="grid grid-cols-3 gap-2">
          {categories?.items?.map((c) => (
            <Link
              key={c.id}
              href={`/catalog?category=${c.slug}`}
              className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#f4f4f5)] p-4 text-center text-sm font-medium"
            >
              {c.name}
            </Link>
          )) || Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-16" />)}
        </div>
      </section>

      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Популярное</h2>
          <Link href="/catalog" className="text-sm text-[var(--tg-theme-link-color,#2481cc)]">Все</Link>
        </div>
        <div className="grid grid-cols-2 gap-3">
          {popular?.items?.map((p) => <ProductCard key={p.id} product={p} />) ||
            Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-48" />)}
        </div>
      </section>

      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Новинки</h2>
        </div>
        <div className="grid grid-cols-2 gap-3">
          {newest?.items?.map((p) => <ProductCard key={p.id} product={p} />)}
        </div>
      </section>
    </div>
  );
}
