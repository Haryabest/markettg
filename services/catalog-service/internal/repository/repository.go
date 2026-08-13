package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
}

type Product struct {
	ID              uuid.UUID       `json:"id"`
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Description     *string         `json:"description"`
	PriceKopecks    int64           `json:"price_kopecks"`
	Currency        string          `json:"currency"`
	ProductType     string          `json:"product_type"`
	DeliveryConfig  json.RawMessage `json:"delivery_config"`
	ImageKey        *string         `json:"image_key"`
	ImageURL        *string         `json:"image_url,omitempty"`
	IsActive        bool            `json:"is_active"`
	PopularityScore int             `json:"popularity_score"`
	CreatedAt       time.Time       `json:"created_at"`
}

type Promotion struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue int64     `json:"discount_value"`
	BannerKey     *string   `json:"banner_key"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	IsActive      bool      `json:"is_active"`
}

type ProductFilter struct {
	Search     string
	CategoryID *uuid.UUID
	ProductType string
	Sort       string
	Limit      int
	Offset     int
	OnSale     bool
	Popular    bool
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListCategories(ctx context.Context, activeOnly bool) ([]Category, error) {
	q := `SELECT id, slug, name, description, sort_order, is_active FROM catalog.categories`
	if activeOnly {
		q += ` WHERE is_active = TRUE`
	}
	q += ` ORDER BY sort_order, name`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.SortOrder, &c.IsActive); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *Repository) ListProducts(ctx context.Context, f ProductFilter) ([]Product, int, error) {
	var conditions []string
	var args []interface{}
	argN := 1

	conditions = append(conditions, "p.is_active = TRUE")

	if f.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			`(p.search_vector @@ plainto_tsquery('russian', $%d) OR similarity(p.name, $%d) > 0.2)`, argN, argN))
		args = append(args, f.Search)
		argN++
	}
	if f.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM catalog.product_categories pc WHERE pc.product_id = p.id AND pc.category_id = $%d)`, argN))
		args = append(args, *f.CategoryID)
		argN++
	}
	if f.ProductType != "" {
		conditions = append(conditions, fmt.Sprintf(`p.product_type = $%d::catalog.product_type`, argN))
		args = append(args, f.ProductType)
		argN++
	}

	where := strings.Join(conditions, " AND ")

	orderBy := "p.popularity_score DESC, p.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderBy = "p.price_kopecks ASC"
	case "price_desc":
		orderBy = "p.price_kopecks DESC"
	case "newest":
		orderBy = "p.created_at DESC"
	case "name":
		orderBy = "p.name ASC"
	}

	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	countQ := `SELECT COUNT(*) FROM catalog.products p WHERE ` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`
		SELECT p.id, p.slug, p.name, p.description, p.price_kopecks, p.currency,
		       p.product_type::text, p.delivery_config, p.image_key, p.is_active,
		       p.popularity_score, p.created_at
		FROM catalog.products p
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, where, orderBy, argN, argN+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
			&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive, &p.PopularityScore, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *Repository) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, price_kopecks, currency, product_type::text,
		       delivery_config, image_key, is_active, popularity_score, created_at
		FROM catalog.products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
		&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive, &p.PopularityScore, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) ([]Product, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, description, price_kopecks, currency, product_type::text,
		       delivery_config, image_key, is_active, popularity_score, created_at
		FROM catalog.products WHERE id = ANY($1) AND is_active = TRUE`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
			&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive, &p.PopularityScore, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *Repository) ListActivePromotions(ctx context.Context) ([]Promotion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, description, discount_type::text, discount_value, banner_key, starts_at, ends_at, is_active
		FROM catalog.promotions
		WHERE is_active = TRUE AND starts_at <= NOW() AND ends_at >= NOW()
		ORDER BY starts_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var promos []Promotion
	for rows.Next() {
		var p Promotion
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.DiscountType, &p.DiscountValue,
			&p.BannerKey, &p.StartsAt, &p.EndsAt, &p.IsActive); err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, nil
}

func (r *Repository) CreateProduct(ctx context.Context, p Product) (*Product, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO catalog.products (slug, name, description, price_kopecks, product_type, delivery_config, image_key, popularity_score)
		VALUES ($1, $2, $3, $4, $5::catalog.product_type, $6, $7, $8)
		RETURNING id, slug, name, description, price_kopecks, currency, product_type::text, delivery_config, image_key, is_active, popularity_score, created_at`,
		p.Slug, p.Name, p.Description, p.PriceKopecks, p.ProductType, p.DeliveryConfig, p.ImageKey, p.PopularityScore,
	).Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
		&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive, &p.PopularityScore, &p.CreatedAt)
	return &p, err
}

func (r *Repository) UpdateProduct(ctx context.Context, id uuid.UUID, p Product) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE catalog.products SET name=$2, description=$3, price_kopecks=$4, delivery_config=$5,
		image_key=$6, is_active=$7, popularity_score=$8, updated_at=NOW() WHERE id=$1`,
		id, p.Name, p.Description, p.PriceKopecks, p.DeliveryConfig, p.ImageKey, p.IsActive, p.PopularityScore)
	return err
}

func (r *Repository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE catalog.products SET is_active=FALSE, updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *Repository) CreateCategory(ctx context.Context, c Category) (*Category, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO catalog.categories (slug, name, description, sort_order)
		VALUES ($1, $2, $3, $4) RETURNING id, slug, name, description, sort_order, is_active`,
		c.Slug, c.Name, c.Description, c.SortOrder,
	).Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.SortOrder, &c.IsActive)
	return &c, err
}

func (r *Repository) UpdateCategory(ctx context.Context, id uuid.UUID, c Category) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE catalog.categories SET name=$2, description=$3, sort_order=$4, is_active=$5, updated_at=NOW() WHERE id=$1`,
		id, c.Name, c.Description, c.SortOrder, c.IsActive)
	return err
}

func (r *Repository) CreatePromotion(ctx context.Context, p Promotion) (*Promotion, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO catalog.promotions (title, description, discount_type, discount_value, banner_key, starts_at, ends_at)
		VALUES ($1, $2, $3::catalog.discount_type, $4, $5, $6, $7)
		RETURNING id, title, description, discount_type::text, discount_value, banner_key, starts_at, ends_at, is_active`,
		p.Title, p.Description, p.DiscountType, p.DiscountValue, p.BannerKey, p.StartsAt, p.EndsAt,
	).Scan(&p.ID, &p.Title, &p.Description, &p.DiscountType, &p.DiscountValue,
		&p.BannerKey, &p.StartsAt, &p.EndsAt, &p.IsActive)
	return &p, err
}
