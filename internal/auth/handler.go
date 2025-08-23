package auth

import (
	"GoShort/internal/commons"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	authService IAuthService
	validator   *validator.Validate
}

func NewHandler(authService IAuthService, val *validator.Validate) *Handler {
	return &Handler{authService: authService, validator: val}
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	err = h.authService.ResetPassword(c.Context(), userUUID, req)
	if err != nil {
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		if errors.Is(err, commons.ErrInvalidToken) {
			return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
				Error: "Invalid or expired token",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Password has been reset successfully",
	})

}

func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	err := h.authService.ForgotPassword(c.Context(), req)
	if err != nil {
		if errors.Is(err, commons.ErrInvalidToken) {
			return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
				Error: "Invalid or expired token",
			})
		}
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Password has been reset successfully",
		Data:    nil,
	})
}

// ForgotPasswordToken generates a password reset token and sends it to the user's email
func (h *Handler) ForgotPasswordToken(c *fiber.Ctx) error {

	var req ForgotPasswordTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	err := h.authService.ForgotPasswordToken(c.Context(), req)
	if err != nil {
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Password reset token sent to email if it exists",
		Data:    nil,
	})
}

// ResendVerificationEmail resends the email verification link to the user
func (h *Handler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	err := h.authService.ResendVerificationEmail(c.Context(), req)
	if err != nil {
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		if errors.Is(err, commons.ErrUserAlreadyActive) {
			return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
				Error: "User is already active",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Verification email resent successfully",
		Data:    nil,
	})
}

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {

	var req VerifyEmailRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}
	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	err := h.authService.VerifyEmail(c.Context(), req)
	if err != nil {
		if errors.Is(err, commons.ErrInvalidToken) {
			return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
				Error: "Invalid or expired token",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Email verified successfully",
		Data:    nil,
	})
}

// UpdatePassword updates the password of the currently authenticated user
// @Godoc UpdatePassword
// @Summary Update user password
// @Description Update the password for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UpdatePasswordRequest true "Update Password Data"
// @Success 200 {object} dto.SuccessResponse "Password updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing fields"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized access, user ID not found"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/update-password [put]
// @Security ApiKeyAuth
func (h *Handler) UpdatePassword(c *fiber.Ctx) error {
	var req UpdatePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	err = h.authService.UpdatePassword(c.Context(), userUUID, req)
	if err != nil {
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Password updated successfully",
		Data:    nil,
	})
}

// UpdateProfile updates the profile of the currently authenticated user
// @Godoc UpdateProfile
// @Summary Update user profile
// @Description Update the profile information for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update Profile Data"
// @Success 200 {object} dto.SuccessResponse "Profile updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing fields"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized access, user ID not found"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/update-profile [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	var req UpdateProfileRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	// Safely get the userID from locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" { // Check if the value exists AND is a non-empty string
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	profile, err := h.authService.UpdateProfile(c.Context(), userUUID, req)
	if err != nil {
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	//return c.JSON(dto.SuccessResponse{
	//	Message: "Profile updated successfully",
	//	Data:    profile,
	//})

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
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
// @Success 200 {object} dto.SuccessResponse "Successfully registered, please verify your email"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing fields"
// @Failure 409 {object} dto.ErrorResponse "User already exists"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /api/v1/register [post]
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	_, err := h.authService.Register(c.Context(), req)
	if err != nil {
		if errors.Is(err, commons.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(commons.ErrorResponse{
				Error: "User already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.JSON(commons.SuccessResponse{
		Message: "Successfully registered, please verify your email",
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
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	// Safely get the userID from locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{
			Error: "Unauthorized access, user ID not found",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	profile, err := h.authService.GetProfileByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, commons.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Server error",
		})
	}

	return c.JSON(commons.SuccessResponse{
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
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		fieldErrors := commons.FormatValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Message: "Validation failed",
			Error:   fieldErrors,
		})
	}

	response, err := h.authService.Login(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, commons.ErrInvalidCredentials):
			return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{
				Error: "Invalid email or password",
			})
		case errors.Is(err, commons.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "User not found",
			})
		case errors.Is(err, commons.ErrUserNotActive):
			return c.Status(fiber.StatusForbidden).JSON(commons.ErrorResponse{
				Error: "User account is not active",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
				Error: "Server error",
			})
		}
	}

	// Set cookie with the JWT token
	cookie := fiber.Cookie{
		Name:     "access_token",
		Value:    response.Token,
		Expires:  response.ExpiresAt,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	}

	c.Cookie(&cookie)

	return c.JSON(commons.SuccessResponse{
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
func (h *Handler) Logout(c *fiber.Ctx) error {
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
	return c.JSON(commons.SuccessResponse{
		Message: "Successfully logged out",
		Data:    nil,
	})

}
