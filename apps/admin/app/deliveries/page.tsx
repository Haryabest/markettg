"use client";

import { useCallback, useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";

type Delivery = {
  id: string;
  order_id: string;
  telegram_id: number;
  handler_type: string;
  status: string;
  attempts: number;
  max_attempts: number;
  last_error?: string | null;
  completed_at?: string | null;
  created_at: string;
};

const STATUS_LABELS: Record<string, string> = {
  PENDING: "Ожидает",
  PROCESSING: "В процессе",
  COMPLETED: "Выполнено",
  FAILED: "Ошибка",
};

function statusClass(status: string) {
  switch (status) {
    case "COMPLETED":
      return "text-emerald-400";
    case "FAILED":
      return "text-red-400";
    case "PROCESSING":
      return "text-amber-400";
    default:
      return "text-slate-300";
  }
}

export default function DeliveriesPage() {
  const apiFetch = useAuthStore((s) => s.apiFetch);
  const [deliveries, setDeliveries] = useState<Delivery[]>([]);
  const [loading, setLoading] = useState(true);
  const [actingId, setActingId] = useState<string | null>(null);
  const [error, setError] = useState("");

  const loadDeliveries = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await apiFetch("/api/v1/admin/deliveries?limit=100");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setDeliveries(data.deliveries || []);
    } catch {
      setError("Не удалось загрузить доставки");
    } finally {
      setLoading(false);
    }
  }, [apiFetch]);

  useEffect(() => {
    loadDeliveries();
  }, [loadDeliveries]);

  const confirmDelivery = async (id: string) => {
    setActingId(id);
    setError("");
    try {
      const res = await apiFetch(`/api/v1/admin/deliveries/${id}/confirm`, { method: "POST" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      await loadDeliveries();
    } catch {
      setError("Не удалось подтвердить доставку");
    } finally {
      setActingId(null);
    }
  };

  const retryDelivery = async (id: string) => {
    setActingId(id);
    setError("");
    try {
      const res = await apiFetch(`/api/v1/admin/deliveries/${id}/retry`, { method: "POST" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      await loadDeliveries();
    } catch {
      setError("Не удалось повторить доставку");
    } finally {
      setActingId(null);
    }
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Доставки</h1>
          <p className="mt-1 text-sm text-slate-400">Задачи на отправку Stars / Premium / Gift</p>
        </div>
        <button
          onClick={loadDeliveries}
          className="rounded-lg bg-slate-700 px-4 py-2 text-sm hover:bg-slate-600"
        >
          Обновить
        </button>
      </div>

      {error ? <p className="mb-4 text-sm text-red-400">{error}</p> : null}

      {loading ? (
        <p className="text-slate-400">Загрузка…</p>
      ) : deliveries.length === 0 ? (
        <p className="rounded-xl bg-[var(--sidebar)] p-6 text-slate-400">
          Задач доставки пока нет — они появятся после оплаты заказа
        </p>
      ) : (
        <div className="overflow-x-auto rounded-xl bg-[var(--sidebar)]">
          <table className="w-full min-w-[900px] text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-left text-slate-400">
                <th className="px-4 py-3">Дата</th>
                <th className="px-4 py-3">Заказ</th>
                <th className="px-4 py-3">TG ID</th>
                <th className="px-4 py-3">Тип</th>
                <th className="px-4 py-3">Статус</th>
                <th className="px-4 py-3">Попытки</th>
                <th className="px-4 py-3">Действие</th>
              </tr>
            </thead>
            <tbody>
              {deliveries.map((delivery) => (
                <tr key={delivery.id} className="border-b border-slate-800">
                  <td className="px-4 py-3 whitespace-nowrap">
                    {new Date(delivery.created_at).toLocaleString("ru-RU")}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs">{delivery.order_id.slice(0, 8)}…</td>
                  <td className="px-4 py-3">{delivery.telegram_id || "—"}</td>
                  <td className="px-4 py-3">{delivery.handler_type}</td>
                  <td className={`px-4 py-3 ${statusClass(delivery.status)}`}>
                    {STATUS_LABELS[delivery.status] || delivery.status}
                  </td>
                  <td className="px-4 py-3">
                    {delivery.attempts}/{delivery.max_attempts}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      {delivery.status !== "COMPLETED" ? (
                        <button
                          onClick={() => confirmDelivery(delivery.id)}
                          disabled={actingId === delivery.id}
                          className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs text-white hover:bg-emerald-500 disabled:opacity-50"
                        >
                          Подтвердить
                        </button>
                      ) : null}
                      {delivery.status === "FAILED" || delivery.status === "PENDING" ? (
                        <button
                          onClick={() => retryDelivery(delivery.id)}
                          disabled={actingId === delivery.id}
                          className="rounded-lg bg-slate-600 px-3 py-1.5 text-xs hover:bg-slate-500 disabled:opacity-50"
                        >
                          Повторить
                        </button>
                      ) : null}
                    </div>
                    {delivery.last_error ? (
                      <p className="mt-1 max-w-xs text-xs text-red-400">{delivery.last_error}</p>
                    ) : null}
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
