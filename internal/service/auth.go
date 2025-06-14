package service

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"GoShort/pkg/token"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"

	"GoShort/pkg/security"
)

type IAuthService interface {
	GetProfileByID(ctx context.Context, id uuid.UUID) (*dto.ProfileResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
}

type AuthService struct {
	repo     *repository.Queries
	jwtMaker *token.JWTMaker
	log      *logger.Logger
}

func NewAuthService(repo *repository.Queries, jwtMaker *token.JWTMaker, log *logger.Logger) IAuthService {
	return &AuthService{
		repo:     repo,
		jwtMaker: jwtMaker,
		log:      log,
	}
}

func (s *AuthService) GetProfileByID(ctx context.Context, id uuid.UUID) (*dto.ProfileResponse, error) {
	// Get user by ID
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrUserNotFound
		default:
			s.log.Error("failed to retrieve user by ID", "id", id, "error", err)
			return nil, err
		}
	}

	// Map user to profile response
	profile := &dto.ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return profile, nil

}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
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

	profile := &dto.ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return &dto.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
		Data:      *profile,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
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

	return &dto.RegisterResponse{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}
