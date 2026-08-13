package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Event struct {
	ID            uuid.UUID       `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Publisher struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
	schema string
	stream string
}

func NewPublisher(pool *pgxpool.Pool, redis *redis.Client, schema, stream string) *Publisher {
	return &Publisher{pool: pool, redis: redis, schema: schema, stream: stream}
}

func (p *Publisher) InsertTx(ctx context.Context, tx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (interface{}, error)
}, aggregateType, aggregateID, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO `+p.schema+`.outbox_events (id, aggregate_type, aggregate_id, event_type, payload, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		uuid.New(), aggregateType, aggregateID, eventType, data,
	)
	return err
}

func (p *Publisher) Run(ctx context.Context, batchSize int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.publishBatch(ctx, batchSize)
		}
	}
}

func (p *Publisher) publishBatch(ctx context.Context, batchSize int) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at
		FROM `+p.schema+`.outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, batchSize)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			continue
		}

		data, _ := json.Marshal(e)
		if err := p.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: p.stream,
			Values: map[string]interface{}{"event": string(data)},
		}).Err(); err != nil {
			continue
		}

		_, _ = p.pool.Exec(ctx,
			`UPDATE `+p.schema+`.outbox_events SET published_at = NOW() WHERE id = $1`, e.ID)
	}
}
