package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/services/payment-service/internal/order"
	"github.com/markettg/markettg/services/payment-service/internal/provider"
	"github.com/markettg/markettg/services/payment-service/internal/repository"
)

type Service struct {
	repo      *repository.Repository
	order     *order.Client
	providers map[string]provider.PaymentProvider
}

func New(repo *repository.Repository, orderClient *order.Client, providers ...provider.PaymentProvider) *Service {
	m := make(map[string]provider.PaymentProvider)
	for _, p := range providers {
		m[p.Method()] = p
	}
	return &Service{repo: repo, order: orderClient, providers: m}
}

type CreatePaymentRequest struct {
	OrderID  uuid.UUID `json:"order_id"`
	Method   string    `json:"method"`
	TelegramID int64   `json:"-"`
}

type CreatePaymentResponse struct {
	PaymentID  uuid.UUID `json:"payment_id"`
	PaymentURL string    `json:"payment_url,omitempty"`
	Status     string    `json:"status"`
	Method     string    `json:"method"`
	Amount     int64     `json:"amount_kopecks"`
}

func (s *Service) CreatePayment(ctx context.Context, userID uuid.UUID, telegramID int64, idempotencyKey string, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	if idempotencyKey == "" {
		return nil, apperrors.New("MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header required", 400)
	}

	existing, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if existing != nil {
		meta := map[string]string{}
		_ = json.Unmarshal(existing.Metadata, &meta)
		return &CreatePaymentResponse{
			PaymentID: existing.ID, PaymentURL: meta["payment_url"],
			Status: existing.Status, Method: existing.Method, Amount: existing.AmountKopecks,
		}, nil
	}

	ord, err := s.order.GetOrder(ctx, req.OrderID)
	if err != nil || ord == nil {
		return nil, apperrors.ErrOrderNotFound
	}
	if ord.UserID != userID {
		return nil, apperrors.ErrForbidden
	}
	if ord.Status != "PAYMENT_PENDING" && ord.Status != "CREATED" {
		return nil, apperrors.New("ORDER_NOT_PAYABLE", "Order is not payable", 400)
	}

	prov, ok := s.providers[req.Method]
	if !ok {
		return nil, apperrors.New("INVALID_PAYMENT_METHOD", "Unsupported payment method", 400)
	}

	paymentID := uuid.New()
	result, err := prov.CreatePayment(ctx, provider.CreatePaymentRequest{
		PaymentID: paymentID, OrderID: req.OrderID, UserID: userID,
		TelegramID: telegramID, AmountKopecks: ord.TotalKopecks,
		Description: fmt.Sprintf("Order %s", req.OrderID), IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	meta, _ := json.Marshal(map[string]string{"payment_url": result.PaymentURL})
	providerID := result.ProviderPaymentID
	payment, err := s.repo.Create(ctx, repository.Payment{
		ID: paymentID, OrderID: req.OrderID, UserID: userID, Method: req.Method,
		AmountKopecks: ord.TotalKopecks, Status: "PENDING",
		ProviderPaymentID: &providerID, IdempotencyKey: idempotencyKey, Metadata: meta,
	})
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	_ = s.order.UpdateStatus(ctx, req.OrderID, "PAYMENT_PENDING")

	return &CreatePaymentResponse{
		PaymentID: payment.ID, PaymentURL: result.PaymentURL,
		Status: payment.Status, Method: payment.Method, Amount: payment.AmountKopecks,
	}, nil
}

func (s *Service) markOrderPaid(ctx context.Context, paymentID uuid.UUID, event *provider.WebhookEvent) error {
	orderID, err := s.repo.MarkPaid(ctx, paymentID, event.ProviderPaymentID, event.ProviderEventID, event.RawPayload)
	if err != nil {
		return err
	}
	if orderID != uuid.Nil {
		_ = s.order.UpdateStatus(ctx, orderID, "PAID")
	}
	return nil
}

func (s *Service) HandleWebhook(ctx context.Context, method string, headers map[string]string, body []byte) error {
	prov, ok := s.providers[method]
	if !ok {
		return apperrors.ErrBadRequest
	}
	event, err := prov.VerifyWebhook(ctx, headers, body)
	if err != nil {
		return apperrors.ErrBadRequest
	}
	if event.Status != "PAID" {
		return nil
	}

	// Find payment by provider ID or payload
	paymentID, err := uuid.Parse(event.ProviderPaymentID)
	if err != nil {
		// try invoice payload format mock-sbp-{uuid} or direct uuid
		paymentID, _ = uuid.Parse(event.ProviderPaymentID[5:])
	}

	return s.markOrderPaid(ctx, paymentID, event)
}

func (s *Service) ListPayments(ctx context.Context, limit, offset int) ([]repository.Payment, error) {
	return s.repo.ListAll(ctx, limit, offset)
}

func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil || p == nil {
		return nil, apperrors.ErrPaymentNotFound
	}
	return p, nil
}
