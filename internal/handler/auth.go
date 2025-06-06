package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
// @Godoc Register a new user
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User Registration Data"
// @Success 200 {object} map[string]interface{} "Successfully registered"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "User already exists"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, password, and username are required",
		})
	}

	// Attempt registration
	resp, err := h.authService.Register(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "User already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User registered successfully",
		"data":    resp,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Attempt login
	response, err := h.authService.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	// Set cookie with the JWT token
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    response.Token,
		Expires:  time.Unix(response.ExpiresAt, 0),
		HTTPOnly: true,
		Secure:   true,  // For HTTPS
		SameSite: "Lax", // Protects against CSRF
		Path:     "/",
	}
	c.Cookie(&cookie)

	// Return minimal response (no token in body)
	return c.JSON(fiber.Map{
		"logged_in":  true,
		"expires_at": response.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the cookie
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set to a past time to delete
		HTTPOnly: true,
		Secure:   true,  // For HTTPS
		SameSite: "Lax", // Protects against CSRF
	}
	c.Cookie(&cookie)
	return c.JSON(fiber.Map{
		"logged_out": true,
		"message":    "Successfully logged out",
	})

}
