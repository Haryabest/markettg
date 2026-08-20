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
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
)

type Category struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type Product struct {
	ID               uuid.UUID       `json:"id"`
	Slug             string          `json:"slug"`
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	PriceKopecks     int64           `json:"price_kopecks"`
	Currency         string          `json:"currency"`
	ProductType      string          `json:"product_type"`
	DeliveryConfig   json.RawMessage `json:"delivery_config"`
	ImageKey         *string         `json:"image_key,omitempty"`
	ImageURL         *string         `json:"image_url,omitempty"`
	IsActive         bool            `json:"is_active"`
	PopularityScore  int             `json:"popularity_score"`
	OnSale           bool            `json:"on_sale,omitempty"`
	SalePriceKopecks *int64          `json:"sale_price_kopecks,omitempty"`
	DiscountType     *string         `json:"discount_type,omitempty"`
	DiscountValue    *int64          `json:"discount_value,omitempty"`
	Categories       []Category      `json:"categories,omitempty"`
	CreatedAt        time.Time       `json:"created_at,omitempty"`
	UpdatedAt        time.Time       `json:"updated_at,omitempty"`
}

type Promotion struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue int64     `json:"discount_value"`
	BannerKey     *string   `json:"banner_key,omitempty"`
	BannerURL     *string   `json:"banner_url,omitempty"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	IsActive      bool      `json:"is_active"`
	ProductIDs    []uuid.UUID `json:"product_ids,omitempty"`
	Products      []Product `json:"products,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type ProductFilter struct {
	Query      string
	Category   string
	Type       string
	Sort       string
	Page       int
	Limit      int
	OnSale     *bool
	ActiveOnly bool
}

type ProductListResult struct {
	Items      []Product `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type CatalogRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) ListCategories(ctx context.Context, activeOnly bool) ([]Category, error) {
	query := `
		SELECT id, slug, name, description, sort_order, is_active, created_at, updated_at
		FROM catalog.categories
	`
	if activeOnly {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY sort_order ASC, name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *CatalogRepository) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	var c Category
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, sort_order, is_active, created_at, updated_at
		FROM catalog.categories WHERE id = $1
	`, id).Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type CategoryInput struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sort_order"`
	IsActive    bool    `json:"is_active"`
}

