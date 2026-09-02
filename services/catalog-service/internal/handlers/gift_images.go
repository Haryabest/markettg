package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/markettg/markettg/services/catalog-service/internal/repository"
)

func decodeImageBase64(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty image")
	}
	if idx := strings.Index(raw, ","); idx >= 0 {
		raw = raw[idx+1:]
	}
	return base64.StdEncoding.DecodeString(raw)
}

func contentTypeForExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "image/webp"
	}
}

func (h *GiftsHandler) StoreGiftImage(
	ctx context.Context,
	idKey string,
	imageBase64 *string,
	thumbFileID *string,
	stickerFileID *string,
	prefix string,
) (string, error) {
	if imageBase64 != nil && strings.TrimSpace(*imageBase64) != "" {
		data, err := decodeImageBase64(*imageBase64)
		if err == nil && len(data) > 0 {
			key := fmt.Sprintf("%s/%s.webp", prefix, repository.SlugForTelegramGift(idKey))
			if err := h.s3.PutBytes(ctx, key, data, contentTypeForExt(detectImageExt(data))); err == nil {
				return key, nil
			}
		}
	}

	fileIDs := make([]string, 0, 2)
	if stickerFileID != nil && strings.TrimSpace(*stickerFileID) != "" {
		fileIDs = append(fileIDs, strings.TrimSpace(*stickerFileID))
	}
	if thumbFileID != nil && strings.TrimSpace(*thumbFileID) != "" {
		fileIDs = append(fileIDs, strings.TrimSpace(*thumbFileID))
	}

	for _, fileID := range fileIDs {
		data, ext, err := h.tg.DownloadFile(ctx, fileID)
		if err != nil || len(data) == 0 {
			continue
		}
		if ext == ".tgs" || ext == ".webm" {
			continue
		}
		key := fmt.Sprintf("%s/%s%s", prefix, repository.SlugForTelegramGift(idKey), ext)
		if err := h.s3.PutBytes(ctx, key, data, contentTypeForExt(ext)); err != nil {
			continue
		}
		return key, nil
	}
	return "", fmt.Errorf("no image")
}

func (h *GiftsHandler) StoreGiftSticker(
	ctx context.Context,
	idKey string,
	stickerBase64 *string,
	prefix string,
) (string, error) {
	if stickerBase64 == nil || strings.TrimSpace(*stickerBase64) == "" {
		return "", fmt.Errorf("empty sticker")
	}
	data, err := decodeImageBase64(*stickerBase64)
	if err != nil || len(data) == 0 {
		return "", fmt.Errorf("invalid sticker")
	}
	key := fmt.Sprintf("%s/%s.tgs", prefix, repository.SlugForTelegramGift(idKey))
	if err := h.s3.PutBytes(ctx, key, data, "application/x-tgsticker"); err != nil {
		return "", err
	}
	return key, nil
}

func detectImageExt(data []byte) string {
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 {
		return ".jpg"
	}
	if len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return ".png"
	}
	if len(data) >= 12 && string(data[8:12]) == "WEBP" {
		return ".webp"
	}
	return ".webp"
}
