"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";

export default function DashboardPage() {
  const apiFetch = useAuthStore((s) => s.apiFetch);
  const [stats, setStats] = useState({ orders: 0, pending: 0, deliveries: 0 });

  useEffect(() => {
    Promise.all([
      apiFetch("/api/v1/admin/orders?limit=100").then((r) => r.json()),
      apiFetch("/api/v1/admin/deliveries?limit=100").then((r) => r.json()),
    ]).then(([ordersData, deliveriesData]) => {
      const orders = ordersData.orders || [];
      const pending = orders.filter((o: { status: string }) =>
        ["PAYMENT_PENDING", "PAID", "DELIVERY_PENDING", "DELIVERING"].includes(o.status)
      ).length;
      setStats({
        orders: orders.length,
        pending,
        deliveries: (deliveriesData.deliveries || []).length,
      });
    });
  }, [apiFetch]);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div className="rounded-xl bg-[var(--sidebar)] p-6">
          <p className="text-sm text-slate-400">Заказы</p>
          <p className="mt-2 text-3xl font-bold">{stats.orders || "—"}</p>
        </div>
        <div className="rounded-xl bg-[var(--sidebar)] p-6">
          <p className="text-sm text-slate-400">Ожидают обработки</p>
          <p className="mt-2 text-3xl font-bold">{stats.pending || "—"}</p>
        </div>
        <div className="rounded-xl bg-[var(--sidebar)] p-6">
          <p className="text-sm text-slate-400">Deliveries</p>
          <p className="mt-2 text-3xl font-bold">{stats.deliveries || "—"}</p>
        </div>
        <div className="rounded-xl bg-[var(--sidebar)] p-6">
          <p className="text-sm text-slate-400">Revenue</p>
          <p className="mt-2 text-3xl font-bold">—</p>
        </div>
      </div>
    </div>
  );
}
