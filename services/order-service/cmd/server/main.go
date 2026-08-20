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
	"github.com/markettg/markettg/services/order-service/internal/cart"
	"github.com/markettg/markettg/services/order-service/internal/catalog"
	"github.com/markettg/markettg/services/order-service/internal/handler"
	"github.com/markettg/markettg/services/order-service/internal/repository"
	"github.com/markettg/markettg/services/order-service/internal/service"
	"github.com/markettg/markettg/services/order-service/internal/user"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("order-service")
	base := config.LoadBase("8083")
	pgCfg := config.LoadPostgres()
	redisCfg := config.LoadRedis()

	pool, err := pgxpool.New(context.Background(), pgCfg.URL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	redisClient := redisutil.NewClient(redisutil.Config{
		Addr: redisCfg.Addr, Password: redisCfg.Password, DB: redisCfg.DB,
	})

	catalogClient := catalog.NewClient(config.GetEnv("CATALOG_SERVICE_URL", "http://catalog-service:8082"))
	cartStore := cart.New(redisClient)
	repo := repository.New(pool)
	userClient := user.NewClient(
		config.GetEnv("USER_SERVICE_URL", "http://user-service:8081"),
		config.GetEnv("BOT_INTERNAL_SECRET", "bot-secret"),
	)
	svc := service.New(repo, cartStore, catalogClient, redisClient, userClient)
	h := handler.New(svc, userClient)

	publisher := outbox.NewPublisher(pool, redisClient, "orders", redisutil.StreamOrders)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go publisher.Run(ctx, 50, 2*time.Second)

	app := fiber.New(fiber.Config{AppName: "order-service", ErrorHandler: httputil.Error})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "order-service"})
	})

	api := app.Group("/api/v1")
	api.Get("/cart", h.GetCart)
	api.Put("/cart/items", h.UpdateCartItem)
	api.Delete("/cart/items/:productId", h.RemoveCartItem)
	api.Post("/orders", h.CreateOrder)
	api.Get("/orders", h.ListOrders)
	api.Get("/orders/:id", h.GetOrder)
	api.Post("/orders/:id/cancel", h.CancelOrder)
	api.Get("/internal/bot/orders", h.BotListOrders)
	api.Patch("/internal/orders/:id/status", h.UpdateOrderStatus)
	api.Get("/internal/orders/:id", h.GetOrderInternal)

	admin := api.Group("/admin")
	admin.Get("/orders", h.AdminListOrders)
	admin.Get("/promo-codes", h.AdminListPromoCodes)
	admin.Post("/promo-codes", h.AdminCreatePromoCode)

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting order-service", zap.String("addr", addr))
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
