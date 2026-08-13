package catalog

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
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{}}
}

type ProductSnapshot struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	PriceKopecks   int64           `json:"price_kopecks"`
	ProductType    string          `json:"product_type"`
	DeliveryConfig json.RawMessage `json:"delivery_config"`
}

type ValidateResponse struct {
	Items        []ProductSnapshot `json:"items"`
	TotalKopecks int64             `json:"total_kopecks"`
}

func (c *Client) ValidateProducts(ctx context.Context, items map[uuid.UUID]int) (*ValidateResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{"items": items})
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/internal/catalog/validate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog validation failed: %d", resp.StatusCode)
	}

	var result ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProduct(ctx context.Context, id uuid.UUID) (*ProductSnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/catalog/products/%s", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product not found")
	}
	var p ProductSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
