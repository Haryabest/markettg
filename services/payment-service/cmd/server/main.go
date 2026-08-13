package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/logger"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
	"github.com/markettg/markettg/packages/go-shared/pkg/outbox"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/markettg/markettg/services/payment-service/internal/handler"
	"github.com/markettg/markettg/services/payment-service/internal/order"
	"github.com/markettg/markettg/services/payment-service/internal/provider"
	"github.com/markettg/markettg/services/payment-service/internal/repository"
	"github.com/markettg/markettg/services/payment-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("payment-service")
	base := config.LoadBase("8084")

	pool, err := pgxpool.New(context.Background(), config.LoadPostgres().URL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	redisClient := redisutil.NewClient(redisutil.Config{
		Addr: config.LoadRedis().Addr, Password: config.LoadRedis().Password, DB: 0,
	})

	orderClient := order.NewClient(config.GetEnv("ORDER_SERVICE_URL", "http://order-service:8083"))
	repo := repository.New(pool)
	svc := service.New(repo, orderClient,
		provider.NewTelegramStarsProvider(),
		provider.NewSBPProvider(),
	)
	h := handler.New(svc)

	publisher := outbox.NewPublisher(pool, redisClient, "payments", redisutil.StreamPayments)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go publisher.Run(ctx, 50, 2*time.Second)

	app := fiber.New(fiber.Config{AppName: "payment-service", ErrorHandler: httputil.Error})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "payment-service"})
	})

	api := app.Group("/api/v1")
	api.Post("/payments", h.CreatePayment)
	api.Post("/payments/webhooks/telegram", h.TelegramWebhook)
	api.Post("/payments/webhooks/sbp", h.SBPWebhook)
	api.Get("/admin/payments", h.AdminListPayments)
	api.Get("/payments/:id", h.GetPayment)

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting payment-service", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	_ = app.Shutdown()
}
