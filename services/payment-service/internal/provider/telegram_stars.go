package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
)

type TelegramStarsProvider struct {
	botToken      string
	starRate      int64
	markupPercent int64
	client        *http.Client
}

func NewTelegramStarsProvider() *TelegramStarsProvider {
	rate, _ := strconv.ParseInt(config.GetEnv("STAR_KOPECKS_RATE", "180"), 10, 64)
	if rate <= 0 {
		rate = 180
	}
	markup, _ := strconv.ParseInt(config.GetEnv("STARS_MARKUP_PERCENT", "0"), 10, 64)
	if markup < 0 {
		markup = 0
	}
	return &TelegramStarsProvider{
		botToken:      config.GetEnv("TELEGRAM_BOT_TOKEN", ""),
		starRate:      rate,
		markupPercent: markup,
		client:        &http.Client{},
	}
}

func (p *TelegramStarsProvider) Method() string { return "STARS" }

// starsForKopecks converts the order total back into Stars using the same rate
// the catalog used when it priced Telegram gifts.
func (p *TelegramStarsProvider) starsForKopecks(amountKopecks int64) int64 {
	stars := (amountKopecks + p.starRate - 1) / p.starRate
	if p.markupPercent > 0 {
		stars = (stars*(100+p.markupPercent) + 99) / 100
	}
	if stars < 1 {
		stars = 1
	}
	return stars
}

func (p *TelegramStarsProvider) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResult, error) {
	starsAmount := p.starsForKopecks(req.AmountKopecks)

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

type successfulPayment struct {
	Currency          string `json:"currency"`
	TotalAmount       int    `json:"total_amount"`
	InvoicePayload    string `json:"invoice_payload"`
	TelegramPaymentID string `json:"telegram_payment_charge_id"`
}

func (p *TelegramStarsProvider) VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) (*WebhookEvent, error) {
	var update struct {
		Message *struct {
			SuccessfulPayment *successfulPayment `json:"successful_payment"`
		} `json:"message"`
		SuccessfulPayment *successfulPayment `json:"successful_payment"`
	}
	if err := json.Unmarshal(body, &update); err != nil {
		return nil, err
	}

	sp := update.SuccessfulPayment
	if sp == nil && update.Message != nil {
		sp = update.Message.SuccessfulPayment
	}
	if sp == nil {
		return nil, fmt.Errorf("unsupported webhook event")
	}

	// invoice_payload carries our payment UUID, set in CreatePayment.
	paymentID, err := uuid.Parse(strings.TrimSpace(sp.InvoicePayload))
	if err != nil {
		return nil, fmt.Errorf("invalid invoice payload: %q", sp.InvoicePayload)
	}

	eventID := sp.TelegramPaymentID
	if eventID == "" {
		eventID = "stars:" + paymentID.String()
	}

	return &WebhookEvent{
		ProviderEventID:   eventID,
		ProviderPaymentID: eventID,
		PaymentID:         paymentID,
		Status:            "PAID",
		RawPayload:        body,
	}, nil
}

// Refund returns the Stars to the buyer via refundStarPayment. Telegram needs
// both the payer telegram id and the charge id, so providerPaymentID is
// expected as "<telegram_id>:<telegram_payment_charge_id>".
func (p *TelegramStarsProvider) Refund(ctx context.Context, providerPaymentID string, amountKopecks int64) error {
	parts := strings.SplitN(providerPaymentID, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("stars refund requires <telegram_id>:<charge_id>, got %q", providerPaymentID)
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram id in %q", providerPaymentID)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":                    userID,
		"telegram_payment_charge_id": parts[1],
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/refundStarPayment", p.botToken)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || !result.OK {
		return fmt.Errorf("refundStarPayment failed: %s", string(respBody))
	}
	return nil
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
