package service

import (
	"GoShort/internal/repository"
	"GoShort/pkg/token"
	"context"
	"errors"
	"github.com/google/uuid"

	"GoShort/pkg/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenFailed        = errors.New("token generation failed")
)

type AuthService struct {
	repo     *repository.Queries
	jwtMaker *token.JWTMaker
}

func NewAuthService(repo *repository.Queries, jwtMaker *token.JWTMaker) *AuthService {
	return &AuthService{
		repo:     repo,
		jwtMaker: jwtMaker,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	pass := security.CheckPassword(req.Password, user.PasswordHash)
	if !pass {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token using the JWTMaker
	tokenString, expiresAt, err := s.jwtMaker.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Role      string  `json:"role"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Check if user already exists
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	_, err = s.repo.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash the password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var id uuid.UUID

	id, err = uuid.NewV7()
	if err != nil {
		return nil, ErrTokenFailed
	}

	var user repository.User

	// Create new user
	user, err = s.repo.CreateUser(ctx, repository.CreateUserParams{
		ID:           id,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Email:        req.Email,
		Role:         repository.UserRoleUser,
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}
