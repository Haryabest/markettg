package proxy

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AdminForward(urls ServiceURLs) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		var target string
		switch {
		case strings.Contains(path, "/admin/orders") || strings.Contains(path, "/admin/promo"):
			target = urls.Order
		case strings.Contains(path, "/admin/users") || strings.Contains(path, "/admin/audit"):
			target = urls.User
		case strings.Contains(path, "/admin/payments"):
			target = urls.Payment
		case strings.Contains(path, "/admin/deliveries"):
			target = urls.Delivery
		default:
			target = urls.Catalog
		}
		return Forward(target)(c)
	}
}
