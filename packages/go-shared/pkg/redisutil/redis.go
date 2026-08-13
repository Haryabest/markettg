package redisutil

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

func NewClient(cfg Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func Ping(ctx context.Context, client *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}

type Lock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewLock(client *redis.Client, key, value string, ttl time.Duration) *Lock {
	return &Lock{client: client, key: key, value: value, ttl: ttl}
}

func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	return l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
}

func (l *Lock) Release(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	return err
}

func RateLimitKey(ip, route string) string {
	return fmt.Sprintf("rate:%s:%s", ip, route)
}

func CartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}

func DeliveryLockKey(orderID string) string {
	return fmt.Sprintf("lock:delivery:%s", orderID)
}

const (
	StreamPayments      = "events:payments"
	StreamOrders        = "events:orders"
	StreamDelivery      = "events:delivery"
	StreamNotifications = "events:notifications"
)

func OrderWSChannel(orderID string) string {
	return fmt.Sprintf("ws:order:%s", orderID)
}
