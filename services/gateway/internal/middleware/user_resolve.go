package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type userCacheEntry struct {
	userID    string
	expiresAt time.Time
}

type UserResolver struct {
	userServiceURL string
	internalSecret string
	httpClient     *http.Client
	cache          sync.Map
}

func NewUserResolver(userServiceURL, internalSecret string) *UserResolver {
	return &UserResolver{
		userServiceURL: userServiceURL,
		internalSecret: internalSecret,
		httpClient:     &http.Client{Timeout: 3 * time.Second},
	}
}

func (r *UserResolver) ResolveUserID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			return c.Next()
		}
		tid, ok := c.Locals("telegram_id").(int64)
		if !ok || tid == 0 {
			return c.Next()
		}
		cacheKey := fmt.Sprintf("%d", tid)
		if cached, ok := r.cache.Load(cacheKey); ok {
			entry := cached.(userCacheEntry)
			if time.Now().Before(entry.expiresAt) {
				c.Locals("user_id", entry.userID)
				return c.Next()
			}
		}
		req, err := http.NewRequestWithContext(c.Context(), http.MethodGet,
			fmt.Sprintf("%s/api/v1/internal/users/telegram/%d", r.userServiceURL, tid), nil)
		if err != nil {
			return c.Next()
		}
		req.Header.Set("X-Internal-Secret", r.internalSecret)
		resp, err := r.httpClient.Do(req)
		if err != nil {
			return c.Next()
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return c.Next()
		}
		var user struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil || user.ID == "" {
			return c.Next()
		}
		r.cache.Store(cacheKey, userCacheEntry{userID: user.ID, expiresAt: time.Now().Add(5 * time.Minute)})
		c.Locals("user_id", user.ID)
		return c.Next()
	}
}
