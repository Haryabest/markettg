"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";
import { useCartStore } from "@/stores/app";
import { Skeleton } from "@/components/ui";
import Image from "next/image";

export default function ProductPage() {
  const { id } = useParams<{ id: string }>();
  const addItem = useCartStore((s) => s.addItem);

  const { data: product, isLoading } = useQuery({
    queryKey: ["product", id],
    queryFn: () => api.getProduct(id),
    enabled: !!id,
  });

  if (isLoading) return <Skeleton className="h-96" />;
  if (!product) return <p>Товар не найден</p>;

  return (
    <div className="space-y-6">
      <div className="relative aspect-square overflow-hidden rounded-2xl bg-zinc-100">
        {product.image_url ? (
          <Image src={product.image_url} alt={product.name} fill className="object-cover" unoptimized />
        ) : (
          <div className="flex h-full items-center justify-center text-6xl">
            {product.product_type === "STARS" ? "⭐" : product.product_type === "PREMIUM" ? "👑" : "🎁"}
          </div>
        )}
      </div>
      <div>
        <h1 className="text-2xl font-bold">{product.name}</h1>
        <p className="mt-2 text-2xl font-semibold text-[var(--tg-theme-link-color,#2481cc)]">
          {formatPrice(product.price_kopecks)}
        </p>
        {product.description && <p className="mt-4 text-zinc-600">{product.description}</p>}
      </div>
      <button
        onClick={() => addItem(product.id, 1)}
        className="w-full rounded-2xl bg-[var(--tg-theme-button-color,#2481cc)] py-4 text-lg font-semibold text-[var(--tg-theme-button-text-color,#fff)]"
      >
        В корзину
      </button>
    </div>
  );
}
