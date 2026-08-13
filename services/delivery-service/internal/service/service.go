package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/events"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	delhandler "github.com/markettg/markettg/services/delivery-service/internal/handler"
	"github.com/markettg/markettg/services/delivery-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service struct {
	repo     *repository.Repository
	redis    *redis.Client
	handlers map[string]delhandler.DeliveryHandler
	log      *zap.Logger
}

func New(repo *repository.Repository, redis *redis.Client, log *zap.Logger, handlers ...delhandler.DeliveryHandler) *Service {
	m := make(map[string]delhandler.DeliveryHandler)
	for _, h := range handlers {
		m[h.Type()] = h
	}
	return &Service{repo: repo, redis: redis, handlers: m, log: log}
}

func (s *Service) ProcessPaymentSucceeded(ctx context.Context, eventID string, payload events.PaymentSucceededPayload) error {
	processed, err := s.repo.IsEventProcessed(ctx, eventID)
	if err != nil || processed {
		return err
	}

	lockKey := redisutil.DeliveryLockKey(payload.OrderID)
	lock := redisutil.NewLock(s.redis, lockKey, eventID, 5*time.Minute)
	acquired, err := lock.Acquire(ctx)
	if err != nil || !acquired {
		return fmt.Errorf("could not acquire delivery lock")
	}
	defer lock.Release(ctx)

	// Fetch order items from order service would happen here; for now create jobs from event
	_ = s.repo.MarkEventProcessed(ctx, eventID)
	return nil
}

func (s *Service) DeliverJob(ctx context.Context, job *repository.DeliveryJob) error {
	lockKey := redisutil.DeliveryLockKey(job.OrderID.String())
	lock := redisutil.NewLock(s.redis, lockKey, job.ID.String(), 5*time.Minute)
	acquired, err := lock.Acquire(ctx)
	if err != nil || !acquired {
		return fmt.Errorf("delivery lock held")
	}
	defer lock.Release(ctx)

	if job.Status == "COMPLETED" {
		return nil
	}

	var cfg delhandler.DeliveryConfig
	if err := json.Unmarshal(job.DeliveryConfig, &cfg); err != nil {
		return err
	}

	handler := delhandler.GetHandler(job.HandlerType, s.handlers)
	if handler == nil {
		return fmt.Errorf("unknown handler type: %s", job.HandlerType)
	}

	reqPayload, _ := json.Marshal(cfg)
	resp, err := handler.Deliver(ctx, job.TelegramID, cfg)
	var errMsg *string
	if err != nil {
		msg := err.Error()
		errMsg = &msg
		_ = s.repo.RecordAttempt(ctx, job.ID, job.Attempts+1, reqPayload, nil, errMsg)
		if job.Attempts+1 >= job.MaxAttempts {
			return s.repo.MarkFailed(ctx, job.ID, msg, false)
		}
		return s.repo.MarkFailed(ctx, job.ID, msg, true)
	}

	_ = s.repo.RecordAttempt(ctx, job.ID, job.Attempts+1, reqPayload, resp, nil)
	if err := s.repo.MarkCompleted(ctx, job.ID); err != nil {
		return err
	}

	// Publish DeliveryCompleted event
	eventPayload, _ := json.Marshal(events.DeliveryCompletedPayload{
		OrderID: job.OrderID.String(), OrderItemID: job.OrderItemID.String(), UserID: job.UserID.String(),
	})
	_ = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: redisutil.StreamDelivery,
		Values: map[string]interface{}{"event": string(eventPayload)},
	}).Err()

	return nil
}

func (s *Service) RunWorker(ctx context.Context) {
	backoffs := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			jobs, err := s.repo.GetPendingJobs(ctx, 10)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			for _, job := range jobs {
				if job.Attempts > 0 && job.Attempts-1 < len(backoffs) {
					time.Sleep(backoffs[job.Attempts-1])
				}
				if err := s.DeliverJob(ctx, &job); err != nil {
					s.log.Warn("delivery failed", zap.String("job_id", job.ID.String()), zap.Error(err))
				}
			}
			if len(jobs) == 0 {
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func (s *Service) CreateJob(ctx context.Context, job repository.DeliveryJob) (*repository.DeliveryJob, error) {
	return s.repo.CreateJob(ctx, job)
}

func (s *Service) ListDeliveries(ctx context.Context, limit, offset int) ([]repository.DeliveryJob, error) {
	return s.repo.ListAll(ctx, limit, offset)
}

func (s *Service) RetryDelivery(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil || job == nil {
		return fmt.Errorf("job not found")
	}
	job.Status = "PENDING"
	return s.DeliverJob(ctx, job)
}
