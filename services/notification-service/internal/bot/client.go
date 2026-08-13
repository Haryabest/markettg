package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/markettg/markettg/packages/go-shared/pkg/config"
)

type Client struct {
	baseURL   string
	secret    string
	http      *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: config.GetEnv("BOT_SERVICE_URL", "http://bot:8090"),
		secret:  config.GetEnv("BOT_INTERNAL_SECRET", "bot-secret"),
		http:    &http.Client{},
	}
}

type NotifyRequest struct {
	TelegramID int64  `json:"telegram_id"`
	Message    string `json:"message"`
}

func (c *Client) SendMessage(ctx context.Context, telegramID int64, message string) error {
	body, _ := json.Marshal(NotifyRequest{TelegramID: telegramID, Message: message})
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/internal/notify", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bot notify failed: %d", resp.StatusCode)
	}
	return nil
}
