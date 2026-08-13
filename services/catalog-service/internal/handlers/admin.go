package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/catalog-service/internal/cache"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
)

type AdminHandler struct {
	repo  *repository.CatalogRepository
	s3    *repository.S3Repository
	cache *cache.Cache
}

func NewAdminHandler(repo *repository.CatalogRepository, s3 *repository.S3Repository, c *cache.Cache) *AdminHandler {
	return &AdminHandler{repo: repo, s3: s3, cache: c}
}

func (h *AdminHandler) invalidate(ctx *fiber.Ctx) {
	_ = h.cache.InvalidateCatalog(ctx.Context())
}

// Products

func (h *AdminHandler) ListProducts(c *fiber.Ctx) error {
	filter := parseProductFilter(c, false)
	res, err := h.repo.ListProducts(c.Context(), filter)
	if err != nil {
		return err
	}
	repository.ApplyImageURLs(res.Items, h.s3)
	return httputil.JSON(c, fiber.StatusOK, res)
}

func (h *AdminHandler) CreateProduct(c *fiber.Ctx) error {
	var in repository.ProductInput
	if err := c.BodyParser(&in); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if in.Slug == "" || in.Name == "" || in.ProductType == "" {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	p, err := h.repo.CreateProduct(c.Context(), in)
	if err != nil {
		return err
	}
	repository.ApplyImageURL(p, h.s3)
	h.invalidate(c)
	return httputil.JSON(c, fiber.StatusCreated, p)
}

func (h *AdminHandler) UpdateProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	var in repository.ProductInput
	if err := c.BodyParser(&in); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	p, err := h.repo.UpdateProduct(c.Context(), id, in)
	if err != nil {
		return err
	}
	repository.ApplyImageURL(p, h.s3)
	h.invalidate(c)
	return httputil.JSON(c, fiber.StatusOK, p)
}

func (h *AdminHandler) DeleteProduct(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if err := h.repo.DeleteProduct(c.Context(), id); err != nil {
		return err
	}
	h.invalidate(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// Categories

func (h *AdminHandler) ListCategories(c *fiber.Ctx) error {
	items, err := h.repo.ListCategories(c.Context(), false)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": items})
}

func (h *AdminHandler) CreateCategory(c *fiber.Ctx) error {
	var in repository.CategoryInput
	if err := c.BodyParser(&in); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if in.Slug == "" || in.Name == "" {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	cat, err := h.repo.CreateCategory(c.Context(), in)
	if err != nil {
		return err
	}
	h.invalidate(c)
	return httputil.JSON(c, fiber.StatusCreated, cat)
}

func (h *AdminHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	var in repository.CategoryInput
	if err := c.BodyParser(&in); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	cat, err := h.repo.UpdateCategory(c.Context(), id, in)
	if err != nil {
		return err
	}
	h.invalidate(c)
	return httputil.JSON(c, fiber.StatusOK, cat)
}

func (h *AdminHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if err := h.repo.DeleteCategory(c.Context(), id); err != nil {
		return err
	}
	h.invalidate(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// Promotions

func (h *AdminHandler) ListPromotions(c *fiber.Ctx) error {
	items, err := h.repo.ListPromotions(c.Context(), false)
	if err != nil {
		return err
	}
	repository.ApplyBannerURLs(items, h.s3)
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": items})
}

func (h *AdminHandler) CreatePromotion(c *fiber.Ctx) error {
	var in repository.PromotionInput
	if err := c.BodyParser(&in); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if in.Title == "" || in.DiscountType == "" || in.DiscountValue <= 0 {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if in.EndsAt.Before(in.StartsAt) {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	pr, err := h.repo.CreatePromotion(c.Context(), in)
	if err != nil {
		return err
	}
	repository.ApplyBannerURLs([]repository.Promotion{*pr}, h.s3)
	h.invalidate(c)
	return httputil.JSON(c, fiber.StatusCreated, pr)
}

func (h *AdminHandler) UpdatePromotion(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	var in repository.PromotionInput
	if err := c.BodyParser(&in); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	pr, err := h.repo.UpdatePromotion(c.Context(), id, in)
	if err != nil {
		return err
	}
	repository.ApplyBannerURLs([]repository.Promotion{*pr}, h.s3)
	h.invalidate(c)
	return httputil.JSON(c, fiber.StatusOK, pr)
}

func (h *AdminHandler) DeletePromotion(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}
	if err := h.repo.DeletePromotion(c.Context(), id); err != nil {
		return err
	}
	h.invalidate(c)
	return c.SendStatus(fiber.StatusNoContent)
}

// Uploads

func (h *AdminHandler) PresignUpload(c *fiber.Ctx) error {
	var req repository.PresignRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.Error(c, apperrors.ErrBadRequest)
	}

	resp, err := h.s3.PresignUpload(c.Context(), req)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, resp)
}
