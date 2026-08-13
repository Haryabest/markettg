package provider

import (
	"context"

	"github.com/google/uuid"
)

type CreatePaymentRequest struct {
	PaymentID      uuid.UUID
	OrderID        uuid.UUID
	UserID         uuid.UUID
	TelegramID     int64
	AmountKopecks  int64
	Description    string
	IdempotencyKey string
}

type CreatePaymentResult struct {
	ProviderPaymentID string
	PaymentURL        string
	InvoicePayload    string
}

type PaymentProvider interface {
	Method() string
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResult, error)
	VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) (*WebhookEvent, error)
	Refund(ctx context.Context, providerPaymentID string, amountKopecks int64) error
}

type WebhookEvent struct {
	ProviderEventID   string
	ProviderPaymentID string
	PaymentID         uuid.UUID
	Status            string
	RawPayload        []byte
}
