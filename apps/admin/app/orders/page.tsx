"use client";

import { useCallback, useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";

type OrderItem = {
  id: string;
  name: string;
  product_type: string;
  quantity: number;
  price_kopecks: number;
};

type Order = {
  id: string;
  user_id: string;
  status: string;
  total_kopecks: number;
  discount_kopecks?: number;
  telegram_id?: number;
  customer_name?: string;
  items?: OrderItem[];
  created_at: string;
};

const STATUS_LABELS: Record<string, string> = {
  PAYMENT_PENDING: "Ожидает оплаты",
  PAID: "Оплачен",
  DELIVERY_PENDING: "Ожидает отправки",
  DELIVERING: "Отправляется",
  COMPLETED: "Выполнен",
  CANCELLED: "Отменён",
  DELIVERY_FAILED: "Ошибка доставки",
};

function statusLabel(status: string) {
  return STATUS_LABELS[status] || status;
}

function statusClass(status: string) {
  switch (status) {
    case "COMPLETED":
      return "text-emerald-400";
    case "PAID":
    case "DELIVERY_PENDING":
    case "DELIVERING":
      return "text-amber-400";
    case "CANCELLED":
    case "DELIVERY_FAILED":
      return "text-red-400";
    default:
      return "text-slate-300";
  }
}

function canConfirm(status: string) {
  return status === "PAID" || status === "DELIVERY_PENDING" || status === "DELIVERING";
}

function formatPrice(kopecks: number) {
  return `${(kopecks / 100).toFixed(0)} ₽`;
}

function formatDate(value: string) {
  return new Date(value).toLocaleString("ru-RU");
}

function itemsSummary(items?: OrderItem[]) {
  if (!items?.length) return "—";
  return items.map((item) => `${item.name} ×${item.quantity}`).join(", ");
}

export default function OrdersPage() {
  const apiFetch = useAuthStore((s) => s.apiFetch);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [confirmingId, setConfirmingId] = useState<string | null>(null);
  const [error, setError] = useState("");

  const loadOrders = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await apiFetch("/api/v1/admin/orders?limit=100");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setOrders(data.orders || []);
    } catch {
      setError("Не удалось загрузить заказы");
    } finally {
      setLoading(false);
    }
  }, [apiFetch]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const confirmShipment = async (orderId: string) => {
    setConfirmingId(orderId);
    setError("");
    try {
      const res = await apiFetch(`/api/v1/admin/orders/${orderId}/confirm-shipment`, {
        method: "POST",
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body?.error?.message || `HTTP ${res.status}`);
      }
      await loadOrders();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось подтвердить отправку");
    } finally {
      setConfirmingId(null);
    }
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Заказы</h1>
          <p className="mt-1 text-sm text-slate-400">Заявки на покупку из Mini App</p>
        </div>
        <button
          onClick={loadOrders}
          className="rounded-lg bg-slate-700 px-4 py-2 text-sm hover:bg-slate-600"
        >
          Обновить
        </button>
      </div>

      {error ? <p className="mb-4 text-sm text-red-400">{error}</p> : null}

      {loading ? (
        <p className="text-slate-400">Загрузка…</p>
      ) : orders.length === 0 ? (
        <p className="rounded-xl bg-[var(--sidebar)] p-6 text-slate-400">Заказов пока нет</p>
      ) : (
        <div className="overflow-x-auto rounded-xl bg-[var(--sidebar)]">
          <table className="w-full min-w-[960px] text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-left text-slate-400">
                <th className="px-4 py-3">Дата</th>
                <th className="px-4 py-3">Покупатель</th>
                <th className="px-4 py-3">Товары</th>
                <th className="px-4 py-3">Сумма</th>
                <th className="px-4 py-3">Статус</th>
                <th className="px-4 py-3">Действие</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id} className="border-b border-slate-800 align-top">
                  <td className="px-4 py-3 whitespace-nowrap">{formatDate(order.created_at)}</td>
                  <td className="px-4 py-3">
                    <div>{order.customer_name || "—"}</div>
                    <div className="text-xs text-slate-500">
                      {order.telegram_id ? `TG ${order.telegram_id}` : order.user_id.slice(0, 8)}
                    </div>
                  </td>
                  <td className="px-4 py-3 max-w-xs">{itemsSummary(order.items)}</td>
                  <td className="px-4 py-3 whitespace-nowrap">{formatPrice(order.total_kopecks)}</td>
                  <td className={`px-4 py-3 whitespace-nowrap ${statusClass(order.status)}`}>
                    {statusLabel(order.status)}
                  </td>
                  <td className="px-4 py-3">
                    {canConfirm(order.status) ? (
                      <button
                        onClick={() => confirmShipment(order.id)}
                        disabled={confirmingId === order.id}
                        className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-emerald-500 disabled:opacity-50"
                      >
                        {confirmingId === order.id ? "…" : "Подтвердить отправку"}
                      </button>
                    ) : order.status === "COMPLETED" ? (
                      <span className="text-xs text-emerald-400">Отправлено</span>
                    ) : (
                      <span className="text-xs text-slate-500">—</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
