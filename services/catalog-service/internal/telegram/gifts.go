package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AvailableGift struct {
	ID             string
	StarCount      int
	Emoji          string
	ThumbFileID    string
	StickerFileID  string
	TotalCount     *int
	RemainingCount *int
}

type GiftsClient struct {
	botToken   string
	httpClient *http.Client
}

func NewGiftsClient(botToken string) *GiftsClient {
	return &GiftsClient{
		botToken:   botToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *GiftsClient) GetAvailableGifts(ctx context.Context) ([]AvailableGift, error) {
	if c.botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is empty")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getAvailableGifts", c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		OK     bool `json:"ok"`
		Result struct {
			Gifts []struct {
				ID        string `json:"id"`
				StarCount int    `json:"star_count"`
				Sticker   struct {
					FileID    string `json:"file_id"`
					Emoji     string `json:"emoji"`
					Thumbnail *struct {
						FileID string `json:"file_id"`
					} `json:"thumbnail"`
					Thumb *struct {
						FileID string `json:"file_id"`
					} `json:"thumb"`
				} `json:"sticker"`
				TotalCount     *int `json:"total_count"`
				RemainingCount *int `json:"remaining_count"`
			} `json:"gifts"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if !parsed.OK {
		return nil, fmt.Errorf("telegram getAvailableGifts failed: %s", strings.TrimSpace(parsed.Description))
	}

	out := make([]AvailableGift, 0, len(parsed.Result.Gifts))
	for _, gift := range parsed.Result.Gifts {
		item := AvailableGift{
			ID:            gift.ID,
			StarCount:     gift.StarCount,
			Emoji:         gift.Sticker.Emoji,
			StickerFileID: gift.Sticker.FileID,
			TotalCount:    gift.TotalCount,
			RemainingCount: gift.RemainingCount,
		}
		if gift.Sticker.Thumbnail != nil {
			item.ThumbFileID = gift.Sticker.Thumbnail.FileID
		} else if gift.Sticker.Thumb != nil {
			item.ThumbFileID = gift.Sticker.Thumb.FileID
		}
		out = append(out, item)
	}
	return out, nil
}
