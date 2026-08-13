"use client";

import Link from "next/link";
import { useAuthStore } from "@/stores/app";

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Профиль</h1>
      {user ? (
        <div className="rounded-2xl bg-zinc-50 p-4">
          <p className="font-semibold">{user.first_name} {user.last_name}</p>
          {user.username && <p className="text-sm text-zinc-500">@{user.username}</p>}
        </div>
      ) : (
        <p className="text-zinc-500">Откройте через Telegram Mini App</p>
      )}
      <nav className="space-y-2">
        <Link href="/orders" className="block rounded-2xl bg-zinc-50 p-4">Мои заказы</Link>
        <Link href="/favorites" className="block rounded-2xl bg-zinc-50 p-4">Избранное</Link>
      </nav>
    </div>
  );
}
