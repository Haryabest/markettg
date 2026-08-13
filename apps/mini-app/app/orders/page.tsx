"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";

const STATUS_LABELS: Record<string, string> = {
  CREATED: "Создан",
  PAYMENT_PENDING: "Ожидает оплаты",
  PAID: "Оплачен",
  DELIVERING: "Доставляется",
  COMPLETED: "Выполнен",
  CANCELLED: "Отменён",
};

export default function OrdersPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["orders"],
    queryFn: () => api.getOrders(),
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Заказы</h1>
      {isLoading && <p>Загрузка...</p>}
      <ul className="space-y-3">
        {data?.orders?.map((o) => (
          <li key={o.id}>
            <Link href={`/orders/${o.id}`} className="block rounded-2xl bg-zinc-50 p-4">
              <div className="flex justify-between">
                <span className="font-medium">#{o.id.slice(0, 8)}</span>
                <span className="text-sm text-zinc-500">{STATUS_LABELS[o.status] || o.status}</span>
              </div>
              <p className="mt-1 font-semibold">{formatPrice(o.total_kopecks)}</p>
            </Link>
          </li>
        ))}
      </ul>
      {!isLoading && !data?.orders?.length && <p className="text-center text-zinc-500">Заказов пока нет</p>}
    </div>
  );
}
