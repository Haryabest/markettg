package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryJob struct {
	ID             uuid.UUID       `json:"id"`
	OrderID        uuid.UUID       `json:"order_id"`
	OrderItemID    uuid.UUID       `json:"order_item_id"`
	UserID         uuid.UUID       `json:"user_id"`
	TelegramID     int64           `json:"telegram_id"`
	HandlerType    string          `json:"handler_type"`
	DeliveryConfig json.RawMessage `json:"delivery_config"`
	Status         string          `json:"status"`
	Attempts       int             `json:"attempts"`
	MaxAttempts    int             `json:"max_attempts"`
	LastError      *string         `json:"last_error"`
	CompletedAt    *time.Time      `json:"completed_at"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM delivery.processed_events WHERE event_id = $1)`, eventID).Scan(&exists)
	return exists, err
}

func (r *Repository) MarkEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO delivery.processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`, eventID)
	return err
}

func (r *Repository) CreateJob(ctx context.Context, job DeliveryJob) (*DeliveryJob, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO delivery.delivery_jobs (order_id, order_item_id, user_id, telegram_id, handler_type, delivery_config, status, max_attempts)
		VALUES ($1, $2, $3, $4, $5, $6, 'PENDING', $7)
		ON CONFLICT (order_item_id) DO UPDATE SET updated_at = NOW()
		RETURNING id, status, attempts, max_attempts, created_at`,
		job.OrderID, job.OrderItemID, job.UserID, job.TelegramID, job.HandlerType, job.DeliveryConfig, job.MaxAttempts,
	).Scan(&job.ID, &job.Status, &job.Attempts, &job.MaxAttempts, &job.CreatedAt)
	return &job, err
}

func (r *Repository) GetPendingJobs(ctx context.Context, limit int) ([]DeliveryJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, order_item_id, user_id, telegram_id, handler_type, delivery_config,
		       status::text, attempts, max_attempts, last_error, completed_at, created_at
		FROM delivery.delivery_jobs
		WHERE status IN ('PENDING', 'PROCESSING') AND attempts < max_attempts
		ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []DeliveryJob
	for rows.Next() {
		var j DeliveryJob
		if err := rows.Scan(&j.ID, &j.OrderID, &j.OrderItemID, &j.UserID, &j.TelegramID,
			&j.HandlerType, &j.DeliveryConfig, &j.Status, &j.Attempts, &j.MaxAttempts,
			&j.LastError, &j.CompletedAt, &j.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (r *Repository) RecordAttempt(ctx context.Context, jobID uuid.UUID, attemptNo int, req, resp json.RawMessage, errMsg *string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO delivery.delivery_attempts (job_id, attempt_no, request_payload, response_payload, error_message)
		VALUES ($1, $2, $3, $4, $5)`, jobID, attemptNo, req, resp, errMsg)
	return err
}

func (r *Repository) MarkCompleted(ctx context.Context, jobID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE delivery.delivery_jobs SET status = 'COMPLETED', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1`, jobID)
	return err
}

func (r *Repository) MarkFailed(ctx context.Context, jobID uuid.UUID, errMsg string, incrementAttempt bool) error {
	if incrementAttempt {
		_, err := r.pool.Exec(ctx, `
			UPDATE delivery.delivery_jobs SET status = 'PENDING', attempts = attempts + 1,
			last_error = $2, updated_at = NOW() WHERE id = $1`, jobID, errMsg)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE delivery.delivery_jobs SET status = 'FAILED', last_error = $2, updated_at = NOW()
		WHERE id = $1`, jobID, errMsg)
	return err
}

func (r *Repository) ListAll(ctx context.Context, limit, offset int) ([]DeliveryJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, order_item_id, user_id, telegram_id, handler_type, delivery_config,
		       status::text, attempts, max_attempts, last_error, completed_at, created_at
		FROM delivery.delivery_jobs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []DeliveryJob
	for rows.Next() {
		var j DeliveryJob
		if err := rows.Scan(&j.ID, &j.OrderID, &j.OrderItemID, &j.UserID, &j.TelegramID,
			&j.HandlerType, &j.DeliveryConfig, &j.Status, &j.Attempts, &j.MaxAttempts,
			&j.LastError, &j.CompletedAt, &j.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (r *Repository) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]DeliveryJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, order_item_id, user_id, telegram_id, handler_type, delivery_config,
		       status::text, attempts, max_attempts, last_error, completed_at, created_at
		FROM delivery.delivery_jobs WHERE order_id = $1 ORDER BY created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []DeliveryJob
	for rows.Next() {
		var j DeliveryJob
		if err := rows.Scan(&j.ID, &j.OrderID, &j.OrderItemID, &j.UserID, &j.TelegramID,
			&j.HandlerType, &j.DeliveryConfig, &j.Status, &j.Attempts, &j.MaxAttempts,
			&j.LastError, &j.CompletedAt, &j.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (r *Repository) GetJob(ctx context.Context, id uuid.UUID) (*DeliveryJob, error) {
	var j DeliveryJob
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, order_item_id, user_id, telegram_id, handler_type, delivery_config,
		       status::text, attempts, max_attempts, last_error, completed_at, created_at
		FROM delivery.delivery_jobs WHERE id = $1`, id,
	).Scan(&j.ID, &j.OrderID, &j.OrderItemID, &j.UserID, &j.TelegramID,
		&j.HandlerType, &j.DeliveryConfig, &j.Status, &j.Attempts, &j.MaxAttempts,
		&j.LastError, &j.CompletedAt, &j.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &j, err
}
