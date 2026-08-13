"use client";

import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import { ProductCard } from "@/components/ui";

export default function SearchPage() {
  const [q, setQ] = useState("");
  const [debounced, setDebounced] = useState("");

  useEffect(() => {
    const t = setTimeout(() => setDebounced(q), 300);
    return () => clearTimeout(t);
  }, [q]);

  const { data } = useQuery({
    queryKey: ["search", debounced],
    queryFn: () => api.getProducts({ q: debounced, limit: "20" }),
    enabled: debounced.length >= 2,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Поиск</h1>
      <input
        type="search"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Premium, Stars, подарок..."
        className="w-full rounded-2xl border border-zinc-200 bg-[var(--tg-theme-secondary-bg-color,#f4f4f5)] px-4 py-3 outline-none"
      />
      {debounced.length >= 2 && (
        <div className="grid grid-cols-2 gap-3">
          {data?.items?.map((p) => <ProductCard key={p.id} product={p} />)}
          {data?.items?.length === 0 && <p className="col-span-2 text-center text-zinc-500">Ничего не найдено</p>}
        </div>
      )}
    </div>
  );
}
