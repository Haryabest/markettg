package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var slugSanitizer = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

type TelegramGiftSyncInput struct {
	TelegramGiftID     string  `json:"telegram_gift_id"`
	StarCount          int     `json:"star_count"`
	Emoji              *string `json:"emoji"`
	StickerThumbFileID *string `json:"sticker_thumb_file_id"`
	StickerFileID      *string `json:"sticker_file_id"`
	ImageBase64        *string `json:"image_base64"`
	TotalCount         *int    `json:"total_count"`
	RemainingCount     *int    `json:"remaining_count"`
}

type NFTGiftSyncInput struct {
	NFTSlug      string  `json:"nft_slug"`
	Title        string  `json:"title"`
	GiftNum      int     `json:"gift_num"`
	BaseGiftID   int64   `json:"base_gift_id"`
	StarCount    int64   `json:"star_count"`
	ImageBase64  *string `json:"image_base64"`
}

type GiftSyncResult struct {
	Upserted int `json:"upserted"`
	Disabled int `json:"disabled"`
}

func SlugForTelegramGift(giftID string) string {
	safe := slugSanitizer.ReplaceAllString(strings.TrimSpace(giftID), "-")
	safe = strings.Trim(safe, "-")
	if safe == "" {
		safe = "gift"
	}
	return "tg-gift-" + strings.ToLower(safe)
}

func slugForTelegramGift(giftID string) string {
	return SlugForTelegramGift(giftID)
}

func giftName(emoji *string, giftID string) string {
	if emoji != nil && strings.TrimSpace(*emoji) != "" {
		return fmt.Sprintf("Подарок %s", strings.TrimSpace(*emoji))
	}
	shortID := giftID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return fmt.Sprintf("Telegram Gift %s", shortID)
}

func giftDescription(in TelegramGiftSyncInput) string {
	parts := []string{"Официальный Telegram Gift из каталога Telegram."}
	if in.TotalCount != nil && in.RemainingCount != nil {
		parts = append(parts, fmt.Sprintf("Лимитированный: осталось %d из %d.", *in.RemainingCount, *in.TotalCount))
	}
	parts = append(parts, fmt.Sprintf("Стоимость в Telegram: %d Stars.", in.StarCount))
	return strings.Join(parts, " ")
}

