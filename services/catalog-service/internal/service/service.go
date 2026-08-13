package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/services/catalog-service/internal/repository"
	"github.com/markettg/markettg/services/catalog-service/internal/storage"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo      *repository.Repository
	redis     *redis.Client
	s3        *storage.S3
	publicURL string
}

func New(repo *repository.Repository, redis *redis.Client, s3 *storage.S3, publicURL string) *Service {
	return &Service{repo: repo, redis: redis, s3: s3, publicURL: publicURL}
}

func (s *Service) withImageURL(p *repository.Product) {
	if p.ImageKey != nil && *p.ImageKey != "" {
		url := fmt.Sprintf("%s/%s", s.publicURL, *p.ImageKey)
		p.ImageURL = &url
	}
}

func (s *Service) ListCategories(ctx context.Context) ([]repository.Category, error) {
	return s.repo.ListCategories(ctx, true)
}

type ProductListResponse struct {
	Products []repository.Product `json:"products"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

func (s *Service) ListProducts(ctx context.Context, f repository.ProductFilter) (*ProductListResponse, error) {
	products, total, err := s.repo.ListProducts(ctx, f)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	for i := range products {
		s.withImageURL(&products[i])
	}
	return &ProductListResponse{Products: products, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (*repository.Product, error) {
	p, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	if p == nil {
		return nil, apperrors.ErrProductNotFound
	}
	s.withImageURL(p)
	return p, nil
}

func (s *Service) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) ([]repository.Product, error) {
	products, err := s.repo.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range products {
		s.withImageURL(&products[i])
	}
	return products, nil
}

func (s *Service) ListPromotions(ctx context.Context) ([]repository.Promotion, error) {
	return s.repo.ListActivePromotions(ctx)
}

func (s *Service) PresignUpload(ctx context.Context) (*storage.PresignResult, error) {
	if s.s3 == nil {
		return nil, apperrors.ErrInternal
	}
	return s.s3.PresignUpload(ctx, "image/jpeg")
}

func (s *Service) CreateProduct(ctx context.Context, p repository.Product) (*repository.Product, error) {
	return s.repo.CreateProduct(ctx, p)
}

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, p repository.Product) error {
	return s.repo.UpdateProduct(ctx, id, p)
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *Service) CreateCategory(ctx context.Context, c repository.Category) (*repository.Category, error) {
	return s.repo.CreateCategory(ctx, c)
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, c repository.Category) error {
	return s.repo.UpdateCategory(ctx, id, c)
}

func (s *Service) CreatePromotion(ctx context.Context, p repository.Promotion) (*repository.Promotion, error) {
	return s.repo.CreatePromotion(ctx, p)
}

// Internal API for order service
type ProductSnapshot struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	PriceKopecks   int64           `json:"price_kopecks"`
	ProductType    string          `json:"product_type"`
	DeliveryConfig json.RawMessage `json:"delivery_config"`
}

func (s *Service) ValidateProducts(ctx context.Context, items map[uuid.UUID]int) ([]ProductSnapshot, int64, error) {
	ids := make([]uuid.UUID, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	products, err := s.repo.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	if len(products) != len(ids) {
		return nil, 0, apperrors.ErrProductNotFound
	}

	var total int64
	var snapshots []ProductSnapshot
	productMap := make(map[uuid.UUID]repository.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	for id, qty := range items {
		p, ok := productMap[id]
		if !ok || !p.IsActive {
			return nil, 0, apperrors.ErrProductNotFound
		}
		total += p.PriceKopecks * int64(qty)
		snapshots = append(snapshots, ProductSnapshot{
			ID: p.ID, Name: p.Name, PriceKopecks: p.PriceKopecks,
			ProductType: p.ProductType, DeliveryConfig: p.DeliveryConfig,
		})
	}
	return snapshots, total, nil
}

func (s *Service) AdminListProducts(ctx context.Context, limit, offset int) (*ProductListResponse, error) {
	return s.ListProducts(ctx, repository.ProductFilter{Limit: limit, Offset: offset})
}

func (s *Service) AdminListCategories(ctx context.Context) ([]repository.Category, error) {
	return s.repo.ListCategories(ctx, false)
}

func (s *Service) AdminListPromotions(ctx context.Context) ([]repository.Promotion, error) {
	return s.repo.ListActivePromotions(ctx)
}

func (s *Service) CacheKey(prefix string) string {
	return fmt.Sprintf("cache:catalog:%s", prefix)
}

func (s *Service) CacheGet(ctx context.Context, key string, dest interface{}) bool {
	if s.redis == nil {
		return false
	}
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return json.Unmarshal([]byte(val), dest) == nil
}

func (s *Service) CacheSet(ctx context.Context, key string, val interface{}, ttl time.Duration) {
	if s.redis == nil {
		return
	}
	data, _ := json.Marshal(val)
	s.redis.Set(ctx, key, data, ttl)
}
