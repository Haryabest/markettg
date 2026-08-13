CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS catalog;

CREATE TYPE catalog.product_type AS ENUM ('STARS', 'PREMIUM', 'GIFT');

CREATE TABLE catalog.categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE catalog.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_kopecks BIGINT NOT NULL CHECK (price_kopecks >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    product_type catalog.product_type NOT NULL,
    delivery_config JSONB NOT NULL DEFAULT '{}',
    image_key TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    popularity_score INT NOT NULL DEFAULT 0,
    search_vector TSVECTOR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_search ON catalog.products USING GIN(search_vector);
CREATE INDEX idx_products_name_trgm ON catalog.products USING GIN(name gin_trgm_ops);
CREATE INDEX idx_products_active ON catalog.products(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_products_popularity ON catalog.products(popularity_score DESC);
CREATE INDEX idx_products_type ON catalog.products(product_type);

CREATE TABLE catalog.product_categories (
    product_id UUID NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES catalog.categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

CREATE TYPE catalog.discount_type AS ENUM ('PERCENT', 'FIXED');

CREATE TABLE catalog.promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount_type catalog.discount_type NOT NULL,
    discount_value BIGINT NOT NULL CHECK (discount_value > 0),
    banner_key TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at)
);

CREATE TABLE catalog.promotion_products (
    promotion_id UUID NOT NULL REFERENCES catalog.promotions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    PRIMARY KEY (promotion_id, product_id)
);

CREATE OR REPLACE FUNCTION catalog.products_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', coalesce(NEW.name, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(NEW.slug, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER products_search_vector_trigger
    BEFORE INSERT OR UPDATE ON catalog.products
    FOR EACH ROW EXECUTE FUNCTION catalog.products_search_vector_update();
