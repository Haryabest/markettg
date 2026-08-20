package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
)

type ServiceURLs struct {
	User         string
	Catalog      string
	Order        string
	Payment      string
	Delivery     string
	Notification string
}

func Forward(baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := baseURL + c.OriginalURL()

		// Route admin requests to appropriate services
		if strings.HasPrefix(c.Path(), "/api/v1/admin/orders") ||
			strings.HasPrefix(c.Path(), "/api/v1/admin/promo") {
			// handled by order service via path rewrite in order service
		}

		req, err := http.NewRequestWithContext(c.Context(), c.Method(), targetURL, strings.NewReader(string(c.Body())))
		if err != nil {
			return err
		}

		c.Request().Header.VisitAll(func(key, value []byte) {
			k := string(key)
			if k == "X-User-Id" {
				return
			}
			req.Header.Set(k, string(value))
		})

		if rid := middleware.GetRequestID(c); rid != "" {
			req.Header.Set(middleware.RequestIDHeader, rid)
		}

		if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
			req.Header.Set("X-User-Id", userID)
		}
		if telegramID, ok := c.Locals("telegram_id").(int64); ok && telegramID != 0 {
			req.Header.Set("X-Telegram-Id", fmt.Sprintf("%d", telegramID))
		}
		if adminID, ok := c.Locals("admin_id").(string); ok && adminID != "" {
			req.Header.Set("X-Admin-Id", adminID)
		}
		if role, ok := c.Locals("admin_role").(string); ok && role != "" {
			req.Header.Set("X-Admin-Role", role)
		}
		if botUserID := c.Get("X-Telegram-User-Id"); botUserID != "" {
			req.Header.Set("X-Telegram-User-Id", botUserID)
		}

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "service unavailable")
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		for k, vals := range resp.Header {
			for _, v := range vals {
				c.Set(k, v)
			}
		}

		return c.Status(resp.StatusCode).Send(body)
	}
}
