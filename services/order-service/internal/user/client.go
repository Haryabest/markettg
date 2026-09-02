package user

import (
	"bytes"
	"context"
	"encoding/json"
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

type User struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   *string   `json:"username"`
	FirstName  *string   `json:"first_name"`
	LastName   *string   `json:"last_name"`
}

type PromoValidation struct {
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	RewardID      string `json:"reward_id"`
}

func (c *Client) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/internal/users/id/%s", c.baseURL, userID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found")
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) ResolveByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/internal/users/telegram/%d", c.baseURL, telegramID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found")
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) ValidateReferralPromo(ctx context.Context, userID uuid.UUID, code string) (*PromoValidation, error) {
	body, _ := json.Marshal(map[string]string{"user_id": userID.String(), "promo_code": code})
	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/v1/internal/referrals/validate-promo", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid promo")
	}
	var promo PromoValidation
	if err := json.NewDecoder(resp.Body).Decode(&promo); err != nil {
		return nil, err
	}
	return &promo, nil
}

func (c *Client) MarkReferralPromoUsed(ctx context.Context, rewardID string) error {
	body, _ := json.Marshal(map[string]string{"reward_id": rewardID})
	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/v1/internal/referrals/mark-promo-used", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mark promo failed")
	}
	return nil
}

func (c *Client) CompleteReferralOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	body, _ := json.Marshal(map[string]string{"user_id": userID.String(), "order_id": orderID.String()})
	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/v1/internal/referrals/order-completed", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("complete referral failed")
	}
	return nil
}
