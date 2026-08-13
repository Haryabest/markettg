package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/redisutil"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, limit: limit, window: window}
}

func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rl.client == nil {
			return c.Next()
		}
		key := redisutil.RateLimitKey(c.IP(), c.Path())
		ctx := context.Background()

		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}
		if count == 1 {
			rl.client.Expire(ctx, key, rl.window)
		}
		if count > int64(rl.limit) {
			return httputil.Error(c, apperrors.ErrRateLimited)
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.limit-int(count)))
		return c.Next()
	}
}
