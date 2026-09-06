package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/markettg/markettg/packages/go-shared/pkg/config"
)

type DeliveryConfig struct {
	Type         string `json:"type"`
	Amount       int    `json:"amount"`
	DurationDays int    `json:"duration_days"`
	GiftID       string `json:"gift_id"`
	StarCount    int    `json:"star_count"`
	NFTSlug      string `json:"nft_slug"`
}

type DeliveryHandler interface {
	Type() string
	Deliver(ctx context.Context, telegramID int64, cfg DeliveryConfig) (json.RawMessage, error)
}

// callBotAPI performs a Bot API call and unwraps Telegram's {ok, result, description}
// envelope so failures surface as readable errors in the delivery job log.
func callBotAPI(ctx context.Context, client *http.Client, botToken, method string, payload map[string]interface{}) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		ErrorCode   int             `json:"error_code"`
		Description string          `json:"description"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("%s: unexpected response %s", method, string(respBody))
	}
	if !envelope.OK {
		return nil, fmt.Errorf("%s failed (%d): %s", method, envelope.ErrorCode, envelope.Description)
	}
	return respBody, nil
}

// StarsDeliveryHandler covers "buy N Stars" products. The Bot API has no
// star-transfer method, so the job is parked for manual fulfilment by an admin.
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
	if h.botToken != "" && telegramID != 0 {
		_, _ = callBotAPI(ctx, h.client, h.botToken, "sendMessage", map[string]interface{}{
			"chat_id": telegramID,
			"text":    fmt.Sprintf("Заказ на %d Stars оплачен. Начисление выполнит оператор — мы напишем, как только всё будет готово.", cfg.Amount),
		})
	}
	return nil, fmt.Errorf("stars top-up requires manual fulfilment: Bot API cannot transfer Stars")
}

// PremiumDeliveryHandler uses Fragment API
type PremiumDeliveryHandler struct {
	apiURL   string
	apiToken string
	client   *http.Client
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
		return nil, fmt.Errorf("premium requires manual fulfilment: FRAGMENT_API_TOKEN not set")
	}
	payload := map[string]interface{}{
		"recipient_id": telegramID,
		"months":       cfg.DurationDays / 30,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL+"/premium/gift", bytes.NewReader(body))
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

// GiftDeliveryHandler sends official Telegram Star Gifts via Bot API sendGift,
// paying from the bot's own Stars balance.
type GiftDeliveryHandler struct {
	botToken      string
	payForUpgrade bool
	notify        bool
	client        *http.Client
}

func NewGiftDeliveryHandler() *GiftDeliveryHandler {
	return &GiftDeliveryHandler{
		botToken:      config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		payForUpgrade: config.GetEnv("GIFT_PAY_FOR_UPGRADE", "false") == "true",
		notify:        config.GetEnv("GIFT_NOTIFY_USER", "true") != "false",
		client:        &http.Client{Timeout: 60 * time.Second},
	}
}

func (h *GiftDeliveryHandler) Type() string { return "GIFT" }

func (h *GiftDeliveryHandler) Deliver(ctx context.Context, telegramID int64, cfg DeliveryConfig) (json.RawMessage, error) {
	if cfg.Type == "NFT" || cfg.NFTSlug != "" {
		return nil, fmt.Errorf("NFT gift %q requires manual transfer: Bot API cannot send collectible gifts", cfg.NFTSlug)
	}
	if cfg.GiftID == "" {
		return nil, fmt.Errorf("gift_id is required")
	}
	if telegramID == 0 {
		return nil, fmt.Errorf("recipient telegram id is unknown")
	}
	if h.botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}

	payload := map[string]interface{}{
		"user_id": telegramID,
		"gift_id": cfg.GiftID,
	}
	if h.payForUpgrade {
		payload["pay_for_upgrade"] = true
	}

	resp, err := callBotAPI(ctx, h.client, h.botToken, "sendGift", payload)
	if err != nil {
		return nil, h.explain(ctx, err)
	}

	if h.notify {
		_, _ = callBotAPI(ctx, h.client, h.botToken, "sendMessage", map[string]interface{}{
			"chat_id": telegramID,
			"text":    "🎁 Подарок отправлен! Проверьте профиль Telegram.",
		})
	}
	return resp, nil
}

// explain enriches a sendGift failure with the bot's current Stars balance,
// which is the usual reason a delivery cannot go through.
func (h *GiftDeliveryHandler) explain(ctx context.Context, cause error) error {
	balance, balErr := h.StarBalance(ctx)
	if balErr != nil {
		return cause
	}
	return fmt.Errorf("%w (bot Stars balance: %d)", cause, balance)
}

func (h *GiftDeliveryHandler) StarBalance(ctx context.Context) (int64, error) {
	resp, err := callBotAPI(ctx, h.client, h.botToken, "getMyStarBalance", map[string]interface{}{})
	if err != nil {
		return 0, err
	}
	var parsed struct {
		Result struct {
			Amount json.Number `json:"amount"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return 0, err
	}
	return strconv.ParseInt(parsed.Result.Amount.String(), 10, 64)
}

func GetHandler(handlerType string, handlers map[string]DeliveryHandler) DeliveryHandler {
	return handlers[handlerType]
}
