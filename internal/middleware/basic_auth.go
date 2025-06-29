package middleware

import (
	"GoShort/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// BasicAuth returns a Fiber handler for Basic Auth protection.
// Customize the users map as needed.
func BasicAuth(c *config.AppConfig) fiber.Handler {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			c.BasicAuth.Username: c.BasicAuth.Password,
		},
		Realm: "Restricted",
	})
}
