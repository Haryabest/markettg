package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/delivery-service/internal/repository"
	"github.com/markettg/markettg/services/delivery-service/internal/service"
)

type Handler struct {
	svc            *service.Service
	internalSecret string
}

func New(svc *service.Service, internalSecret string) *Handler {
	return &Handler{svc: svc, internalSecret: internalSecret}
}

func (h *Handler) RequireInternal(c *fiber.Ctx) error {
	if h.internalSecret == "" || c.Get("X-Internal-Secret") != h.internalSecret {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

func (h *Handler) ListDeliveries(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	jobs, err := h.svc.ListDeliveries(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"deliveries": jobs})
}

func (h *Handler) RetryDelivery(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.RetryDelivery(c.Context(), id); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) ConfirmDelivery(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	job, err := h.svc.ConfirmDelivery(c.Context(), id)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, job)
}

func (h *Handler) InternalConfirmOrderDeliveries(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("orderId"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.ConfirmOrderDeliveries(c.Context(), orderID); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) CreateJob(c *fiber.Ctx) error {
	var job repository.DeliveryJob
	if err := c.BodyParser(&job); err != nil {
		return fiber.ErrBadRequest
	}
	created, err := h.svc.CreateJob(c.Context(), job)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, created)
}
