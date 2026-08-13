const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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

    const res = await fetch(`${API_URL}${path}`, { ...options, headers });
    if (!res.ok) {
      const err = (await res.json().catch(() => ({}))) as ApiError;
      throw new Error(err.error?.message || `HTTP ${res.status}`);
    }
    if (res.status === 204) return {} as T;
    return res.json();
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

export type Product = {
  id: string;
  slug: string;
  name: string;
  description?: string;
  price_kopecks: number;
  product_type: string;
  image_url?: string;
  popularity_score: number;
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
