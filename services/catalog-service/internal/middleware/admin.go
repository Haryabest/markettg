package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/httputil"
)

const AdminRoleHeader = "X-Admin-Role"

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "superadmin"
	RoleManager    AdminRole = "manager"
	RoleViewer     AdminRole = "viewer"
)

var validRoles = map[AdminRole]int{
	RoleViewer:     1,
	RoleManager:    2,
	RoleSuperAdmin: 3,
}

func GetAdminRole(c *fiber.Ctx) (AdminRole, error) {
	role := AdminRole(strings.ToLower(strings.TrimSpace(c.Get(AdminRoleHeader))))
	if role == "" {
		return "", apperrors.ErrUnauthorized
	}
	if _, ok := validRoles[role]; !ok {
		return "", apperrors.ErrForbidden
	}
	return role, nil
}

func RequireAdmin(minRole AdminRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, err := GetAdminRole(c)
		if err != nil {
			return httputil.Error(c, err)
		}
		if validRoles[role] < validRoles[minRole] {
			return httputil.Error(c, apperrors.ErrForbidden)
		}
		c.Locals("admin_role", string(role))
		if id := c.Get("X-Admin-Id"); id != "" {
			c.Locals("admin_id", id)
		}
		return c.Next()
	}
}
