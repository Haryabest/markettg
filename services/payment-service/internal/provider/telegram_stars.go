package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/markettg/markettg/packages/go-shared/pkg/config"
)

type TelegramStarsProvider struct {
	botToken string
	client   *http.Client
}

func NewTelegramStarsProvider() *TelegramStarsProvider {
	return &TelegramStarsProvider{
		botToken: config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		client:   &http.Client{},
	}
}

func (p *TelegramStarsProvider) Method() string { return "STARS" }

func (p *TelegramStarsProvider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResult, error) {
	starsAmount := req.AmountKopecks / 100 // approximate stars conversion
	if starsAmount < 1 {
		starsAmount = 1
	}

	payload := map[string]interface{}{
		"title":       req.Description,
		"description": fmt.Sprintf("Order %s", req.OrderID),
		"payload":     req.PaymentID.String(),
		"currency":    "XTR",
		"prices": []map[string]interface{}{
			{"label": req.Description, "amount": starsAmount},
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/createInvoiceLink", p.botToken)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool   `json:"ok"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || !result.OK {
		return nil, fmt.Errorf("telegram createInvoiceLink failed: %s", string(respBody))
	}

	return &CreatePaymentResult{
		ProviderPaymentID: req.PaymentID.String(),
		PaymentURL:        result.Result,
		InvoicePayload:    req.PaymentID.String(),
	}, nil
}

func (p *TelegramStarsProvider) VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) (*WebhookEvent, error) {
	var update struct {
		PreCheckoutQuery *struct {
			ID             string `json:"id"`
			InvoicePayload string `json:"invoice_payload"`
		} `json:"pre_checkout_query"`
		SuccessfulPayment *struct {
			InvoicePayload    string `json:"invoice_payload"`
			TelegramPaymentID string `json:"telegram_payment_charge_id"`
			TotalAmount       int    `json:"total_amount"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &update); err != nil {
		return nil, err
	}

	if update.SuccessfulPayment != nil {
		return &WebhookEvent{
			ProviderEventID:   update.SuccessfulPayment.TelegramPaymentID,
			ProviderPaymentID: update.SuccessfulPayment.TelegramPaymentID,
			Status:            "PAID",
			RawPayload:        body,
		}, nil
	}
	return nil, fmt.Errorf("unsupported webhook event")
}

func (p *TelegramStarsProvider) Refund(ctx context.Context, providerPaymentID string, amountKopecks int64) error {
	return fmt.Errorf("stars refund not supported via API")
}

// SBPProvider — abstract provider; concrete adapter added when bank is chosen.
type SBPProvider struct {
	webhookSecret string
	mockMode      bool
}

func NewSBPProvider() *SBPProvider {
	return &SBPProvider{
		webhookSecret: config.GetEnv("SBP_WEBHOOK_SECRET", "dev-sbp-secret"),
		mockMode:      config.GetEnv("SBP_MOCK_MODE", "true") == "true",
	}
}

func (p *SBPProvider) Method() string { return "SBP" }

func (p *SBPProvider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResult, error) {
	if p.mockMode {
		return &CreatePaymentResult{
			ProviderPaymentID: "mock-" + req.PaymentID.String(),
			PaymentURL:        fmt.Sprintf("http://localhost:8080/mock-sbp/%s", req.PaymentID),
		}, nil
	}
	return nil, fmt.Errorf("SBP provider not configured: set SBP_PROVIDER_* env vars")
}

func (p *SBPProvider) VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) (*WebhookEvent, error) {
	sig := headers["X-SBP-Signature"]
	if !p.mockMode && sig != p.webhookSecret {
		return nil, fmt.Errorf("invalid signature")
	}
	var event struct {
		EventID   string `json:"event_id"`
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}
	status := "PAID"
	if strings.ToLower(event.Status) != "succeeded" {
		status = "FAILED"
	}
	return &WebhookEvent{
		ProviderEventID:   event.EventID,
		ProviderPaymentID: event.PaymentID,
		Status:            status,
		RawPayload:        body,
	}, nil
}

func (p *SBPProvider) Refund(ctx context.Context, providerPaymentID string, amountKopecks int64) error {
	if p.mockMode {
		return nil
	}
	return fmt.Errorf("SBP refund not configured")
}
