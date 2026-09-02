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
	"github.com/markettg/markettg/services/delivery-service/internal/order"
	"github.com/markettg/markettg/services/delivery-service/internal/repository"
	"github.com/markettg/markettg/services/delivery-service/internal/userclient"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service struct {
	repo     *repository.Repository
	redis    *redis.Client
	handlers map[string]delhandler.DeliveryHandler
	orders   *order.Client
	users    *userclient.Client
	log      *zap.Logger
}

func New(repo *repository.Repository, redis *redis.Client, log *zap.Logger, orderClient *order.Client, userClient *userclient.Client, handlers ...delhandler.DeliveryHandler) *Service {
	m := make(map[string]delhandler.DeliveryHandler)
	for _, h := range handlers {
		m[h.Type()] = h
	}
	return &Service{repo: repo, redis: redis, handlers: m, orders: orderClient, users: userClient, log: log}
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

	if s.orders == nil {
		_ = s.repo.MarkEventProcessed(ctx, eventID)
		return nil
	}

	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return err
	}
	ord, err := s.orders.GetOrder(ctx, orderID)
	if err != nil || ord == nil {
		return fmt.Errorf("order not found: %s", payload.OrderID)
	}

	var telegramID int64
	if s.users != nil {
		if u, err := s.users.GetByID(ctx, ord.UserID); err == nil && u != nil {
			telegramID = u.TelegramID
		}
	}

	for _, item := range ord.Items {
		handlerType := item.ProductType
		if handlerType == "NFT" {
			handlerType = "GIFT"
		}
		for q := 0; q < item.Quantity; q++ {
			_, err := s.repo.CreateJob(ctx, repository.DeliveryJob{
				OrderID:        ord.ID,
				OrderItemID:    item.ID,
				UserID:         ord.UserID,
				TelegramID:     telegramID,
				HandlerType:    handlerType,
				DeliveryConfig: item.DeliveryConfig,
				MaxAttempts:    5,
			})
			if err != nil {
				s.log.Warn("failed to create delivery job", zap.String("order_id", ord.ID.String()), zap.Error(err))
			}
		}
	}

	_ = s.orders.UpdateStatus(ctx, orderID, "DELIVERY_PENDING")
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

func (s *Service) ConfirmDelivery(ctx context.Context, jobID uuid.UUID) (*repository.DeliveryJob, error) {
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil || job == nil {
		return nil, fmt.Errorf("job not found")
	}
	if job.Status == "COMPLETED" {
		return job, nil
	}
	if err := s.repo.MarkCompleted(ctx, jobID); err != nil {
		return nil, err
	}
	job.Status = "COMPLETED"
	now := time.Now()
	job.CompletedAt = &now
	s.maybeCompleteOrder(ctx, job.OrderID)
	return job, nil
}

func (s *Service) ConfirmOrderDeliveries(ctx context.Context, orderID uuid.UUID) error {
	jobs, err := s.repo.ListByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Status != "COMPLETED" {
			if err := s.repo.MarkCompleted(ctx, job.ID); err != nil {
				return err
			}
		}
	}
	if s.orders != nil {
		_ = s.orders.UpdateStatus(ctx, orderID, "COMPLETED")
	}
	return nil
}

func (s *Service) maybeCompleteOrder(ctx context.Context, orderID uuid.UUID) {
	jobs, err := s.repo.ListByOrderID(ctx, orderID)
	if err != nil || len(jobs) == 0 {
		return
	}
	for _, job := range jobs {
		if job.Status != "COMPLETED" {
			return
		}
	}
	if s.orders != nil {
		_ = s.orders.UpdateStatus(ctx, orderID, "COMPLETED")
	}
}
