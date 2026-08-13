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
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/markettg/markettg/services/catalog-service/internal/cache"
	"github.com/markettg/markettg/services/catalog-service/internal/handlers"
	catmw "github.com/markettg/markettg/services/catalog-service/internal/middleware"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("catalog-service")
	base := config.LoadBase("8082")

	pgCfg := config.LoadPostgres()
	pool, err := pgxpool.New(context.Background(), pgCfg.URL)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	redisCfg := config.LoadRedis()
	redisClient := redisutil.NewClient(redisutil.Config{
		Addr: redisCfg.Addr, Password: redisCfg.Password, DB: redisCfg.DB,
	})
	ctx := context.Background()
	if err := redisutil.Ping(ctx, redisClient); err != nil {
		log.Warn("redis not available at startup", zap.Error(err))
	}

	s3Cfg := config.LoadS3()
	s3Repo, err := repository.NewS3Repository(s3Cfg)
	if err != nil {
		log.Fatal("s3 init failed", zap.Error(err))
	}

	cacheTTL := config.GetEnvDuration("CACHE_TTL", 5*time.Minute)
	catalogCache := cache.New(redisClient, cacheTTL)

	catalogRepo := repository.NewCatalogRepository(pool)
	catalogHandler := handlers.NewCatalogHandler(catalogRepo, s3Repo, catalogCache)
	adminHandler := handlers.NewAdminHandler(catalogRepo, s3Repo, catalogCache)

	app := fiber.New(fiber.Config{
		AppName:      "catalog-service",
		ErrorHandler: httputil.Error,
	})
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "catalog-service"})
	})

	api := app.Group("/api/v1")

	// Public catalog routes
	catalog := api.Group("/catalog")
	catalog.Get("/categories", catalogHandler.ListCategories)
	catalog.Get("/products", catalogHandler.ListProducts)
	catalog.Get("/products/:id", catalogHandler.GetProduct)
	catalog.Get("/promotions", catalogHandler.ListPromotions)

	// Internal
	api.Post("/internal/catalog/validate", catalogHandler.InternalValidateProducts)

	// Admin routes — gateway forwards X-Admin-Role after JWT validation
	admin := api.Group("/admin", catmw.RequireAdmin(catmw.RoleViewer))

	// Products
	admin.Get("/products", adminHandler.ListProducts)
	admin.Post("/products", catmw.RequireAdmin(catmw.RoleManager), adminHandler.CreateProduct)
	admin.Put("/products/:id", catmw.RequireAdmin(catmw.RoleManager), adminHandler.UpdateProduct)
	admin.Delete("/products/:id", catmw.RequireAdmin(catmw.RoleSuperAdmin), adminHandler.DeleteProduct)

	// Categories
	admin.Get("/categories", adminHandler.ListCategories)
	admin.Post("/categories", catmw.RequireAdmin(catmw.RoleManager), adminHandler.CreateCategory)
	admin.Put("/categories/:id", catmw.RequireAdmin(catmw.RoleManager), adminHandler.UpdateCategory)
	admin.Delete("/categories/:id", catmw.RequireAdmin(catmw.RoleSuperAdmin), adminHandler.DeleteCategory)

	// Promotions
	admin.Get("/promotions", adminHandler.ListPromotions)
	admin.Post("/promotions", catmw.RequireAdmin(catmw.RoleManager), adminHandler.CreatePromotion)
	admin.Put("/promotions/:id", catmw.RequireAdmin(catmw.RoleManager), adminHandler.UpdatePromotion)
	admin.Delete("/promotions/:id", catmw.RequireAdmin(catmw.RoleSuperAdmin), adminHandler.DeletePromotion)

	// Uploads
	admin.Post("/uploads/presign", catmw.RequireAdmin(catmw.RoleManager), adminHandler.PresignUpload)

	go func() {
		addr := fmt.Sprintf(":%s", base.Port)
		log.Info("starting catalog-service", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down catalog-service")
	_ = app.Shutdown()
}
