package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/packages/go-shared/pkg/events"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/logger"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/markettg/markettg/packages/go-shared/pkg/stream"
	delhandler "github.com/markettg/markettg/services/delivery-service/internal/handler"
	"github.com/markettg/markettg/services/delivery-service/internal/repository"
	"github.com/markettg/markettg/services/delivery-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("delivery-service")
	base := config.LoadBase("8085")

	pool, err := pgxpool.New(context.Background(), config.LoadPostgres().URL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	redisClient := redisutil.NewClient(redisutil.Config{
		Addr: config.LoadRedis().Addr, Password: config.LoadRedis().Password, DB: 0,
	})

	repo := repository.New(pool)
	svc := service.New(repo, redisClient, log,
		delhandler.NewStarsDeliveryHandler(),
		delhandler.NewPremiumDeliveryHandler(),
		delhandler.NewGiftDeliveryHandler(),
	)
	h := delhandler.New(svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.RunWorker(ctx)

	consumer := stream.NewConsumer(redisClient, redisutil.StreamPayments, "delivery-workers", "delivery-1")
	_ = consumer.EnsureGroup(ctx)
	go consumer.Run(ctx, func(c context.Context, eventID string, payload []byte) error {
		var event struct {
			EventType string          `json:"event_type"`
			Payload   json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		if event.EventType == events.PaymentSucceeded {
			var p events.PaymentSucceededPayload
			if err := json.Unmarshal(event.Payload, &p); err != nil {
				return err
			}
			return svc.ProcessPaymentSucceeded(c, eventID, p)
		}
		return nil
	})

	app := fiber.New(fiber.Config{AppName: "delivery-service", ErrorHandler: httputil.Error})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "delivery-service"})
	})

	api := app.Group("/api/v1")
	api.Get("/admin/deliveries", h.ListDeliveries)
	api.Post("/admin/deliveries/:id/retry", h.RetryDelivery)
	api.Post("/internal/deliveries", h.CreateJob)

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting delivery-service", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	time.Sleep(time.Second)
	_ = app.Shutdown()
}
