package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/services/catalog-service/internal/cache"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
)

type CatalogHandler struct {
	repo  *repository.CatalogRepository
	s3    *repository.S3Repository
	cache *cache.Cache
}

func NewCatalogHandler(repo *repository.CatalogRepository, s3 *repository.S3Repository, c *cache.Cache) *CatalogHandler {
	return &CatalogHandler{repo: repo, s3: s3, cache: c}
}

func (h *CatalogHandler) ListCategories(c *fiber.Ctx) error {
	ctx := c.Context()
	key := cache.CategoriesKey()

	var categories []repository.Category
	if found, err := h.cache.Get(ctx, key, &categories); err == nil && found {
		return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": categories})
	}

	items, err := h.repo.ListCategories(ctx, true)
	if err != nil {
		return err
	}
	_ = h.cache.Set(ctx, key, items)
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": items})
}

func (h *CatalogHandler) ListProducts(c *fiber.Ctx) error {
	ctx := c.Context()
	filter := parseProductFilter(c, true)
	cacheKey := cache.ProductsKey(fmt.Sprintf("%+v", filter))

	var result repository.ProductListResult
	if found, err := h.cache.Get(ctx, cacheKey, &result); err == nil && found {
		repository.ApplyImageURLs(result.Items, h.s3)
		return httputil.JSON(c, fiber.StatusOK, result)
	}

	res, err := h.repo.ListProducts(ctx, filter)
	if err != nil {
		return err
	}
	repository.ApplyImageURLs(res.Items, h.s3)
	_ = h.cache.Set(ctx, cacheKey, res)
	return httputil.JSON(c, fiber.StatusOK, res)
}

func (h *CatalogHandler) GetProduct(c *fiber.Ctx) error {
	ctx := c.Context()
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httputil.JSON(c, fiber.StatusBadRequest, fiber.Map{
			"error": fiber.Map{"code": "BAD_REQUEST", "message": "invalid product id"},
		})
	}

	key := cache.ProductKey(id.String())
	var product repository.Product
	if found, err := h.cache.Get(ctx, key, &product); err == nil && found {
		repository.ApplyImageURL(&product, h.s3)
		return httputil.JSON(c, fiber.StatusOK, product)
	}

	p, err := h.repo.GetProduct(ctx, id, true)
	if err != nil {
		return err
	}
	repository.ApplyImageURL(p, h.s3)
	_ = h.cache.Set(ctx, key, p)
	return httputil.JSON(c, fiber.StatusOK, p)
}

func (h *CatalogHandler) ListPromotions(c *fiber.Ctx) error {
	ctx := c.Context()
	key := cache.PromotionsKey()

	var promos []repository.Promotion
	if found, err := h.cache.Get(ctx, key, &promos); err == nil && found {
		repository.ApplyBannerURLs(promos, h.s3)
		return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": promos})
	}

	items, err := h.repo.ListPromotions(ctx, true)
	if err != nil {
		return err
	}
	repository.ApplyBannerURLs(items, h.s3)
	_ = h.cache.Set(ctx, key, items)
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": items})
}

func parseProductFilter(c *fiber.Ctx, activeOnly bool) repository.ProductFilter {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filter := repository.ProductFilter{
		Query:      c.Query("q"),
		Category:   c.Query("category"),
		Type:       c.Query("type"),
		Sort:       c.Query("sort", "popularity"),
		Page:       page,
		Limit:      limit,
		ActiveOnly: activeOnly,
	}

	if onSale := c.Query("on_sale"); onSale != "" {
		val := onSale == "true" || onSale == "1"
		filter.OnSale = &val
	}

	if filter.Query != "" && filter.Sort == "popularity" {
		filter.Sort = "relevance"
	}

	return filter
}

func (h *CatalogHandler) InternalValidateProducts(c *fiber.Ctx) error {
	var req struct {
		Items map[uuid.UUID]int `json:"items"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}
	snapshots, total, err := h.repo.ValidateProducts(c.Context(), req.Items)
	if err != nil {
		return err
	}
	return httputil.JSON(c, fiber.StatusOK, fiber.Map{"items": snapshots, "total_kopecks": total})
}
