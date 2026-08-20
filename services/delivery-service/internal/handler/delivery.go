package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/markettg/markettg/packages/go-shared/pkg/config"
)

type DeliveryConfig struct {
	Type string `json:"type"`
	Amount int `json:"amount"`
	DurationDays int `json:"duration_days"`
	GiftID string `json:"gift_id"`
}

type DeliveryHandler interface {
	Type() string
	Deliver(ctx context.Context, telegramID int64, cfg DeliveryConfig) (json.RawMessage, error)
}

// StarsDeliveryHandler uses Telegram Bot API
type StarsDeliveryHandler struct {
	botToken string
	client   *http.Client
}

func NewStarsDeliveryHandler() *StarsDeliveryHandler {
	return &StarsDeliveryHandler{
		botToken: config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *StarsDeliveryHandler) Type() string { return "STARS" }

func (h *StarsDeliveryHandler) Deliver(ctx context.Context, telegramID int64, cfg DeliveryConfig) (json.RawMessage, error) {
	// Telegram Bot API: transfer stars to user
	payload := map[string]interface{}{
		"user_id": telegramID,
		"star_count": cfg.Amount,
		"text": fmt.Sprintf("Ваш заказ: %d Stars", cfg.Amount),
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", h.botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("telegram API error: %s", string(respBody))
	}
	return respBody, nil
}

// PremiumDeliveryHandler uses Fragment API
type PremiumDeliveryHandler struct {
	apiURL    string
	apiToken  string
	client    *http.Client
}

func NewPremiumDeliveryHandler() *PremiumDeliveryHandler {
	return &PremiumDeliveryHandler{
		apiURL:   config.GetEnv("FRAGMENT_API_URL", "https://fragment.com/api"),
		apiToken: config.GetEnv("FRAGMENT_API_TOKEN", ""),
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (h *PremiumDeliveryHandler) Type() string { return "PREMIUM" }

func (h *PremiumDeliveryHandler) Deliver(ctx context.Context, telegramID int64, cfg DeliveryConfig) (json.RawMessage, error) {
	if h.apiToken == "" {
		return json.RawMessage(`{"status":"simulated","message":"FRAGMENT_API_TOKEN not set"}`), nil
	}
	payload := map[string]interface{}{
		"recipient_id": telegramID,
		"months":       cfg.DurationDays / 30,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", h.apiURL+"/premium/gift", jsonReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.apiToken)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fragment API error: %s", string(respBody))
	}
	return respBody, nil
}

// GiftDeliveryHandler sends official Telegram Star Gifts via Bot API.
type GiftDeliveryHandler struct {
	botToken string
	client   *http.Client
}

func NewGiftDeliveryHandler() *GiftDeliveryHandler {
	return &GiftDeliveryHandler{
		botToken: config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (h *GiftDeliveryHandler) Type() string { return "GIFT" }

func (h *GiftDeliveryHandler) Deliver(ctx context.Context, telegramID int64, cfg DeliveryConfig) (json.RawMessage, error) {
	giftID := cfg.GiftID
	if giftID == "" {
		return nil, fmt.Errorf("gift_id is required")
	}
	if h.botToken == "" {
		return json.RawMessage(fmt.Sprintf(`{"status":"simulated","gift_id":"%s"}`, giftID)), nil
	}

	payload := map[string]interface{}{
		"user_id": telegramID,
		"gift_id": giftID,
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendGift", h.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, jsonReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("telegram sendGift error: %s", string(respBody))
	}
	return respBody, nil
}

type jsonReaderType struct {
	data []byte
	pos  int
}

func jsonReader(data []byte) *jsonReaderType { return &jsonReaderType{data: data} }
func (r *jsonReaderType) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func GetHandler(handlerType string, handlers map[string]DeliveryHandler) DeliveryHandler {
	return handlers[handlerType]
}
