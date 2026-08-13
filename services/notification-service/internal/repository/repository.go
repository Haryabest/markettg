package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"`
	Channel   string          `json:"channel"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	SentAt    *time.Time      `json:"sent_at"`
	CreatedAt time.Time       `json:"created_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, n Notification) (*Notification, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notifications.notifications (user_id, type, channel, payload, status)
		VALUES ($1, $2, $3::notifications.notification_channel, $4, 'PENDING')
		RETURNING id, created_at`,
		n.UserID, n.Type, n.Channel, n.Payload,
	).Scan(&n.ID, &n.CreatedAt)
	return &n, err
}

func (r *Repository) MarkSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications.notifications SET status = 'SENT', sent_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notifications.processed_events WHERE event_id = $1)`, eventID).Scan(&exists)
	return exists, err
}

func (r *Repository) MarkEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO notifications.processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`, eventID)
	return err
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, type, channel::text, payload, status::text, sent_at, created_at
		FROM notifications.notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Payload, &n.Status, &n.SentAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}
