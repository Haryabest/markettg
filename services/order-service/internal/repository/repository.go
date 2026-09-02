package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Order struct {
	ID              uuid.UUID   `json:"id"`
	UserID          uuid.UUID   `json:"user_id"`
	Status          string      `json:"status"`
	TotalKopecks    int64       `json:"total_kopecks"`
	DiscountKopecks int64       `json:"discount_kopecks"`
	PromoCodeID     *uuid.UUID  `json:"promo_code_id"`
	IdempotencyKey  *string     `json:"idempotency_key,omitempty"`
	Items           []OrderItem `json:"items,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID             uuid.UUID       `json:"id"`
	OrderID        uuid.UUID       `json:"order_id"`
	ProductID      uuid.UUID       `json:"product_id"`
	Name           string          `json:"name"`
	PriceKopecks   int64           `json:"price_kopecks"`
	DeliveryConfig json.RawMessage `json:"delivery_config"`
	ProductType    string          `json:"product_type"`
	Quantity       int             `json:"quantity"`
}

type PromoCode struct {
	ID            uuid.UUID  `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue int64      `json:"discount_value"`
	MaxUses       *int       `json:"max_uses"`
	UsedCount     int        `json:"used_count"`
	ValidFrom     time.Time  `json:"valid_from"`
	ValidTo       *time.Time `json:"valid_to"`
	IsActive      bool       `json:"is_active"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateOrder(ctx context.Context, order Order, items []OrderItem) (*Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO orders.orders (user_id, status, total_kopecks, discount_kopecks, promo_code_id, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		order.UserID, order.Status, order.TotalKopecks, order.DiscountKopecks, order.PromoCodeID, order.IdempotencyKey,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		_, err = tx.Exec(ctx, `
			INSERT INTO orders.order_items (order_id, product_id, name, price_kopecks, delivery_config, product_type, quantity)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			order.ID, item.ProductID, item.Name, item.PriceKopecks, item.DeliveryConfig, item.ProductType, item.Quantity)
		if err != nil {
			return nil, err
		}
	}

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID.String(),
		"user_id":  order.UserID.String(),
		"total":    order.TotalKopecks,
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO orders.outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, 'order', $2, 'OrderCreated', $3)`,
		uuid.New(), order.ID.String(), eventPayload)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *Repository) GetOrder(ctx context.Context, id uuid.UUID) (*Order, error) {
	var o Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status::text, total_kopecks, discount_kopecks, promo_code_id, created_at, updated_at
		FROM orders.orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalKopecks, &o.DiscountKopecks, &o.PromoCodeID, &o.CreatedAt, &o.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *Repository) getOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, product_id, name, price_kopecks, delivery_config, product_type, quantity
		FROM orders.order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Name,
			&item.PriceKopecks, &item.DeliveryConfig, &item.ProductType, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, status::text, total_kopecks, discount_kopecks, promo_code_id, created_at, updated_at
		FROM orders.orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalKopecks, &o.DiscountKopecks,
			&o.PromoCodeID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders.orders SET status = $2::orders.order_status, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (r *Repository) GetPromoCode(ctx context.Context, code string) (*PromoCode, error) {
	var p PromoCode
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, discount_type::text, discount_value, max_uses, used_count, valid_from, valid_to, is_active
		FROM orders.promo_codes WHERE code = $1 AND is_active = TRUE`, code,
	).Scan(&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.MaxUses, &p.UsedCount, &p.ValidFrom, &p.ValidTo, &p.IsActive)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) UsePromoCode(ctx context.Context, promoID, userID, orderID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var usedCount int
	var maxUses *int
	err = tx.QueryRow(ctx, `SELECT used_count, max_uses FROM orders.promo_codes WHERE id = $1 FOR UPDATE`, promoID).Scan(&usedCount, &maxUses)
	if err != nil {
		return err
	}
	if maxUses != nil && usedCount >= *maxUses {
		return pgx.ErrNoRows
	}

	_, err = tx.Exec(ctx, `UPDATE orders.promo_codes SET used_count = used_count + 1 WHERE id = $1`, promoID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO orders.promo_code_usages (promo_code_id, user_id, order_id) VALUES ($1, $2, $3)`, promoID, userID, orderID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListAllOrders(ctx context.Context, limit, offset int) ([]Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, status::text, total_kopecks, discount_kopecks, promo_code_id, created_at, updated_at
		FROM orders.orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalKopecks, &o.DiscountKopecks,
			&o.PromoCodeID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *Repository) ListAllOrdersWithItems(ctx context.Context, limit, offset int) ([]Order, error) {
	orders, err := r.ListAllOrders(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		items, err := r.getOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}
	return orders, nil
}

func (r *Repository) ListPromoCodes(ctx context.Context) ([]PromoCode, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, discount_type::text, discount_value, max_uses, used_count, valid_from, valid_to, is_active
		FROM orders.promo_codes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []PromoCode
	for rows.Next() {
		var p PromoCode
		if err := rows.Scan(&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.MaxUses,
			&p.UsedCount, &p.ValidFrom, &p.ValidTo, &p.IsActive); err != nil {
			return nil, err
		}
		codes = append(codes, p)
	}
	return codes, nil
}

func (r *Repository) CreatePromoCode(ctx context.Context, p PromoCode) (*PromoCode, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO orders.promo_codes (code, discount_type, discount_value, max_uses, valid_from, valid_to)
		VALUES ($1, $2::orders.discount_type, $3, $4, $5, $6)
		RETURNING id, code, discount_type::text, discount_value, max_uses, used_count, valid_from, valid_to, is_active`,
		p.Code, p.DiscountType, p.DiscountValue, p.MaxUses, p.ValidFrom, p.ValidTo,
	).Scan(&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.MaxUses, &p.UsedCount, &p.ValidFrom, &p.ValidTo, &p.IsActive)
	return &p, err
}
