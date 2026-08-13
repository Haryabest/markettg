package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/logger"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/markettg/markettg/services/notification-service/internal/bot"
	"github.com/markettg/markettg/services/notification-service/internal/handler"
	"github.com/markettg/markettg/services/notification-service/internal/repository"
	"github.com/markettg/markettg/services/notification-service/internal/service"
	"github.com/markettg/markettg/services/notification-service/internal/ws"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("notification-service")
	base := config.LoadBase("8086")

	pool, err := pgxpool.New(context.Background(), config.LoadPostgres().URL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	redisClient := redisutil.NewClient(redisutil.Config{
		Addr: config.LoadRedis().Addr, Password: config.LoadRedis().Password, DB: 0,
	})

	hub := ws.NewHub()
	repo := repository.New(pool)
	botClient := bot.NewClient()
	svc := service.New(repo, hub, botClient, redisClient, log)
	h := handler.New(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.RunConsumers(ctx)

	app := fiber.New(fiber.Config{AppName: "notification-service", ErrorHandler: httputil.Error})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "notification-service"})
	})

	api := app.Group("/api/v1")
	api.Get("/ws/orders/:id", handler.WebSocketUpgrade, websocket.New(h.OrderWebSocket))

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting notification-service", zap.String("addr", addr))
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
