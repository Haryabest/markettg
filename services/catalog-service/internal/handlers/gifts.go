package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
)

func (h *GiftsHandler) SyncTelegramGifts(c *fiber.Ctx) error {
	var req struct {
		Gifts []repository.TelegramGiftSyncInput `json:"gifts"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	ctx := c.Context()
	imageKeys := make(map[string]string)

	for _, gift := range req.Gifts {
		existing, _ := h.repo.FindTelegramGiftImageKey(ctx, gift.TelegramGiftID)
		if existing != nil && *existing != "" {
			imageKeys[gift.TelegramGiftID] = *existing
			continue
		}

		key, err := h.StoreGiftImage(
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

	result, err := h.repo.SyncTelegramGifts(ctx, req.Gifts, imageKeys, h.starRate)
	if err != nil {
		return err
	}
	_ = h.cache.InvalidateCatalog(context.Background())
	return httputil.JSON(c, fiber.StatusOK, result)
}

func (h *GiftsHandler) SyncNFTGifts(c *fiber.Ctx) error {
	var req struct {
		Items []repository.NFTGiftSyncInput `json:"items"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	ctx := c.Context()
	imageKeys := make(map[string]string)

	for _, item := range req.Items {
		telegramID := "nft:" + item.NFTSlug
		existing, _ := h.repo.FindTelegramGiftImageKey(ctx, telegramID)
		if existing != nil && *existing != "" {
			imageKeys[telegramID] = *existing
			continue
		}

		key, err := h.StoreGiftImage(ctx, telegramID, item.ImageBase64, nil, nil, "gifts/nft")
		if err == nil {
			imageKeys[telegramID] = key
		}
	}

	result, err := h.repo.SyncNFTGifts(ctx, req.Items, imageKeys, h.starRate)
	if err != nil {
		return err
	}
	_ = h.cache.InvalidateCatalog(context.Background())
	return httputil.JSON(c, fiber.StatusOK, result)
}
