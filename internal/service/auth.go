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
	UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, req dto.UpdatePasswordRequest) error
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

// UpdatePassword updates the password of the currently authenticated user
func (s *AuthService) UpdatePassword(ctx context.Context, id uuid.UUID, req dto.UpdatePasswordRequest) error {
	// Get user by ID
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrUserNotFound
		default:
			s.log.Error("failed to retrieve user by ID", "id", id, "error", err)
			return err
		}
	}

	// Verify old password
	if !security.CheckPassword(req.OldPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	// Hash the new password
	hashedPassword, err := security.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update user's password in the database
	user, err = s.repo.UpdateUser(ctx, repository.UpdateUserParams{
		ID:           id,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		s.log.Error("failed to update user password", "id", id, "error", err)
		return err
	}

	return nil
}

// UpdateProfile updates the profile of the currently authenticated user
func (s *AuthService) UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
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

	// Update user fields if provided
	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}
	if req.Username != nil {
		user.Username = *req.Username
	}

	// Save updated user to the database
	user, err = s.repo.UpdateUser(ctx, repository.UpdateUserParams{
		ID:        id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	})
	if err != nil {
		s.log.Error("failed to update user profile", "id", id, "error", err)
		return nil, err
	}

	profile := &dto.ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	return profile, nil
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
		ExpiresAt: expiresAt,
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