func (r *CatalogRepository) CreateCategory(ctx context.Context, in CategoryInput) (*Category, error) {
	var c Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO catalog.categories (slug, name, description, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, slug, name, description, sort_order, is_active, created_at, updated_at
	`, in.Slug, in.Name, in.Description, in.SortOrder, in.IsActive).
		Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CatalogRepository) UpdateCategory(ctx context.Context, id uuid.UUID, in CategoryInput) (*Category, error) {
	var c Category
	err := r.pool.QueryRow(ctx, `
		UPDATE catalog.categories
		SET slug = $2, name = $3, description = $4, sort_order = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING id, slug, name, description, sort_order, is_active, created_at, updated_at
	`, id, in.Slug, in.Name, in.Description, in.SortOrder, in.IsActive).
		Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CatalogRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM catalog.categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) ListProducts(ctx context.Context, f ProductFilter) (*ProductListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	args := []interface{}{}
	argN := 1
	conditions := []string{"1=1"}

	if f.ActiveOnly {
		conditions = append(conditions, "p.is_active = TRUE")
	}

	if f.Category != "" {
		conditions = append(conditions, fmt.Sprintf(
			`EXISTS (
				SELECT 1 FROM catalog.product_categories pc
				JOIN catalog.categories c ON c.id = pc.category_id
				WHERE pc.product_id = p.id AND (c.slug = $%d OR c.id::text = $%d)
			)`, argN, argN,
		))
		args = append(args, f.Category)
		argN++
	}

	if f.Type != "" {
		conditions = append(conditions, fmt.Sprintf("p.product_type::text = $%d", argN))
		args = append(args, strings.ToUpper(f.Type))
		argN++
	}

	if f.OnSale != nil {
		if *f.OnSale {
			conditions = append(conditions, "ap.product_id IS NOT NULL")
		} else {
			conditions = append(conditions, "ap.product_id IS NULL")
		}
	}

	searchRank := "0::float8"
	searchSim := "0::float8"
	if f.Query != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			p.search_vector @@ plainto_tsquery('russian', $%d)
			OR similarity(p.name, $%d) > 0.2
			OR similarity(coalesce(p.description, ''), $%d) > 0.15
		)`, argN, argN, argN))
		searchRank = fmt.Sprintf("ts_rank(p.search_vector, plainto_tsquery('russian', $%d))", argN)
		searchSim = fmt.Sprintf("GREATEST(similarity(p.name, $%d), similarity(coalesce(p.description, ''), $%d))", argN, argN)
		args = append(args, f.Query)
		argN++
	}

	where := strings.Join(conditions, " AND ")

	orderBy := "p.popularity_score DESC, p.name ASC"
	switch f.Sort {
	case "price_asc":
		orderBy = "p.price_kopecks ASC, p.name ASC"
	case "price_desc":
		orderBy = "p.price_kopecks DESC, p.name ASC"
	case "name":
		orderBy = "p.name ASC"
	case "newest":
		orderBy = "p.created_at DESC, p.name ASC"
	case "relevance":
		if f.Query != "" {
			orderBy = fmt.Sprintf("(%s * 0.7 + %s * 0.3) DESC, p.popularity_score DESC", searchRank, searchSim)
		}
	case "popularity":
		orderBy = "p.popularity_score DESC, p.name ASC"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT p.id)
		FROM catalog.products p
		LEFT JOIN LATERAL (
			SELECT pp.product_id, pr.discount_type, pr.discount_value
			FROM catalog.promotion_products pp
			JOIN catalog.promotions pr ON pr.id = pp.promotion_id
			WHERE pp.product_id = p.id
			  AND pr.is_active = TRUE
			  AND NOW() BETWEEN pr.starts_at AND pr.ends_at
			ORDER BY pr.discount_value DESC
			LIMIT 1
		) ap ON TRUE
		WHERE %s
	`, where)

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	limitArg := argN
	offsetArg := argN + 1
	args = append(args, f.Limit, offset)

	listQuery := fmt.Sprintf(`
		SELECT
			p.id, p.slug, p.name, p.description, p.price_kopecks, p.currency,
			p.product_type::text, p.delivery_config, p.image_key, p.is_active,
			p.popularity_score, p.created_at, p.updated_at,
			(ap.product_id IS NOT NULL) AS on_sale,
			ap.discount_type::text,
			ap.discount_value,
			CASE
				WHEN ap.discount_type = 'PERCENT' THEN
					GREATEST(p.price_kopecks - (p.price_kopecks * ap.discount_value / 100), 0)
				WHEN ap.discount_type = 'FIXED' THEN
					GREATEST(p.price_kopecks - ap.discount_value, 0)
				ELSE NULL
			END AS sale_price_kopecks,
			%s AS rank_score,
			%s AS sim_score
		FROM catalog.products p
		LEFT JOIN LATERAL (
			SELECT pp.product_id, pr.discount_type, pr.discount_value
			FROM catalog.promotion_products pp
			JOIN catalog.promotions pr ON pr.id = pp.promotion_id
			WHERE pp.product_id = p.id
			  AND pr.is_active = TRUE
			  AND NOW() BETWEEN pr.starts_at AND pr.ends_at
			ORDER BY pr.discount_value DESC
			LIMIT 1
		) ap ON TRUE
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, searchRank, searchSim, where, orderBy, limitArg, offsetArg)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products, err := scanProducts(rows, true)
	if err != nil {
		return nil, err
	}

	if len(products) > 0 {
		if err := r.attachCategories(ctx, products); err != nil {
			return nil, err
		}
	}

	totalPages := int(total) / f.Limit
	if int(total)%f.Limit != 0 {
		totalPages++
	}

	if products == nil {
		products = []Product{}
	}

	return &ProductListResult{
		Items:      products,
		Total:      total,
		Page:       f.Page,
		Limit:      f.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *CatalogRepository) GetProduct(ctx context.Context, id uuid.UUID, activeOnly bool) (*Product, error) {
	query := `
		SELECT
			p.id, p.slug, p.name, p.description, p.price_kopecks, p.currency,
			p.product_type::text, p.delivery_config, p.image_key, p.is_active,
			p.popularity_score, p.created_at, p.updated_at,
			(ap.product_id IS NOT NULL) AS on_sale,
			ap.discount_type::text,
			ap.discount_value,
			CASE
				WHEN ap.discount_type = 'PERCENT' THEN
					GREATEST(p.price_kopecks - (p.price_kopecks * ap.discount_value / 100), 0)
				WHEN ap.discount_type = 'FIXED' THEN
					GREATEST(p.price_kopecks - ap.discount_value, 0)
				ELSE NULL
			END AS sale_price_kopecks
		FROM catalog.products p
		LEFT JOIN LATERAL (
			SELECT pp.product_id, pr.discount_type, pr.discount_value
			FROM catalog.promotion_products pp
			JOIN catalog.promotions pr ON pr.id = pp.promotion_id
			WHERE pp.product_id = p.id
			  AND pr.is_active = TRUE
			  AND NOW() BETWEEN pr.starts_at AND pr.ends_at
			ORDER BY pr.discount_value DESC
			LIMIT 1
		) ap ON TRUE
		WHERE p.id = $1
	`
	if activeOnly {
		query += ` AND p.is_active = TRUE`
	}

	row := r.pool.QueryRow(ctx, query, id)
	p, err := scanProductRow(row)
	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	products := []Product{*p}
	if err := r.attachCategories(ctx, products); err != nil {
		return nil, err
	}
	return &products[0], nil
}

