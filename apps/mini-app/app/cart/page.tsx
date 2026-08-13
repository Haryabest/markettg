"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useCartStore } from "@/stores/app";
import { formatPrice } from "@/lib/utils";

export default function CartPage() {
  const { cart, fetchCart, addItem, removeItem } = useCartStore();

  useEffect(() => {
    fetchCart();
  }, [fetchCart]);

  const items = cart?.items || [];

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Корзина</h1>
      {items.length === 0 ? (
        <div className="py-12 text-center text-zinc-500">
          <p>Корзина пуста</p>
          <Link href="/catalog" className="mt-4 inline-block text-[var(--tg-theme-link-color,#2481cc)]">
            Перейти в каталог
          </Link>
        </div>
      ) : (
        <>
          <ul className="space-y-3">
            {items.map((item) => (
              <li key={item.product_id} className="flex items-center justify-between rounded-2xl bg-zinc-50 p-4">
                <Link href={`/product/${item.product_id}`} className="text-sm font-medium">
                  Товар {item.product_id.slice(0, 8)}...
                </Link>
                <div className="flex items-center gap-2">
                  <button onClick={() => addItem(item.product_id, Math.max(0, item.quantity - 1))} className="h-8 w-8 rounded-full bg-zinc-200">−</button>
                  <span>{item.quantity}</span>
                  <button onClick={() => addItem(item.product_id, item.quantity + 1)} className="h-8 w-8 rounded-full bg-zinc-200">+</button>
                  <button onClick={() => removeItem(item.product_id)} className="ml-2 text-red-500">✕</button>
                </div>
              </li>
            ))}
          </ul>
          <Link
            href="/checkout"
            className="block w-full rounded-2xl bg-[var(--tg-theme-button-color,#2481cc)] py-4 text-center text-lg font-semibold text-white"
          >
            Оформить заказ
          </Link>
        </>
      )}
    </div>
  );
}
