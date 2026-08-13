package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyCategories = "catalog:categories"
	keyPromotions   = "catalog:promotions"
	keyProduct      = "catalog:product:%s"
	keyProducts     = "catalog:products:%s"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{client: client, ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Cache) InvalidateCatalog(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, "catalog:*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func CategoriesKey() string {
	return keyCategories
}

func PromotionsKey() string {
	return keyPromotions
}

func ProductKey(id string) string {
	return fmt.Sprintf(keyProduct, id)
}

func ProductsKey(params string) string {
	sum := sha256.Sum256([]byte(params))
	return fmt.Sprintf(keyProducts, hex.EncodeToString(sum[:16]))
}
