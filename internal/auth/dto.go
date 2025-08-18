package auth

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string          `json:"token"`
	ExpiresAt time.Time       `json:"expires_at"`
	Data      ProfileResponse `json:"data"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type RegisterResponse struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Role      string  `json:"role"`
}

type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	Role      string    `json:"role"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=3,max=50"`
	LastName  *string `json:"last_name" validate:"omitempty,min=3,max=50"`
	Username  *string `json:"username" validate:"omitempty,min=3,max=50"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6,max=100"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=100, differs=OldPassword"`
}
