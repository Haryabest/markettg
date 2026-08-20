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
			if err := h.s3.PutBytes(ctx, key, data, "image/webp"); err == nil {
				return key, nil
			}
		}
	}

	fileIDs := make([]string, 0, 2)
	if thumbFileID != nil && strings.TrimSpace(*thumbFileID) != "" {
		fileIDs = append(fileIDs, strings.TrimSpace(*thumbFileID))
	}
	if stickerFileID != nil && strings.TrimSpace(*stickerFileID) != "" {
		fileIDs = append(fileIDs, strings.TrimSpace(*stickerFileID))
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
