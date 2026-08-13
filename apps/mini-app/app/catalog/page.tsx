"use client";

import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { api } from "@/lib/api";
import { ProductCard, Skeleton } from "@/components/ui";

function CatalogContent() {
  const params = useSearchParams();
  const [sort, setSort] = useState("popularity");
  const category = params.get("category") || "";
  const type = params.get("type") || "";

  const { data, isLoading } = useQuery({
    queryKey: ["products", category, type, sort],
    queryFn: () => api.getProducts({
      ...(category && { category }),
      ...(type && { type }),
      sort,
      limit: "20",
    }),
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Каталог</h1>
      <div className="flex gap-2 overflow-x-auto pb-2">
        {[
          { v: "popularity", l: "Популярные" },
          { v: "newest", l: "Новые" },
          { v: "price_asc", l: "Дешевле" },
          { v: "price_desc", l: "Дороже" },
        ].map((s) => (
          <button
            key={s.v}
            onClick={() => setSort(s.v)}
            className={`shrink-0 rounded-full px-4 py-1.5 text-sm ${
              sort === s.v ? "bg-[var(--tg-theme-button-color,#2481cc)] text-white" : "bg-zinc-100"
            }`}
          >
            {s.l}
          </button>
        ))}
      </div>
      <div className="grid grid-cols-2 gap-3">
        {isLoading
          ? Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-48" />)
          : data?.items?.map((p) => <ProductCard key={p.id} product={p} />)}
      </div>
    </div>
  );
}

export default function CatalogPage() {
  return (
    <Suspense>
      <CatalogContent />
    </Suspense>
  );
}
