"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";

export default function ProductsPage() {
  const apiFetch = useAuthStore((s) => s.apiFetch);
  const [products, setProducts] = useState<any[]>([]);

  useEffect(() => {
    apiFetch("/api/v1/admin/products?limit=50").then((r) => r.json()).then((d) => setProducts(d.items || []));
  }, [apiFetch]);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Products</h1>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-700 text-left text-slate-400">
            <th className="pb-2">Name</th>
            <th className="pb-2">Type</th>
            <th className="pb-2">Price</th>
            <th className="pb-2">Active</th>
          </tr>
        </thead>
        <tbody>
          {products.map((p) => (
            <tr key={p.id} className="border-b border-slate-800">
              <td className="py-3">{p.name}</td>
              <td>{p.product_type}</td>
              <td>{(p.price_kopecks / 100).toFixed(0)} ₽</td>
              <td>{p.is_active ? "✓" : "✗"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
