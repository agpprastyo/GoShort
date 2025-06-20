package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"errors"
	"github.com/google/uuid"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService service.IAuthService
}

func NewAuthHandler(authService service.IAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// UpdatePassword updates the password of the currently authenticated user
func (h *AuthHandler) UpdatePassword(c *fiber.Ctx) error {
	var req dto.UpdatePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	err = h.authService.UpdatePassword(c.Context(), userUUID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Message: "Password updated successfully",
		Data:    nil,
	})
}

// UpdateProfile updates the profile of the currently authenticated user
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	var req dto.UpdateProfileRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Safely get the userID from locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" { // Check if the value exists AND is a non-empty string
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	profile, err := h.authService.UpdateProfile(c.Context(), userUUID, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Server error",
		})
	}

	//return c.JSON(dto.SuccessResponse{
	//	Message: "Profile updated successfully",
	//	Data:    profile,
	//})

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Message: "Profile updated successfully",
		Data:    profile,
	})
}

// Register handles user registration
// @Godoc Register
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User Registration Data"
// @Success 200 {object} dto.SuccessResponse "Successfully registered"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing fields"
// @Failure 409 {object} dto.ErrorResponse "User already exists"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Email, password, and username are required",
		})
	}

	// Attempt registration
	resp, err := h.authService.Register(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error: "User already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Message: "Successfully registered",
		Data:    resp,
	})
}

// GetProfile retrieves the profile of the currently authenticated user
// @Godoc GetProfile
// @Summary Get user profile
// @Description Retrieve the profile of the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponse "Profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized access, user ID not found"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	// Safely get the userID from locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" { // Check if the value exists AND is a non-empty string
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	profile, err := h.authService.GetProfileByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Message: "Profile retrieved successfully",
		Data:    profile,
	})
}

// Login handles user login
// @Godoc Login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User Login Data"
// @Success 200 {object} dto.SuccessResponse "Successfully logged in"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing fields"
// @Failure 401 {object} dto.ErrorResponse "Invalid email or password"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Email and password are required",
		})
	}

	// Attempt login
	response, err := h.authService.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error: "Invalid email or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Server error",
		})
	}

	// Set cookie with the JWT token
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    response.Token,
		Expires:  response.ExpiresAt,
		HTTPOnly: true,
		Secure:   true,  // For HTTPS
		SameSite: "Lax", // Protects against CSRF
		Path:     "/",
	}
	c.Cookie(&cookie)

	return c.JSON(dto.SuccessResponse{
		Message: "Successfully logged in",
		Data: fiber.Map{
			"logged_in":  true,
			"expires_at": response.ExpiresAt,
			"data":       response.Data,
		},
	})
}

// Logout handles user logout
// @Godoc Logout
// @Summary User logout
// @Description Clear user session and delete JWT cookie
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponse "Successfully logged out"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized access, user ID not found"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/logout [delete]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the cookie
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set to a pastime to delete
		HTTPOnly: true,
		Secure:   true,  // For HTTPS
		SameSite: "Lax", // Protects against CSRF
	}
	c.Cookie(&cookie)
	return c.JSON(dto.SuccessResponse{
		Message: "Successfully logged out",
		Data:    nil,
	})

}
