package auth

import (
	"GoShort/internal/commons"
	"GoShort/internal/datastore"
	"GoShort/pkg/logger"
	"GoShort/pkg/mail"
	"GoShort/pkg/token"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/google/uuid"

	"GoShort/pkg/security"
)

type IAuthService interface {
	GetProfileByID(ctx context.Context, id uuid.UUID) (*ProfileResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req UpdateProfileRequest) (*ProfileResponse, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, req UpdatePasswordRequest) error
}

type Service struct {
	repo     datastore.Querier
	jwtMaker *token.JWTMaker
	log      *logger.Logger
	mail     mail.ISendGridService
}

func NewService(repo datastore.Querier, jwtMaker *token.JWTMaker, log *logger.Logger, mail mail.ISendGridService) IAuthService {
	return &Service{
		repo:     repo,
		jwtMaker: jwtMaker,
		log:      log,
		mail:     mail,
	}
}

// UpdatePassword updates the password of the currently authenticated user
func (s *Service) UpdatePassword(ctx context.Context, id uuid.UUID, req UpdatePasswordRequest) error {
	// Get user by ID
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return commons.ErrUserNotFound
		default:
			s.log.Error("failed to retrieve user by ID", "id", id, "error", err)
			return err
		}
	}

	// Verify old password
	if !security.CheckPassword(req.OldPassword, user.PasswordHash) {
		return commons.ErrInvalidCredentials
	}

	// Hash the new password
	hashedPassword, err := security.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update user's password in the database
	user, err = s.repo.UpdateUser(ctx, datastore.UpdateUserParams{
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
func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, req UpdateProfileRequest) (*ProfileResponse, error) {
	// Get user by ID
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, commons.ErrUserNotFound
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
	user, err = s.repo.UpdateUser(ctx, datastore.UpdateUserParams{
		ID:        id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	})
	if err != nil {
		s.log.Error("failed to update user profile", "id", id, "error", err)
		return nil, err
	}

	profile := &ProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	return profile, nil
}

func (s *Service) GetProfileByID(ctx context.Context, id uuid.UUID) (*ProfileResponse, error) {
	// Get user by ID
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, commons.ErrUserNotFound
		default:
			s.log.Error("failed to retrieve user by ID", "id", id, "error", err)
			return nil, err
		}
	}

	// Map user to profile response
	profile := &ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return profile, nil

}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, commons.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, commons.ErrUserNotActive
	}

	pass := security.CheckPassword(req.Password, user.PasswordHash)
	if !pass {
		return nil, commons.ErrInvalidCredentials
	}

	tokenString, expiresAt, err := s.jwtMaker.GenerateToken(user)
	if err != nil {
		return nil, commons.ErrTokenFailed
	}

	profile := &ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		Data:      *profile,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, commons.ErrUserAlreadyExists
	}

	_, err = s.repo.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return nil, commons.ErrUserAlreadyExists
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var id uuid.UUID

	id, err = uuid.NewV7()
	if err != nil {
		return nil, commons.ErrTokenFailed
	}

	var user datastore.User

	user, err = s.repo.CreateUser(ctx, datastore.CreateUserParams{
		ID:           id,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Email:        req.Email,
		Role:         datastore.UserRoleUser,
	})

	if err != nil {
		return nil, err
	}

	// Send verification code email
	verificationCode, err := token.GenerateVerificationCode()
	if err != nil {
		s.log.Error("failed to generate verification code", "error", err)
		return nil, commons.ErrTokenFailed
	}

	s.mail.SendEmailWithTemplate(user.Email)

	return &RegisterResponse{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}
