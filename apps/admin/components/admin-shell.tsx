"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth";

const NAV = [
  { href: "/", label: "Dashboard" },
  { href: "/products", label: "Products" },
  { href: "/categories", label: "Categories" },
  { href: "/orders", label: "Orders" },
  { href: "/users", label: "Users" },
  { href: "/promo-codes", label: "Promo Codes" },
  { href: "/promotions", label: "Promotions" },
  { href: "/payments", label: "Payments" },
  { href: "/deliveries", label: "Deliveries" },
  { href: "/audit-log", label: "Audit Log" },
];

export function AdminShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { token, logout } = useAuthStore();

  if (pathname === "/login") {
    return <>{children}</>;
  }

  if (!token) {
    if (typeof window !== "undefined") window.location.href = "/login";
    return null;
  }

  return (
    <div className="flex min-h-screen">
      <aside className="w-56 shrink-0 bg-[var(--sidebar)] p-4">
        <h1 className="mb-6 text-lg font-bold">MarketTG</h1>
        <nav className="space-y-1">
          {NAV.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`block rounded-lg px-3 py-2 text-sm ${
                pathname === item.href ? "bg-[var(--accent)] text-white" : "text-slate-300 hover:bg-slate-700"
              }`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <button onClick={logout} className="mt-8 text-sm text-slate-400 hover:text-white">
          Выйти
        </button>
      </aside>
      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}
