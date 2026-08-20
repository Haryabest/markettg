package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
)

func InternalAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if secret == "" || c.Get("X-Internal-Secret") != secret {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}
		return c.Next()
	}
}
