"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [liveStatus, setLiveStatus] = useState<string | null>(null);

  const { data: order, refetch } = useQuery({
    queryKey: ["order", id],
    queryFn: () => api.getOrder(id),
    enabled: !!id,
    refetchInterval: 10000,
  });

  useEffect(() => {
    if (!id) return;
    const wsUrl = API_URL.replace("http", "ws") + `/api/v1/ws/orders/${id}`;
    let ws: WebSocket;
    try {
      ws = new WebSocket(wsUrl);
      ws.onmessage = (e) => {
        const data = JSON.parse(e.data);
        if (data.status) setLiveStatus(data.status);
      };
      ws.onclose = () => setTimeout(() => refetch(), 3000);
    } catch {
      // fallback polling via refetchInterval
    }
    return () => ws?.close();
  }, [id, refetch]);

  if (!order) return <p>Загрузка...</p>;

  const status = liveStatus || order.status;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Заказ #{order.id.slice(0, 8)}</h1>
      <div className="rounded-2xl bg-zinc-50 p-4">
        <p className="text-sm text-zinc-500">Статус</p>
        <p className="text-lg font-semibold">{status}</p>
      </div>
      <p className="text-xl font-bold">{formatPrice(order.total_kopecks)}</p>
      {order.items && (
        <ul className="space-y-2">
          {order.items.map((item) => (
            <li key={item.id} className="flex justify-between rounded-xl bg-zinc-50 p-3 text-sm">
              <span>{item.name} × {item.quantity}</span>
              <span>{formatPrice(item.price_kopecks * item.quantity)}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
