package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/services/order-service/internal/cart"
	"github.com/markettg/markettg/services/order-service/internal/catalog"
	"github.com/markettg/markettg/services/order-service/internal/delivery"
	"github.com/markettg/markettg/services/order-service/internal/repository"
	"github.com/markettg/markettg/services/order-service/internal/user"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo     *repository.Repository
	cart     *cart.Store
	catalog  *catalog.Client
	redis    *redis.Client
	users    *user.Client
	delivery *delivery.Client
}

func New(repo *repository.Repository, cartStore *cart.Store, catalogClient *catalog.Client, redis *redis.Client, users *user.Client, deliveryClient *delivery.Client) *Service {
	return &Service{repo: repo, cart: cartStore, catalog: catalogClient, redis: redis, users: users, delivery: deliveryClient}
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
	var referralRewardID string

	if req.PromoCode != "" {
		promo, err := s.repo.GetPromoCode(ctx, req.PromoCode)
		if err != nil || promo == nil {
			if s.users != nil {
				refPromo, refErr := s.users.ValidateReferralPromo(ctx, userID, req.PromoCode)
				if refErr != nil || refPromo == nil {
					return nil, apperrors.ErrInvalidPromoCode
				}
				switch refPromo.DiscountType {
				case "PERCENT":
					discount = total * refPromo.DiscountValue / 100
				case "FIXED":
					discount = refPromo.DiscountValue
				}
				referralRewardID = refPromo.RewardID
			} else {
				return nil, apperrors.ErrInvalidPromoCode
			}
		} else {
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
	if referralRewardID != "" && s.users != nil {
		_ = s.users.MarkReferralPromoUsed(ctx, referralRewardID)
	}
	if s.users != nil {
		_ = s.users.CompleteReferralOrder(ctx, userID, created.ID)
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

type AdminOrder struct {
	repository.Order
	TelegramID   int64  `json:"telegram_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
}

func (s *Service) AdminListOrders(ctx context.Context, limit, offset int) ([]AdminOrder, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	orders, err := s.repo.ListAllOrdersWithItems(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]AdminOrder, 0, len(orders))
	for _, order := range orders {
		adminOrder := AdminOrder{Order: order}
		if s.users != nil {
			if u, err := s.users.GetByID(ctx, order.UserID); err == nil && u != nil {
				adminOrder.TelegramID = u.TelegramID
				adminOrder.CustomerName = customerName(u)
			}
		}
		result = append(result, adminOrder)
	}
	return result, nil
}

func (s *Service) AdminGetOrder(ctx context.Context, orderID uuid.UUID) (*AdminOrder, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil || order == nil {
		return nil, apperrors.ErrOrderNotFound
	}
	adminOrder := &AdminOrder{Order: *order}
	if s.users != nil {
		if u, err := s.users.GetByID(ctx, order.UserID); err == nil && u != nil {
			adminOrder.TelegramID = u.TelegramID
			adminOrder.CustomerName = customerName(u)
		}
	}
	return adminOrder, nil
}

func (s *Service) AdminConfirmShipment(ctx context.Context, orderID uuid.UUID) (*repository.Order, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil || order == nil {
		return nil, apperrors.ErrOrderNotFound
	}
	switch order.Status {
	case "PAID", "DELIVERY_PENDING", "DELIVERING":
	default:
		return nil, apperrors.New("INVALID_STATUS", "Заказ нельзя подтвердить в текущем статусе", 400)
	}
	if err := s.repo.UpdateStatus(ctx, orderID, "COMPLETED"); err != nil {
		return nil, apperrors.ErrInternal
	}
	if s.delivery != nil {
		_ = s.delivery.ConfirmOrderDeliveries(ctx, orderID)
	}
	order.Status = "COMPLETED"
	return order, nil
}

func customerName(u *user.User) string {
	if u.FirstName != nil && *u.FirstName != "" {
		name := *u.FirstName
		if u.LastName != nil && *u.LastName != "" {
			name += " " + *u.LastName
		}
		return name
	}
	if u.Username != nil && *u.Username != "" {
		return "@" + *u.Username
	}
	return ""
}

func (s *Service) ListPromoCodes(ctx context.Context) ([]repository.PromoCode, error) {
	return s.repo.ListPromoCodes(ctx)
}

func (s *Service) CreatePromoCode(ctx context.Context, p repository.PromoCode) (*repository.PromoCode, error) {
	return s.repo.CreatePromoCode(ctx, p)
}
