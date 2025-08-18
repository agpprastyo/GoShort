package middleware

import (
	"GoShort/pkg/logger"
	"GoShort/pkg/token"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware handles JWT authentication via cookies
type AuthMiddleware struct {
	jwtMaker *token.JWTMaker
	log      *logger.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtMaker *token.JWTMaker, log *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtMaker: jwtMaker,
		log:      log,
	}
}

// Authenticate verifies the JWT in the cookie and manages user session
func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get JWT from cookie
		cookie := c.Cookies("access_token")
		m.log.Print("Cookie: ", cookie)
		if cookie == "" {
			m.log.Print("Unauthorized - missing token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - missing token",
			})
		}

		// Verify the token
		payload, err := m.jwtMaker.VerifyToken(cookie)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - invalid token",
			})
		}

		c.Locals("user_id", payload.UserID)
		c.Locals("username", payload.Username)
		c.Locals("role", payload.Role)

		return c.Next()
	}
}