type ProductInput struct {
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Description     *string         `json:"description"`
	PriceKopecks    int64           `json:"price_kopecks"`
	Currency        string          `json:"currency"`
	ProductType     string          `json:"product_type"`
	DeliveryConfig  json.RawMessage `json:"delivery_config"`
	ImageKey        *string         `json:"image_key"`
	IsActive        bool            `json:"is_active"`
	PopularityScore int             `json:"popularity_score"`
	CategoryIDs     []uuid.UUID     `json:"category_ids"`
}

func (r *CatalogRepository) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	if len(in.DeliveryConfig) == 0 {
		in.DeliveryConfig = json.RawMessage(`{}`)
	}
	if in.Currency == "" {
		in.Currency = "RUB"
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var p Product
	err = tx.QueryRow(ctx, `
		INSERT INTO catalog.products (
			slug, name, description, price_kopecks, currency, product_type,
			delivery_config, image_key, is_active, popularity_score
		) VALUES ($1, $2, $3, $4, $5, $6::catalog.product_type, $7, $8, $9, $10)
		RETURNING id, slug, name, description, price_kopecks, currency,
			product_type::text, delivery_config, image_key, is_active,
			popularity_score, created_at, updated_at
	`, in.Slug, in.Name, in.Description, in.PriceKopecks, in.Currency, strings.ToUpper(in.ProductType),
		in.DeliveryConfig, in.ImageKey, in.IsActive, in.PopularityScore).
		Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
			&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive,
			&p.PopularityScore, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := r.setProductCategories(ctx, tx, p.ID, in.CategoryIDs); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	products := []Product{p}
	if err := r.attachCategories(ctx, products); err != nil {
		return nil, err
	}
	return &products[0], nil
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, id uuid.UUID, in ProductInput) (*Product, error) {
	if len(in.DeliveryConfig) == 0 {
		in.DeliveryConfig = json.RawMessage(`{}`)
	}
	if in.Currency == "" {
		in.Currency = "RUB"
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var p Product
	err = tx.QueryRow(ctx, `
		UPDATE catalog.products SET
			slug = $2, name = $3, description = $4, price_kopecks = $5, currency = $6,
			product_type = $7::catalog.product_type, delivery_config = $8, image_key = $9,
			is_active = $10, popularity_score = $11, updated_at = NOW()
		WHERE id = $1
		RETURNING id, slug, name, description, price_kopecks, currency,
			product_type::text, delivery_config, image_key, is_active,
			popularity_score, created_at, updated_at
	`, id, in.Slug, in.Name, in.Description, in.PriceKopecks, in.Currency, strings.ToUpper(in.ProductType),
		in.DeliveryConfig, in.ImageKey, in.IsActive, in.PopularityScore).
		Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
			&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive,
			&p.PopularityScore, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	if in.CategoryIDs != nil {
		if err := r.setProductCategories(ctx, tx, p.ID, in.CategoryIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	products := []Product{p}
	if err := r.attachCategories(ctx, products); err != nil {
		return nil, err
	}
	return &products[0], nil
}

func (r *CatalogRepository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM catalog.products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrProductNotFound
	}
	return nil
}

func (r *CatalogRepository) ListPromotions(ctx context.Context, activeOnly bool) ([]Promotion, error) {
	query := `
		SELECT id, title, description, discount_type::text, discount_value,
			banner_key, starts_at, ends_at, is_active, created_at, updated_at
		FROM catalog.promotions
	`
	if activeOnly {
		query += ` WHERE is_active = TRUE AND NOW() BETWEEN starts_at AND ends_at`
	}
	query += ` ORDER BY starts_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []Promotion
	for rows.Next() {
		var pr Promotion
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.Description, &pr.DiscountType, &pr.DiscountValue,
			&pr.BannerKey, &pr.StartsAt, &pr.EndsAt, &pr.IsActive, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, err
		}
		promos = append(promos, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range promos {
		ids, err := r.getPromotionProductIDs(ctx, promos[i].ID)
		if err != nil {
			return nil, err
		}
		promos[i].ProductIDs = ids
	}
	return promos, nil
}

func (r *CatalogRepository) GetPromotion(ctx context.Context, id uuid.UUID) (*Promotion, error) {
	var pr Promotion
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, description, discount_type::text, discount_value,
			banner_key, starts_at, ends_at, is_active, created_at, updated_at
		FROM catalog.promotions WHERE id = $1
	`, id).Scan(&pr.ID, &pr.Title, &pr.Description, &pr.DiscountType, &pr.DiscountValue,
		&pr.BannerKey, &pr.StartsAt, &pr.EndsAt, &pr.IsActive, &pr.CreatedAt, &pr.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	pr.ProductIDs, err = r.getPromotionProductIDs(ctx, pr.ID)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

type PromotionInput struct {
	Title         string      `json:"title"`
	Description   *string     `json:"description"`
	DiscountType  string      `json:"discount_type"`
	DiscountValue int64       `json:"discount_value"`
	BannerKey     *string     `json:"banner_key"`
	StartsAt      time.Time   `json:"starts_at"`
	EndsAt        time.Time   `json:"ends_at"`
	IsActive      bool        `json:"is_active"`
	ProductIDs    []uuid.UUID `json:"product_ids"`
}

func (r *CatalogRepository) CreatePromotion(ctx context.Context, in PromotionInput) (*Promotion, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var pr Promotion
	err = tx.QueryRow(ctx, `
		INSERT INTO catalog.promotions (
			title, description, discount_type, discount_value, banner_key,
			starts_at, ends_at, is_active
		) VALUES ($1, $2, $3::catalog.discount_type, $4, $5, $6, $7, $8)
		RETURNING id, title, description, discount_type::text, discount_value,
			banner_key, starts_at, ends_at, is_active, created_at, updated_at
	`, in.Title, in.Description, strings.ToUpper(in.DiscountType), in.DiscountValue,
		in.BannerKey, in.StartsAt, in.EndsAt, in.IsActive).
		Scan(&pr.ID, &pr.Title, &pr.Description, &pr.DiscountType, &pr.DiscountValue,
			&pr.BannerKey, &pr.StartsAt, &pr.EndsAt, &pr.IsActive, &pr.CreatedAt, &pr.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := r.setPromotionProducts(ctx, tx, pr.ID, in.ProductIDs); err != nil {
		return nil, err
	}
	pr.ProductIDs = in.ProductIDs

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *CatalogRepository) UpdatePromotion(ctx context.Context, id uuid.UUID, in PromotionInput) (*Promotion, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var pr Promotion
	err = tx.QueryRow(ctx, `
		UPDATE catalog.promotions SET
			title = $2, description = $3, discount_type = $4::catalog.discount_type,
			discount_value = $5, banner_key = $6, starts_at = $7, ends_at = $8,
			is_active = $9, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, description, discount_type::text, discount_value,
			banner_key, starts_at, ends_at, is_active, created_at, updated_at
	`, id, in.Title, in.Description, strings.ToUpper(in.DiscountType), in.DiscountValue,
		in.BannerKey, in.StartsAt, in.EndsAt, in.IsActive).
		Scan(&pr.ID, &pr.Title, &pr.Description, &pr.DiscountType, &pr.DiscountValue,
			&pr.BannerKey, &pr.StartsAt, &pr.EndsAt, &pr.IsActive, &pr.CreatedAt, &pr.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if in.ProductIDs != nil {
		if err := r.setPromotionProducts(ctx, tx, pr.ID, in.ProductIDs); err != nil {
			return nil, err
		}
		pr.ProductIDs = in.ProductIDs
	} else {
		pr.ProductIDs, err = r.getPromotionProductIDs(ctx, pr.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *CatalogRepository) DeletePromotion(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM catalog.promotions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) setProductCategories(ctx context.Context, tx pgx.Tx, productID uuid.UUID, categoryIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM catalog.product_categories WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for _, catID := range categoryIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO catalog.product_categories (product_id, category_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, productID, catID); err != nil {
			return err
		}
	}
	return nil
}

func (r *CatalogRepository) setPromotionProducts(ctx context.Context, tx pgx.Tx, promotionID uuid.UUID, productIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM catalog.promotion_products WHERE promotion_id = $1`, promotionID); err != nil {
		return err
	}
	for _, pid := range productIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO catalog.promotion_products (promotion_id, product_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, promotionID, pid); err != nil {
			return err
		}
	}
	return nil
}

func (r *CatalogRepository) getPromotionProductIDs(ctx context.Context, promotionID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT product_id FROM catalog.promotion_products WHERE promotion_id = $1
	`, promotionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *CatalogRepository) attachCategories(ctx context.Context, products []Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(products))
	index := make(map[uuid.UUID]int, len(products))
	for i, p := range products {
		ids[i] = p.ID
		index[p.ID] = i
	}

	rows, err := r.pool.Query(ctx, `
		SELECT pc.product_id, c.id, c.slug, c.name, c.description, c.sort_order, c.is_active
		FROM catalog.product_categories pc
		JOIN catalog.categories c ON c.id = pc.category_id
		WHERE pc.product_id = ANY($1)
		ORDER BY c.sort_order ASC
	`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var productID uuid.UUID
		var cat Category
		if err := rows.Scan(&productID, &cat.ID, &cat.Slug, &cat.Name, &cat.Description, &cat.SortOrder, &cat.IsActive); err != nil {
			return err
		}
		if i, ok := index[productID]; ok {
			products[i].Categories = append(products[i].Categories, cat)
		}
	}
	return rows.Err()
}

type productScanner interface {
	Scan(dest ...interface{}) error
}

func scanProductRow(row productScanner) (*Product, error) {
	var p Product
	var discountType *string
	var discountValue *int64
	var salePrice *int64
	err := row.Scan(
		&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
		&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive,
		&p.PopularityScore, &p.CreatedAt, &p.UpdatedAt,
		&p.OnSale, &discountType, &discountValue, &salePrice,
	)
	if err != nil {
		return nil, err
	}
	p.DiscountType = discountType
	p.DiscountValue = discountValue
	p.SalePriceKopecks = salePrice
	return &p, nil
}

func scanProducts(rows pgx.Rows, withScores bool) ([]Product, error) {
	products := make([]Product, 0)
	for rows.Next() {
		var p Product
		var discountType *string
		var discountValue *int64
		var salePrice *int64
		dest := []interface{}{
			&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceKopecks, &p.Currency,
			&p.ProductType, &p.DeliveryConfig, &p.ImageKey, &p.IsActive,
			&p.PopularityScore, &p.CreatedAt, &p.UpdatedAt,
			&p.OnSale, &discountType, &discountValue, &salePrice,
		}
		if withScores {
			var rankScore, simScore float64
			dest = append(dest, &rankScore, &simScore)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		p.DiscountType = discountType
		p.DiscountValue = discountValue
		p.SalePriceKopecks = salePrice
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *CatalogRepository) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) ([]Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.slug, p.name, p.description, p.price_kopecks, p.currency,
		       p.product_type::text, p.delivery_config, p.image_key, p.is_active,
		       p.popularity_score, p.created_at, p.updated_at,
		       FALSE, NULL::text, NULL::bigint, NULL::bigint
		FROM catalog.products p
		WHERE p.id = ANY($1) AND p.is_active = TRUE`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows, false)
}

type ProductSnapshot struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	PriceKopecks   int64           `json:"price_kopecks"`
	ProductType    string          `json:"product_type"`
	DeliveryConfig json.RawMessage `json:"delivery_config"`
}

func (r *CatalogRepository) ValidateProducts(ctx context.Context, items map[uuid.UUID]int) ([]ProductSnapshot, int64, error) {
	ids := make([]uuid.UUID, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	products, err := r.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	if len(products) != len(ids) {
		return nil, 0, apperrors.ErrProductNotFound
	}

	productMap := make(map[uuid.UUID]Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	var total int64
	var snapshots []ProductSnapshot
	for id, qty := range items {
		p, ok := productMap[id]
		if !ok {
			return nil, 0, apperrors.ErrProductNotFound
		}
		price := p.PriceKopecks
		if p.SalePriceKopecks != nil {
			price = *p.SalePriceKopecks
		}
		total += price * int64(qty)
		snapshots = append(snapshots, ProductSnapshot{
			ID: p.ID, Name: p.Name, PriceKopecks: price,
			ProductType: p.ProductType, DeliveryConfig: p.DeliveryConfig,
		})
	}
	return snapshots, total, nil
}
