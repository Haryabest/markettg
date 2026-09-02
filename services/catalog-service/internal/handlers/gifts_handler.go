package handlers

import (
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/services/catalog-service/internal/cache"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
	"github.com/markettg/markettg/services/catalog-service/internal/telegram"
)

type GiftsHandler struct {
	repo     *repository.CatalogRepository
	s3       *repository.S3Repository
	cache    *cache.Cache
	tg       *telegram.Client
	starRate int64
}

func NewGiftsHandler(repo *repository.CatalogRepository, s3 *repository.S3Repository, c *cache.Cache) *GiftsHandler {
	return &GiftsHandler{
		repo:     repo,
		s3:       s3,
		cache:    c,
		tg:       telegram.NewClient(config.GetEnv("TELEGRAM_BOT_TOKEN", "")),
		starRate: int64(config.GetEnvInt("STAR_KOPECKS_RATE", 180)),
	}
}
