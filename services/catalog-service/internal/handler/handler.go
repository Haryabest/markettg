package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
	"github.com/markettg/markettg/services/catalog-service/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListCategories(c *fiber.Ctx) error {
	cats, err := h.svc.ListCategories(c.Context())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"categories": cats})
}

func (h *Handler) ListProducts(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	f := repository.ProductFilter{
		Search:      c.Query("q"),
		ProductType: c.Query("type"),
		Sort:        c.Query("sort"),
		Limit:       limit,
		Offset:      offset,
	}
	if catID := c.Query("category_id"); catID != "" {
		id, err := uuid.Parse(catID)
		if err == nil {
			f.CategoryID = &id
		}
	}
	resp, err := h.svc.ListProducts(c.Context(), f)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, resp)
}

func (h *Handler) GetProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	p, err := h.svc.GetProduct(c.Context(), id)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, p)
}

func (h *Handler) ListPromotions(c *fiber.Ctx) error {
	promos, err := h.svc.ListPromotions(c.Context())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"promotions": promos})
}

func (h *Handler) PresignUpload(c *fiber.Ctx) error {
	result, err := h.svc.PresignUpload(c.Context())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, result)
}

func (h *Handler) AdminListProducts(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	resp, err := h.svc.AdminListProducts(c.Context(), limit, offset)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, resp)
}

func (h *Handler) AdminCreateProduct(c *fiber.Ctx) error {
	var p repository.Product
	if err := c.BodyParser(&p); err != nil {
		return fiber.ErrBadRequest
	}
	created, err := h.svc.CreateProduct(c.Context(), p)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, created)
}

func (h *Handler) AdminUpdateProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	var p repository.Product
	if err := c.BodyParser(&p); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.UpdateProduct(c.Context(), id, p); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) AdminDeleteProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.DeleteProduct(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) AdminListCategories(c *fiber.Ctx) error {
	cats, err := h.svc.AdminListCategories(c.Context())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"categories": cats})
}

func (h *Handler) AdminCreateCategory(c *fiber.Ctx) error {
	var cat repository.Category
	if err := c.BodyParser(&cat); err != nil {
		return fiber.ErrBadRequest
	}
	created, err := h.svc.CreateCategory(c.Context(), cat)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, created)
}

func (h *Handler) AdminUpdateCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	var cat repository.Category
	if err := c.BodyParser(&cat); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.svc.UpdateCategory(c.Context(), id, cat); err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"ok": true})
}

func (h *Handler) AdminListPromotions(c *fiber.Ctx) error {
	promos, err := h.svc.AdminListPromotions(c.Context())
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"promotions": promos})
}

func (h *Handler) AdminCreatePromotion(c *fiber.Ctx) error {
	var p repository.Promotion
	if err := c.BodyParser(&p); err != nil {
		return fiber.ErrBadRequest
	}
	if p.StartsAt.IsZero() {
		p.StartsAt = time.Now()
	}
	created, err := h.svc.CreatePromotion(c.Context(), p)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusCreated, created)
}

// Internal endpoint for order service
func (h *Handler) InternalValidateProducts(c *fiber.Ctx) error {
	var req struct {
		Items map[uuid.UUID]int `json:"items"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	snapshots, total, err := h.svc.ValidateProducts(c.Context(), req.Items)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": snapshots, "total_kopecks": total})
}
