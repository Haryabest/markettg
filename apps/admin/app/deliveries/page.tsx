"use client";
import { useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";
export default function Page() {
  const apiFetch = useAuthStore((s) => s.apiFetch);
  const [data, setData] = useState<any>(null);
  useEffect(() => { apiFetch("/api/v1/admin/deliveries").then(r => r.json()).then(setData); }, [apiFetch]);
  return <div><h1 className="text-2xl font-bold mb-6">Deliveries</h1><pre className="text-xs bg-[var(--sidebar)] p-4 rounded-xl overflow-auto">{JSON.stringify(data, null, 2)}</pre></div>;
}
