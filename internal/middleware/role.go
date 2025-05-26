package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// RoleMiddleware handles role-based access control
type RoleMiddleware struct {
	// Add any dependencies here

}

// NewRoleMiddleware creates a new role middleware
func NewRoleMiddleware() *RoleMiddleware {
	return &RoleMiddleware{}
}

// RequireRole checks if the user has the required role
func (m *RoleMiddleware) RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := fmt.Sprintf("%v", c.Locals("role"))
		if role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fmt.Sprintf("%s role required", requiredRole),
			})
		}
		return c.Next()
	}
}

// RequireAdmin is a shorthand for admin role check
func (m *RoleMiddleware) RequireAdmin() fiber.Handler {
	return m.RequireRole("admin")
}

// RequireUser is a shorthand for user role check
func (m *RoleMiddleware) RequireUser() fiber.Handler {
	return m.RequireRole("user")
}
