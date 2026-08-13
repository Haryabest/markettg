package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/delivery-service/internal/repository"
	"github.com/markettg/markettg/services/delivery-service/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
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
