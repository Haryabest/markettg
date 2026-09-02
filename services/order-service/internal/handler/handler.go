package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/order-service/internal/repository"
	"github.com/markettg/markettg/services/order-service/internal/service"
	"github.com/markettg/markettg/services/order-service/internal/user"
)

type Handler struct {
	svc    *service.Service
	users  *user.Client
}

func New(svc *service.Service, users *user.Client) *Handler {
	return &Handler{svc: svc, users: users}
}

func (h *Handler) getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	if uid := c.Get("X-User-Id"); uid != "" {
		return uuid.Parse(uid)
	}
	tidStr := c.Get("X-Telegram-Id")
	if tidStr == "" {
		return uuid.Nil, fiber.ErrUnauthorized
	}
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		return uuid.Nil, fiber.ErrUnauthorized
	}
	u, err := h.users.ResolveByTelegramID(c.Context(), tid)
	if err != nil {
		return uuid.Nil, fiber.ErrUnauthorized
	}
	return u.ID, nil
}

func (h *Handler) GetCart(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	cart, err := h.svc.GetCart(c.Context(), userID.String())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, cart)
}

func (h *Handler) UpdateCartItem(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int       `json:"quantity"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	cart, err := h.svc.UpdateCartItem(c.Context(), userID.String(), req.ProductID, req.Quantity)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, cart)
}

func (h *Handler) RemoveCartItem(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	pid, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	cart, err := h.svc.RemoveCartItem(c.Context(), userID.String(), pid)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, cart)
}

func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	var req service.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = c.Get("Idempotency-Key")
	}
	order, err := h.svc.CreateOrder(c.Context(), userID, req)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, order)
}

func (h *Handler) ListOrders(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	orders, err := h.svc.ListOrders(c.Context(), userID, limit, offset)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"orders": orders})
}

func (h *Handler) GetOrder(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	order, err := h.svc.GetOrder(c.Context(), userID, orderID)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, order)
}

func (h *Handler) CancelOrder(c *fiber.Ctx) error {
	userID, err := h.getUserID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.CancelOrder(c.Context(), userID, orderID); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.UpdateOrderStatus(c.Context(), orderID, req.Status); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) BotListOrders(c *fiber.Ctx) error {
	tidStr := c.Get("X-Telegram-User-Id")
	if tidStr == "" {
		return fiber.ErrUnauthorized
	}
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	u, err := h.users.ResolveByTelegramID(c.Context(), tid)
	if err != nil {
		return httputil.JSON(c, fiber.StatusOK, fiber.Map{"orders": []any{}})
	}
	orders, err := h.svc.ListOrders(c.Context(), u.ID, 10, 0)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"orders": orders})
}

func (h *Handler) GetOrderInternal(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	order, err := h.svc.GetOrderInternal(c.Context(), orderID)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, order)
}

func (h *Handler) AdminListOrders(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	orders, err := h.svc.AdminListOrders(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"orders": orders})
}

func (h *Handler) AdminGetOrder(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	order, err := h.svc.AdminGetOrder(c.Context(), orderID)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, order)
}

func (h *Handler) AdminConfirmShipment(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	order, err := h.svc.AdminConfirmShipment(c.Context(), orderID)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, order)
}

func (h *Handler) AdminListPromoCodes(c *fiber.Ctx) error {
	codes, err := h.svc.ListPromoCodes(c.Context())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"promo_codes": codes})
}

func (h *Handler) AdminCreatePromoCode(c *fiber.Ctx) error {
	var p repository.PromoCode
	if err := c.BodyParser(&p); err != nil {
		return fiber.ErrBadRequest
	}
	created, err := h.svc.CreatePromoCode(c.Context(), p)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, created)
}
