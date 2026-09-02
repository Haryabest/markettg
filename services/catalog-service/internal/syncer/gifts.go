package syncer

import (
	"context"
	"time"

	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/markettg/markettg/services/catalog-service/internal/cache"
	"github.com/markettg/markettg/services/catalog-service/internal/handlers"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
	"github.com/markettg/markettg/services/catalog-service/internal/telegram"
	"go.uber.org/zap"
)

type GiftSyncer struct {
	repo     *repository.CatalogRepository
	s3       *repository.S3Repository
	cache    *cache.Cache
	handler  *handlers.GiftsHandler
	tgGifts  *telegram.GiftsClient
	interval time.Duration
	log      *zap.Logger
}

func NewGiftSyncer(
	repo *repository.CatalogRepository,
	s3 *repository.S3Repository,
	c *cache.Cache,
	h *handlers.GiftsHandler,
	log *zap.Logger,
) *GiftSyncer {
	intervalSec := config.GetEnvInt("GIFTS_SYNC_INTERVAL_SEC", 600)
	if intervalSec < 60 {
		intervalSec = 60
	}
	return &GiftSyncer{
		repo:     repo,
		s3:       s3,
		cache:    c,
		handler:  h,
		tgGifts:  telegram.NewGiftsClient(config.GetEnv("TELEGRAM_BOT_TOKEN", "")),
		interval: time.Duration(intervalSec) * time.Second,
		log:      log,
	}
}

func (s *GiftSyncer) Run(ctx context.Context) {
	s.syncOnce(ctx)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncOnce(ctx)
		}
	}
}

func (s *GiftSyncer) syncOnce(ctx context.Context) {
	gifts, err := s.tgGifts.GetAvailableGifts(ctx)
	if err != nil {
		s.log.Warn("telegram gifts sync skipped", zap.Error(err))
		return
	}

	inputs := make([]repository.TelegramGiftSyncInput, 0, len(gifts))
	for _, gift := range gifts {
		var emoji *string
		if gift.Emoji != "" {
			emoji = &gift.Emoji
		}
		var thumb *string
		if gift.ThumbFileID != "" {
			thumb = &gift.ThumbFileID
		}
		var sticker *string
		if gift.StickerFileID != "" {
			sticker = &gift.StickerFileID
		}
		inputs = append(inputs, repository.TelegramGiftSyncInput{
			TelegramGiftID:     gift.ID,
			StarCount:          gift.StarCount,
			Emoji:              emoji,
			StickerThumbFileID: thumb,
			StickerFileID:      sticker,
			TotalCount:         gift.TotalCount,
			RemainingCount:     gift.RemainingCount,
		})
	}

	imageKeys := make(map[string]string)
	for _, gift := range inputs {
		existing, _ := s.repo.FindTelegramGiftImageKey(ctx, gift.TelegramGiftID)
		if existing != nil && *existing != "" {
			imageKeys[gift.TelegramGiftID] = *existing
			continue
		}
		key, err := s.handler.StoreGiftImage(
			ctx,
			gift.TelegramGiftID,
			gift.ImageBase64,
			gift.StickerThumbFileID,
			gift.StickerFileID,
			"gifts/telegram",
		)
		if err == nil {
			imageKeys[gift.TelegramGiftID] = key
		}
	}

	rate := int64(config.GetEnvInt("STAR_KOPECKS_RATE", 180))
	result, err := s.repo.SyncTelegramGifts(ctx, inputs, imageKeys, map[string]string{}, rate)
	if err != nil {
		s.log.Warn("telegram gifts sync failed", zap.Error(err))
		return
	}
	_ = s.cache.InvalidateCatalog(ctx)
	s.log.Info("telegram gifts synced",
		zap.Int("upserted", result.Upserted),
		zap.Int("disabled", result.Disabled),
	)
}
