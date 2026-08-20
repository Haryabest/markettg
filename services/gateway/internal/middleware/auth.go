package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
	"github.com/markettg/markettg/packages/go-shared/pkg/telegram"
)

type AuthConfig struct {
	BotToken       string
	BotSecret      string
	JWTSecret      string
	AdminJWTSecret string
}

type Auth struct {
	cfg AuthConfig
}

func NewAuth(cfg AuthConfig) *Auth {
	return &Auth{cfg: cfg}
}

func (a *Auth) TelegramAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "tma ") {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}

		initData := strings.TrimPrefix(authHeader, "tma ")
		data, err := telegram.ValidateInitData(initData, a.cfg.BotToken, 24*time.Hour)
		if err != nil || data.User == nil {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}

		c.Locals("telegram_id", data.User.ID)
		c.Locals("telegram_user", data.User)
		return c.Next()
	}
}

func (a *Auth) AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(a.cfg.AdminJWTSecret), nil
		})
		if err != nil || !token.Valid {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}

		if sub, ok := claims["sub"].(string); ok {
			c.Locals("admin_id", sub)
		}
		if role, ok := claims["role"].(string); ok {
			c.Locals("admin_role", role)
		}
		return c.Next()
	}
}

func (a *Auth) BotAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("X-Bot-Secret") != a.cfg.BotSecret {
			return httputil.Error(c, apperrors.ErrUnauthorized)
		}
		return c.Next()
	}
}
