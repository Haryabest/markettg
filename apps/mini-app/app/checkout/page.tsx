"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { formatPrice } from "@/lib/utils";

export default function CheckoutPage() {
  const router = useRouter();
  const [promo, setPromo] = useState("");
  const [method, setMethod] = useState<"STARS" | "SBP">("STARS");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleCheckout = async () => {
    setLoading(true);
    setError("");
    try {
      const order = await api.createOrder(promo || undefined);
      const payment = await api.createPayment(order.id, method);
      if (payment.payment_url) {
        window.open(payment.payment_url, "_blank");
      }
      router.push(`/orders/${order.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка оформления");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Оформление</h1>
      <input
        value={promo}
        onChange={(e) => setPromo(e.target.value)}
        placeholder="Промокод"
        className="w-full rounded-2xl border border-zinc-200 px-4 py-3"
      />
      <div className="space-y-2">
        <p className="font-medium">Способ оплаты</p>
        {(["STARS", "SBP"] as const).map((m) => (
          <button
            key={m}
            onClick={() => setMethod(m)}
            className={`w-full rounded-2xl border p-4 text-left ${
              method === m ? "border-[var(--tg-theme-button-color,#2481cc)] bg-blue-50" : "border-zinc-200"
            }`}
          >
            {m === "STARS" ? "⭐ Telegram Stars" : "💳 СБП"}
          </button>
        ))}
      </div>
      {error && <p className="text-red-500">{error}</p>}
      <button
        onClick={handleCheckout}
        disabled={loading}
        className="w-full rounded-2xl bg-[var(--tg-theme-button-color,#2481cc)] py-4 text-lg font-semibold text-white disabled:opacity-50"
      >
        {loading ? "Обработка..." : "Оплатить"}
      </button>
    </div>
  );
}
