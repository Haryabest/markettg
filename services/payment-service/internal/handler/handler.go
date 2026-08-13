package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/payment-service/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreatePayment(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Get("X-User-Id"))
	if err != nil {
		return fiber.ErrUnauthorized
	}
	telegramID, _ := strconv.ParseInt(c.Get("X-Telegram-Id"), 10, 64)

	var req service.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	req.TelegramID = telegramID

	idempotencyKey := c.Get("Idempotency-Key")
	resp, err := h.svc.CreatePayment(c.Context(), userID, telegramID, idempotencyKey, req)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, resp)
}

func (h *Handler) TelegramWebhook(c *fiber.Ctx) error {
	headers := map[string]string{}
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})
	if err := h.svc.HandleWebhook(c.Context(), "STARS", headers, c.Body()); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) SBPWebhook(c *fiber.Ctx) error {
	headers := map[string]string{}
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})
	if err := h.svc.HandleWebhook(c.Context(), "SBP", headers, c.Body()); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) AdminListPayments(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	payments, err := h.svc.ListPayments(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"payments": payments})
}

func (h *Handler) GetPayment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	p, err := h.svc.GetPayment(c.Context(), id)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, p)
}
