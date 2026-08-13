"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ProductCard } from "@/components/ui";

export default function FavoritesPage() {
  const { data: favs } = useQuery({ queryKey: ["favorites"], queryFn: () => api.getFavorites() });
  const ids = favs?.product_ids?.join(",") || "";

  const { data: products } = useQuery({
    queryKey: ["fav-products", ids],
    queryFn: async () => {
      if (!favs?.product_ids?.length) return { items: [] };
      const results = await Promise.all(favs.product_ids.map((id) => api.getProduct(id)));
      return { items: results };
    },
    enabled: !!favs?.product_ids?.length,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Избранное</h1>
      <div className="grid grid-cols-2 gap-3">
        {products?.items?.map((p) => <ProductCard key={p.id} product={p} />)}
      </div>
      {!products?.items?.length && <p className="text-center text-zinc-500">Пока пусто</p>}
    </div>
  );
}
