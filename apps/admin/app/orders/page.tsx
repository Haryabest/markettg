"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";

function DataPage({ title, path, columns }: { title: string; path: string; columns: string[] }) {
  const apiFetch = useAuthStore((s) => s.apiFetch);
  const [items, setItems] = useState<any[]>([]);

  useEffect(() => {
    apiFetch(path).then((r) => r.json()).then((d) => {
      const key = Object.keys(d).find((k) => Array.isArray(d[k]));
      setItems(key ? d[key] : []);
    });
  }, [apiFetch, path]);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">{title}</h1>
      <pre className="overflow-auto rounded-xl bg-[var(--sidebar)] p-4 text-xs">
        {JSON.stringify(items, null, 2)}
      </pre>
    </div>
  );
}

export default function OrdersPage() {
  return <DataPage title="Orders" path="/api/v1/admin/orders" columns={["id", "status"]} />;
}
