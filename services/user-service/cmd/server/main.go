package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/logger"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
	"github.com/markettg/markettg/services/user-service/internal/handler"
	usermw "github.com/markettg/markettg/services/user-service/internal/middleware"
	"github.com/markettg/markettg/services/user-service/internal/repository"
	"github.com/markettg/markettg/services/user-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("user-service")
	base := config.LoadBase("8081")
	pgCfg := config.LoadPostgres()

	pool, err := pgxpool.New(context.Background(), pgCfg.URL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	repo := repository.New(pool)
	svc := service.New(repo, config.GetEnv("ADMIN_JWT_SECRET", "dev-admin-jwt-secret"))
	h := handler.New(svc)

	app := fiber.New(fiber.Config{AppName: "user-service", ErrorHandler: httputil.Error})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "user-service"})
	})

	api := app.Group("/api/v1")
	api.Post("/auth/telegram", h.AuthTelegram)
	api.Post("/admin/auth/login", h.AdminLogin)
	api.Post("/admin/auth/refresh", h.AdminRefresh)
	api.Get("/me", usermw.RequireTelegramID(), h.GetMe)
	api.Get("/favorites", usermw.RequireTelegramID(), h.ListFavorites)
	api.Post("/favorites/:productId", usermw.RequireTelegramID(), h.AddFavorite)
	api.Delete("/favorites/:productId", usermw.RequireTelegramID(), h.RemoveFavorite)
	api.Get("/referrals/me", usermw.RequireTelegramID(), h.GetReferrals)

	internalSecret := config.GetEnv("BOT_INTERNAL_SECRET", "bot-secret")
	internal := api.Group("/internal", usermw.InternalAuth(internalSecret))
	internal.Get("/users/telegram/:telegramId", h.ResolveByTelegramID)
	internal.Post("/referrals/validate-promo", h.InternalValidateReferralPromo)
	internal.Post("/referrals/mark-promo-used", h.InternalMarkReferralPromoUsed)
	internal.Post("/referrals/order-completed", h.InternalCompleteReferralOrder)

	// Admin
	admin := api.Group("/admin")
	admin.Get("/users", h.AdminListUsers)
	admin.Get("/audit-log", h.AdminAuditLog)

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting user-service", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = app.Shutdown()
}
