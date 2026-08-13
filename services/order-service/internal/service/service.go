package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/services/order-service/internal/cart"
	"github.com/markettg/markettg/services/order-service/internal/catalog"
	"github.com/markettg/markettg/services/order-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo    *repository.Repository
	cart    *cart.Store
	catalog *catalog.Client
	redis   *redis.Client
}

func New(repo *repository.Repository, cartStore *cart.Store, catalogClient *catalog.Client, redis *redis.Client) *Service {
	return &Service{repo: repo, cart: cartStore, catalog: catalogClient, redis: redis}
}

func (s *Service) GetCart(ctx context.Context, userID string) (*cart.Cart, error) {
	return s.cart.Get(ctx, userID)
}

func (s *Service) UpdateCartItem(ctx context.Context, userID string, productID uuid.UUID, quantity int) (*cart.Cart, error) {
	if quantity > 99 {
		return nil, apperrors.New("INVALID_QUANTITY", "Quantity must be between 1 and 99", 400)
	}
	return s.cart.UpsertItem(ctx, userID, productID, quantity)
}

func (s *Service) RemoveCartItem(ctx context.Context, userID string, productID uuid.UUID) (*cart.Cart, error) {
	return s.cart.RemoveItem(ctx, userID, productID)
}

type CreateOrderRequest struct {
	PromoCode      string `json:"promo_code"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (s *Service) CreateOrder(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*repository.Order, error) {
	cartData, err := s.cart.Get(ctx, userID.String())
	if err != nil || len(cartData.Items) == 0 {
		return nil, apperrors.New("EMPTY_CART", "Cart is empty", 400)
	}

	validation, err := s.catalog.ValidateProducts(ctx, cartData.ToMap())
	if err != nil {
		return nil, apperrors.ErrProductNotFound
	}

	total := validation.TotalKopecks
	var discount int64
	var promoID *uuid.UUID

	if req.PromoCode != "" {
		promo, err := s.repo.GetPromoCode(ctx, req.PromoCode)
		if err != nil || promo == nil {
			return nil, apperrors.ErrInvalidPromoCode
		}
		if promo.ValidTo != nil && promo.ValidTo.Before(time.Now()) {
			return nil, apperrors.ErrInvalidPromoCode
		}
		promoID = &promo.ID
		switch promo.DiscountType {
		case "PERCENT":
			discount = total * promo.DiscountValue / 100
		case "FIXED":
			discount = promo.DiscountValue
		}
		if discount > total {
			discount = total
		}
		total -= discount
	}

	var items []repository.OrderItem
	itemMap := make(map[uuid.UUID]catalog.ProductSnapshot)
	for _, snap := range validation.Items {
		itemMap[snap.ID] = snap
	}
	for _, cartItem := range cartData.Items {
		snap := itemMap[cartItem.ProductID]
		items = append(items, repository.OrderItem{
			ProductID:      cartItem.ProductID,
			Name:           snap.Name,
			PriceKopecks:   snap.PriceKopecks,
			DeliveryConfig: snap.DeliveryConfig,
			ProductType:    snap.ProductType,
			Quantity:       cartItem.Quantity,
		})
	}

	order := repository.Order{
		UserID:          userID,
		Status:          "PAYMENT_PENDING",
		TotalKopecks:    total,
		DiscountKopecks: discount,
		PromoCodeID:     promoID,
	}
	if req.IdempotencyKey != "" {
		order.IdempotencyKey = &req.IdempotencyKey
	}

	created, err := s.repo.CreateOrder(ctx, order, items)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	if promoID != nil {
		_ = s.repo.UsePromoCode(ctx, *promoID, userID, created.ID)
	}

	_ = s.cart.Clear(ctx, userID.String())
	return created, nil
}

func (s *Service) GetOrder(ctx context.Context, userID, orderID uuid.UUID) (*repository.Order, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil || order == nil {
		return nil, apperrors.ErrOrderNotFound
	}
	if order.UserID != userID {
		return nil, apperrors.ErrForbidden
	}
	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.Order, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.ListOrdersByUser(ctx, userID, limit, offset)
}

func (s *Service) CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	order, err := s.GetOrder(ctx, userID, orderID)
	if err != nil {
		return err
	}
	if order.Status != "CREATED" && order.Status != "PAYMENT_PENDING" {
		return apperrors.New("ORDER_NOT_CANCELLABLE", "Order cannot be cancelled", 400)
	}
	return s.repo.UpdateStatus(ctx, orderID, "CANCELLED")
}

func (s *Service) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, orderID, status)
}

func (s *Service) GetOrderInternal(ctx context.Context, orderID uuid.UUID) (*repository.Order, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil || order == nil {
		return nil, apperrors.ErrOrderNotFound
	}
	return order, nil
}

func (s *Service) AdminListOrders(ctx context.Context, limit, offset int) ([]repository.Order, error) {
	return s.repo.ListAllOrders(ctx, limit, offset)
}

func (s *Service) ListPromoCodes(ctx context.Context) ([]repository.PromoCode, error) {
	return s.repo.ListPromoCodes(ctx)
}

func (s *Service) CreatePromoCode(ctx context.Context, p repository.PromoCode) (*repository.PromoCode, error) {
	return s.repo.CreatePromoCode(ctx, p)
}
