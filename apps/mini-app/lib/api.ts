import { dismissBanner, notifyError, notifySlow } from "@/stores/banners";

function apiBase(): string {
  const raw = (process.env.NEXT_PUBLIC_API_URL ?? "").replace(/\/$/, "");
  if (raw) {
    if (typeof window !== "undefined" && (raw.includes("localhost") || raw.includes("127.0.0.1"))) {
      return "";
    }
    return raw;
  }
  return typeof window === "undefined" ? "http://localhost:8080" : "";
}

const API_URL = apiBase();

const SLOW_REQUEST_MS = 2500;

export type ApiError = {
  error: { code: string; message: string; request_id?: string };
};

export class ApiClient {
  private initData: string = "";

  setInitData(initData: string) {
    this.initData = initData;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };
    if (this.initData) {
      headers["Authorization"] = `tma ${this.initData}`;
    }

    let slowId: string | null = null;
    let slowTimer: number | undefined;
    if (typeof window !== "undefined") {
      slowTimer = window.setTimeout(() => {
        slowId = notifySlow("Подключаемся…", "Сервер отвечает дольше обычного, подождите немного.");
      }, SLOW_REQUEST_MS);
    }

    try {
      const res = await fetch(`${API_URL}${path}`, {
        ...options,
        headers,
        signal: options.signal ?? AbortSignal.timeout(12_000),
      });
      if (!res.ok) {
        const err = (await res.json().catch(() => ({}))) as ApiError;
        throw new Error(err.error?.message || `HTTP ${res.status}`);
      }
      if (res.status === 204) return {} as T;
      return res.json();
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") {
        notifyError("Сервер не отвечает", "Проверьте сеть или попробуйте позже.");
        throw new Error("Сервер не отвечает. Откройте магазин заново или проверьте сеть.");
      }
      throw error;
    } finally {
      if (slowTimer !== undefined) window.clearTimeout(slowTimer);
      if (slowId) dismissBanner(slowId);
    }
  }

  authTelegram(initData: string) {
    return this.request<{ user: User }>("/api/v1/auth/telegram", {
      method: "POST",
      body: JSON.stringify({ init_data: initData }),
    });
  }

  getMe() {
    return this.request<User>("/api/v1/me");
  }

  getCategories() {
    return this.request<{ items: Category[] }>("/api/v1/catalog/categories");
  }

  getProducts(params?: Record<string, string>) {
    const qs = params ? "?" + new URLSearchParams(params).toString() : "";
    return this.request<ProductListResult>(`/api/v1/catalog/products${qs}`);
  }

  getProduct(id: string) {
    return this.request<Product>(`/api/v1/catalog/products/${id}`);
  }

  getPromotions() {
    return this.request<{ items: Promotion[] }>("/api/v1/catalog/promotions");
  }

  getCart() {
    return this.request<Cart>("/api/v1/cart");
  }

  updateCartItem(productId: string, quantity: number) {
    return this.request<Cart>("/api/v1/cart/items", {
      method: "PUT",
      body: JSON.stringify({ product_id: productId, quantity }),
    });
  }

  removeCartItem(productId: string) {
    return this.request<Cart>(`/api/v1/cart/items/${productId}`, { method: "DELETE" });
  }

  createOrder(promoCode?: string) {
    return this.request<Order>("/api/v1/orders", {
      method: "POST",
      body: JSON.stringify({ promo_code: promoCode || "" }),
      headers: { "Idempotency-Key": crypto.randomUUID() },
    });
  }

  getOrders() {
    return this.request<{ orders: Order[] }>("/api/v1/orders");
  }

  getOrder(id: string) {
    return this.request<Order>(`/api/v1/orders/${id}`);
  }

  createPayment(orderId: string, method: "STARS" | "SBP") {
    return this.request<PaymentResult>("/api/v1/payments", {
      method: "POST",
      body: JSON.stringify({ order_id: orderId, method }),
      headers: { "Idempotency-Key": crypto.randomUUID() },
    });
  }

  getFavorites() {
    return this.request<{ product_ids: string[] }>("/api/v1/favorites");
  }

  addFavorite(productId: string) {
    return this.request(`/api/v1/favorites/${productId}`, { method: "POST" });
  }

  removeFavorite(productId: string) {
    return this.request(`/api/v1/favorites/${productId}`, { method: "DELETE" });
  }

  getReferrals() {
    return this.request<ReferralProfile>("/api/v1/referrals/me");
  }
}

export const api = new ApiClient();

export type User = {
  id: string;
  telegram_id: number;
  username?: string;
  first_name?: string;
  last_name?: string;
  photo_url?: string;
};

export type Category = {
  id: string;
  slug: string;
  name: string;
  description?: string;
};

export type DeliveryConfig = {
  type?: string;
  amount?: number;
  duration_days?: number;
  gift_id?: string;
  telegram_gift_id?: string;
  star_count?: number;
  total_count?: number;
  remaining_count?: number;
  telegram_synced?: boolean;
  nft_slug?: string;
  gift_num?: number;
  base_gift_id?: number;
  title?: string;
  sticker_url?: string;
};

export type Product = {
  id: string;
  slug: string;
  name: string;
  description?: string;
  price_kopecks: number;
  sale_price_kopecks?: number | null;
  on_sale?: boolean;
  currency?: string;
  product_type: string;
  image_url?: string;
  popularity_score: number;
  categories?: Category[];
  delivery_config?: DeliveryConfig;
  created_at?: string;
};

export type ProductListResult = {
  items: Product[];
  total: number;
  page: number;
  limit: number;
};

export type Promotion = {
  id: string;
  title: string;
  description?: string;
  discount_type: string;
  discount_value: number;
};

export type CartItem = { product_id: string; quantity: number };
export type Cart = { items: CartItem[]; updated_at?: string };

export type Order = {
  id: string;
  status: string;
  total_kopecks: number;
  discount_kopecks?: number;
  items?: OrderItem[];
  created_at: string;
};

export type OrderItem = {
  id: string;
  name: string;
  price_kopecks: number;
  quantity: number;
  product_type: string;
};

export type PaymentResult = {
  payment_id: string;
  payment_url?: string;
  status: string;
  method: string;
  amount_kopecks: number;
};

export type ReferralReward = {
  id: string;
  promo_code: string;
  discount_type: string;
  discount_value: number;
  is_used: boolean;
  reward_type: string;
};

export type ReferralProfile = {
  code: string;
  link: string;
  invited_count: number;
  qualified_count: number;
  pending_count: number;
  total_bonus_kopecks: number;
  rewards: ReferralReward[];
  bonus_info: {
    referred_welcome: string;
    referrer_reward: string;
  };
};
