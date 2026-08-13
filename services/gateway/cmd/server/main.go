package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/logger"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/markettg/markettg/services/gateway/internal/proxy"
	gwmw "github.com/markettg/markettg/services/gateway/internal/middleware"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("gateway")
	base := config.LoadBase("8080")

	redisCfg := config.LoadRedis()
	redisClient := redisutil.NewClient(redisutil.Config{
		Addr: redisCfg.Addr, Password: redisCfg.Password, DB: redisCfg.DB,
	})
	ctx := context.Background()
	if err := redisutil.Ping(ctx, redisClient); err != nil {
		log.Warn("redis not available at startup", zap.Error(err))
	}

	services := proxy.ServiceURLs{
		User:         config.GetEnv("USER_SERVICE_URL", "http://user-service:8081"),
		Catalog:      config.GetEnv("CATALOG_SERVICE_URL", "http://catalog-service:8082"),
		Order:        config.GetEnv("ORDER_SERVICE_URL", "http://order-service:8083"),
		Payment:      config.GetEnv("PAYMENT_SERVICE_URL", "http://payment-service:8084"),
		Delivery:     config.GetEnv("DELIVERY_SERVICE_URL", "http://delivery-service:8085"),
		Notification: config.GetEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8086"),
	}

	app := fiber.New(fiber.Config{
		AppName:      "markettg-gateway",
		ErrorHandler: errorHandler,
	})

	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.GetEnv("CORS_ORIGINS", "*"),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID, Idempotency-Key",
	}))

	rateLimiter := gwmw.NewRateLimiter(redisClient, 100, time.Minute)
	app.Use(rateLimiter.Middleware())

	auth := gwmw.NewAuth(gwmw.AuthConfig{
		BotToken:      config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		BotSecret:     config.GetEnv("BOT_INTERNAL_SECRET", "bot-secret"),
		JWTSecret:     config.GetEnv("JWT_SECRET", "dev-jwt-secret"),
		AdminJWTSecret: config.GetEnv("ADMIN_JWT_SECRET", "dev-admin-jwt-secret"),
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "gateway"})
	})

	api := app.Group("/api/v1")

	// Public catalog routes
	api.Get("/catalog/categories", proxy.Forward(services.Catalog))
	api.Get("/catalog/products", proxy.Forward(services.Catalog))
	api.Get("/catalog/products/:id", proxy.Forward(services.Catalog))
	api.Get("/catalog/promotions", proxy.Forward(services.Catalog))

	// Auth
	api.Post("/auth/telegram", proxy.Forward(services.User))
	api.Post("/admin/auth/login", proxy.Forward(services.User))
	api.Post("/admin/auth/refresh", proxy.Forward(services.User))

	// Protected user routes
	userRoutes := api.Group("", auth.TelegramAuth())
	userRoutes.Get("/me", proxy.Forward(services.User))
	userRoutes.Get("/favorites", proxy.Forward(services.User))
	userRoutes.Post("/favorites/:productId", proxy.Forward(services.User))
	userRoutes.Delete("/favorites/:productId", proxy.Forward(services.User))

	// Cart & orders
	userRoutes.Get("/cart", proxy.Forward(services.Order))
	userRoutes.Put("/cart/items", proxy.Forward(services.Order))
	userRoutes.Delete("/cart/items/:productId", proxy.Forward(services.Order))
	userRoutes.Post("/orders", proxy.Forward(services.Order))
	userRoutes.Get("/orders", proxy.Forward(services.Order))
	userRoutes.Get("/orders/:id", proxy.Forward(services.Order))
	userRoutes.Post("/orders/:id/cancel", proxy.Forward(services.Order))

	// Payments
	userRoutes.Post("/payments", proxy.Forward(services.Payment))
	api.Post("/payments/webhooks/telegram", proxy.Forward(services.Payment))
	api.Post("/payments/webhooks/sbp", proxy.Forward(services.Payment))

	// WebSocket
	userRoutes.Get("/ws/orders/:id", proxy.Forward(services.Notification))

	// Admin routes
	adminRoutes := api.Group("/admin", auth.AdminAuth())
	adminRoutes.All("/*", proxy.AdminForward(services))

	// Internal bot routes
	botRoutes := api.Group("/internal/bot", auth.BotAuth())
	botRoutes.Get("/orders", proxy.Forward(services.Order))

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting gateway", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down gateway")
	_ = app.Shutdown()
}

func errorHandler(c *fiber.Ctx, err error) error {
	return httputil.Error(c, err)
}
