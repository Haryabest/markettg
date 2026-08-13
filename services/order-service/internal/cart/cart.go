package cart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/redis/go-redis/v9"
)

type Item struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type Cart struct {
	Items     []Item    `json:"items"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	redis *redis.Client
	ttl   time.Duration
}

func New(redis *redis.Client) *Store {
	return &Store{redis: redis, ttl: 30 * 24 * time.Hour}
}

func (s *Store) Get(ctx context.Context, userID string) (*Cart, error) {
	data, err := s.redis.Get(ctx, redisutil.CartKey(userID)).Bytes()
	if err == redis.Nil {
		return &Cart{Items: []Item{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cart Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func (s *Store) Save(ctx context.Context, userID string, cart *Cart) error {
	cart.UpdatedAt = time.Now()
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, redisutil.CartKey(userID), data, s.ttl).Err()
}

func (s *Store) Clear(ctx context.Context, userID string) error {
	return s.redis.Del(ctx, redisutil.CartKey(userID)).Err()
}

func (s *Store) UpsertItem(ctx context.Context, userID string, productID uuid.UUID, quantity int) (*Cart, error) {
	cart, err := s.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			if quantity <= 0 {
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			} else {
				cart.Items[i].Quantity = quantity
			}
			found = true
			break
		}
	}
	if !found && quantity > 0 {
		cart.Items = append(cart.Items, Item{ProductID: productID, Quantity: quantity})
	}
	if err := s.Save(ctx, userID, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *Store) RemoveItem(ctx context.Context, userID string, productID uuid.UUID) (*Cart, error) {
	return s.UpsertItem(ctx, userID, productID, 0)
}

func (c *Cart) ToMap() map[uuid.UUID]int {
	m := make(map[uuid.UUID]int)
	for _, item := range c.Items {
		m[item.ProductID] = item.Quantity
	}
	return m
}
