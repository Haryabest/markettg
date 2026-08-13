package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Payment struct {
	ID                uuid.UUID       `json:"id"`
	OrderID           uuid.UUID       `json:"order_id"`
	UserID            uuid.UUID       `json:"user_id"`
	Method            string          `json:"method"`
	AmountKopecks     int64           `json:"amount_kopecks"`
	Status            string          `json:"status"`
	ProviderPaymentID *string         `json:"provider_payment_id"`
	IdempotencyKey    string          `json:"idempotency_key"`
	Metadata          json.RawMessage `json:"metadata"`
	CreatedAt         time.Time       `json:"created_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByIdempotencyKey(ctx context.Context, key string) (*Payment, error) {
	var p Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, user_id, method::text, amount_kopecks, status::text,
		       provider_payment_id, idempotency_key, metadata, created_at
		FROM payments.payments WHERE idempotency_key = $1`, key,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Method, &p.AmountKopecks, &p.Status,
		&p.ProviderPaymentID, &p.IdempotencyKey, &p.Metadata, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Create(ctx context.Context, p Payment) (*Payment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO payments.payments (order_id, user_id, method, amount_kopecks, status, provider_payment_id, idempotency_key, metadata)
		VALUES ($1, $2, $3::payments.payment_method, $4, $5::payments.payment_status, $6, $7, $8)
		RETURNING id, created_at`,
		p.OrderID, p.UserID, p.Method, p.AmountKopecks, p.Status, p.ProviderPaymentID, p.IdempotencyKey, p.Metadata,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	eventPayload, _ := json.Marshal(map[string]string{"payment_id": p.ID.String(), "order_id": p.OrderID.String()})
	_, err = tx.Exec(ctx, `
		INSERT INTO payments.outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, 'payment', $2, 'PaymentCreated', $3)`,
		uuid.New(), p.ID.String(), eventPayload)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) MarkPaid(ctx context.Context, paymentID uuid.UUID, providerPaymentID string, eventID string, rawPayload []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM payments.payment_events WHERE provider_event_id = $1)`, eventID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, `
		UPDATE payments.payments SET status = 'PAID', provider_payment_id = $2, updated_at = NOW()
		WHERE id = $1 AND status != 'PAID'`, paymentID, providerPaymentID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments.payment_events (payment_id, event_type, provider_event_id, raw_payload)
		VALUES ($1, 'PAID', $2, $3)`, paymentID, eventID, rawPayload)
	if err != nil {
		return err
	}

	var orderID, userID uuid.UUID
	var amount int64
	var method string
	_ = tx.QueryRow(ctx, `SELECT order_id, user_id, amount_kopecks, method::text FROM payments.payments WHERE id = $1`, paymentID).
		Scan(&orderID, &userID, &amount, &method)

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"payment_id": paymentID.String(), "order_id": orderID.String(),
		"user_id": userID.String(), "amount_kopecks": amount, "method": method,
	})
	_, err = tx.Exec(ctx, `
		INSERT INTO payments.outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, 'payment', $2, 'PaymentSucceeded', $3)`,
		uuid.New(), paymentID.String(), eventPayload)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	var p Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, user_id, method::text, amount_kopecks, status::text,
		       provider_payment_id, idempotency_key, metadata, created_at
		FROM payments.payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Method, &p.AmountKopecks, &p.Status,
		&p.ProviderPaymentID, &p.IdempotencyKey, &p.Metadata, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) ListAll(ctx context.Context, limit, offset int) ([]Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, user_id, method::text, amount_kopecks, status::text,
		       provider_payment_id, idempotency_key, metadata, created_at
		FROM payments.payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Method, &p.AmountKopecks, &p.Status,
			&p.ProviderPaymentID, &p.IdempotencyKey, &p.Metadata, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