func (r *CatalogRepository) GetCategoryIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM catalog.categories WHERE slug = $1 AND is_active = TRUE`, slug).Scan(&id)
	return id, err
}

func (r *CatalogRepository) SyncTelegramGifts(
	ctx context.Context,
	inputs []TelegramGiftSyncInput,
	imageKeys map[string]string,
	starKopecksRate int64,
) (*GiftSyncResult, error) {
	if len(inputs) == 0 {
		return &GiftSyncResult{}, nil
	}
	if starKopecksRate <= 0 {
		starKopecksRate = 180
	}

	giftsCategoryID, err := r.GetCategoryIDBySlug(ctx, "gifts")
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := &GiftSyncResult{}
	activeIDs := make([]string, 0, len(inputs))

	for _, in := range inputs {
		if strings.TrimSpace(in.TelegramGiftID) == "" || in.StarCount <= 0 {
			continue
		}

		active := true
		if in.RemainingCount != nil && *in.RemainingCount <= 0 {
			active = false
		}

		slug := slugForTelegramGift(in.TelegramGiftID)
		name := giftName(in.Emoji, in.TelegramGiftID)
		desc := giftDescription(in)
		price := int64(in.StarCount) * starKopecksRate

		deliveryCfg, _ := json.Marshal(map[string]interface{}{
			"type":              "GIFT",
			"gift_id":           in.TelegramGiftID,
			"telegram_gift_id":  in.TelegramGiftID,
			"star_count":        in.StarCount,
			"total_count":       in.TotalCount,
			"remaining_count":   in.RemainingCount,
			"telegram_synced":   true,
		})

		var imageKey *string
		if key, ok := imageKeys[in.TelegramGiftID]; ok && key != "" {
			imageKey = &key
		}

		popularity := in.StarCount
		if in.RemainingCount != nil && *in.RemainingCount > 0 {
			popularity += *in.RemainingCount
		}

		var productID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO catalog.products (
				slug, name, description, price_kopecks, currency, product_type,
				delivery_config, image_key, is_active, popularity_score, telegram_gift_id
			) VALUES (
				$1, $2, $3, $4, 'RUB', 'GIFT', $5, $6, $7, $8, $9
			)
			ON CONFLICT ON CONSTRAINT products_telegram_gift_id_key DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				price_kopecks = EXCLUDED.price_kopecks,
				delivery_config = EXCLUDED.delivery_config,
				image_key = COALESCE(EXCLUDED.image_key, catalog.products.image_key),
				is_active = EXCLUDED.is_active,
				popularity_score = EXCLUDED.popularity_score,
				updated_at = NOW()
			RETURNING id
		`, slug, name, desc, price, deliveryCfg, imageKey, active, popularity, in.TelegramGiftID).Scan(&productID)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO catalog.product_categories (product_id, category_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, productID, giftsCategoryID)
		if err != nil {
			return nil, err
		}

		activeIDs = append(activeIDs, in.TelegramGiftID)
		result.Upserted++
	}

	_, err = tx.Exec(ctx, `
		UPDATE catalog.products
		SET is_active = FALSE, updated_at = NOW()
		WHERE product_type = 'GIFT'
		  AND telegram_gift_id IS NULL`)
	if err != nil {
		return nil, err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE catalog.products
		SET is_active = FALSE, updated_at = NOW()
		WHERE product_type = 'GIFT'
		  AND telegram_gift_id IS NOT NULL
		  AND NOT (telegram_gift_id = ANY($1))`, activeIDs)
	if err != nil {
		return nil, err
	}
	result.Disabled = int(tag.RowsAffected())

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *CatalogRepository) FindTelegramGiftImageKey(ctx context.Context, telegramGiftID string) (*string, error) {
	var key *string
	err := r.pool.QueryRow(ctx, `
		SELECT image_key FROM catalog.products
		WHERE telegram_gift_id = $1`, telegramGiftID).Scan(&key)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return key, err
}

func (r *CatalogRepository) SyncNFTGifts(
	ctx context.Context,
	inputs []NFTGiftSyncInput,
	imageKeys map[string]string,
	starKopecksRate int64,
) (*GiftSyncResult, error) {
	if len(inputs) == 0 {
		return &GiftSyncResult{}, nil
	}
	if starKopecksRate <= 0 {
		starKopecksRate = 180
	}

	nftCategoryID, err := r.GetCategoryIDBySlug(ctx, "nft")
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := &GiftSyncResult{}
	activeIDs := make([]string, 0, len(inputs))

	for _, in := range inputs {
		slug := strings.TrimSpace(in.NFTSlug)
		if slug == "" || in.StarCount <= 0 {
			continue
		}

		telegramID := "nft:" + slug
		productSlug := "tg-nft-" + slugSanitizer.ReplaceAllString(strings.ToLower(slug), "-")
		productSlug = strings.Trim(productSlug, "-")
		if productSlug == "" {
			productSlug = "tg-nft-item"
		}

		title := strings.TrimSpace(in.Title)
		if title == "" {
			title = slug
		}
		if in.GiftNum > 0 {
			title = fmt.Sprintf("%s #%d", title, in.GiftNum)
		}

		desc := fmt.Sprintf("Коллекционный Telegram Gift «%s». Уникальный номер, перепродажа через Telegram.", slug)
		price := in.StarCount * starKopecksRate

		deliveryCfg, _ := json.Marshal(map[string]interface{}{
			"type":             "NFT",
			"nft_slug":         slug,
			"gift_num":         in.GiftNum,
			"base_gift_id":     in.BaseGiftID,
			"star_count":       in.StarCount,
			"telegram_synced":  true,
		})

		var imageKey *string
		if key, ok := imageKeys[telegramID]; ok && key != "" {
			imageKey = &key
		}

		popularity := int(in.StarCount)
		if in.GiftNum > 0 {
			popularity += 10000 - min(in.GiftNum, 9999)
		}

		var productID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO catalog.products (
				slug, name, description, price_kopecks, currency, product_type,
				delivery_config, image_key, is_active, popularity_score, telegram_gift_id
			) VALUES (
				$1, $2, $3, $4, 'RUB', 'NFT', $5, $6, TRUE, $7, $8
			)
			ON CONFLICT ON CONSTRAINT products_telegram_gift_id_key DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				price_kopecks = EXCLUDED.price_kopecks,
				delivery_config = EXCLUDED.delivery_config,
				image_key = COALESCE(EXCLUDED.image_key, catalog.products.image_key),
				is_active = TRUE,
				popularity_score = EXCLUDED.popularity_score,
				updated_at = NOW()
			RETURNING id
		`, productSlug, title, desc, price, deliveryCfg, imageKey, popularity, telegramID).Scan(&productID)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO catalog.product_categories (product_id, category_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, productID, nftCategoryID)
		if err != nil {
			return nil, err
		}

		activeIDs = append(activeIDs, telegramID)
		result.Upserted++
	}

	tag, err := tx.Exec(ctx, `
		UPDATE catalog.products
		SET is_active = FALSE, updated_at = NOW()
		WHERE product_type = 'NFT'
		  AND telegram_gift_id IS NOT NULL
		  AND NOT (telegram_gift_id = ANY($1))`, activeIDs)
	if err != nil {
		return nil, err
	}
	result.Disabled = int(tag.RowsAffected())

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
