package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/events"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/markettg/markettg/services/notification-service/internal/bot"
	"github.com/markettg/markettg/services/notification-service/internal/repository"
	"github.com/markettg/markettg/services/notification-service/internal/ws"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service struct {
	repo *repository.Repository
	hub  *ws.Hub
	bot  *bot.Client
	redis *redis.Client
	log  *zap.Logger
}

func New(repo *repository.Repository, hub *ws.Hub, botClient *bot.Client, redis *redis.Client, log *zap.Logger) *Service {
	return &Service{repo: repo, hub: hub, bot: botClient, redis: redis, log: log}
}

func (s *Service) NotifyOrderStatus(ctx context.Context, orderID, userID, status string, telegramID int64) error {
	s.hub.BroadcastOrderStatus(orderID, userID, status)

	payload, _ := json.Marshal(map[string]string{"order_id": orderID, "status": status})
	n, err := s.repo.Create(ctx, repository.Notification{
		UserID: mustParseUUID(userID), Type: "order_status", Channel: "WEBSOCKET", Payload: payload,
	})
	if err != nil {
		return err
	}
	_ = s.repo.MarkSent(ctx, n.ID)

	if telegramID > 0 {
		msg := fmt.Sprintf("Статус заказа #%s: %s", orderID[:8], status)
		_ = s.bot.SendMessage(ctx, telegramID, msg)
	}
	return nil
}

func (s *Service) ProcessEvent(ctx context.Context, eventID string, eventType string, payload json.RawMessage) error {
	processed, err := s.repo.IsEventProcessed(ctx, eventID)
	if err != nil || processed {
		return err
	}

	switch eventType {
	case events.PaymentSucceeded:
		var p events.PaymentSucceededPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		_ = s.NotifyOrderStatus(ctx, p.OrderID, p.UserID, "PAID", 0)
	case events.DeliveryCompleted:
		var p events.DeliveryCompletedPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		_ = s.NotifyOrderStatus(ctx, p.OrderID, p.UserID, "COMPLETED", 0)
	case events.OrderCreated:
		var p events.OrderCreatedPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		_ = s.NotifyOrderStatus(ctx, p.OrderID, p.UserID, "PAYMENT_PENDING", 0)
	}

	return s.repo.MarkEventProcessed(ctx, eventID)
}

func (s *Service) RunConsumers(ctx context.Context) {
	streams := []string{redisutil.StreamPayments, redisutil.StreamOrders, redisutil.StreamDelivery}
	for _, stream := range streams {
		go s.consumeStream(ctx, stream)
	}
}

func (s *Service) consumeStream(ctx context.Context, streamName string) {
	lastID := "0"
	for {
		select {
		case <-ctx.Done():
			return
		default:
			streams, err := s.redis.XRead(ctx, &redis.XReadArgs{
				Streams: []string{streamName, lastID},
				Count:   10,
				Block:   5 * 1000000000, // 5s in nanoseconds - actually use time.Second
			}).Result()
			if err != nil {
				continue
			}
			for _, stream := range streams {
				for _, msg := range stream.Messages {
					lastID = msg.ID
					eventRaw, _ := msg.Values["event"].(string)
					var event struct {
						EventType string          `json:"event_type"`
						Payload   json.RawMessage `json:"payload"`
					}
					if err := json.Unmarshal([]byte(eventRaw), &event); err != nil {
						continue
					}
					_ = s.ProcessEvent(ctx, msg.ID, event.EventType, event.Payload)
				}
			}
		}
	}
}

func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
