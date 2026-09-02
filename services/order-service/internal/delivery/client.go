package delivery

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

func NewClient(baseURL, secret string) *Client {
	return &Client{baseURL: baseURL, secret: secret, http: &http.Client{}}
}

func (c *Client) ConfirmOrderDeliveries(ctx context.Context, orderID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/internal/deliveries/orders/%s/confirm", c.baseURL, orderID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("confirm deliveries failed")
	}
	return nil
}
