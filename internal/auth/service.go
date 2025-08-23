package auth

import (
	"GoShort/internal/commons"
	"GoShort/internal/datastore"
	"GoShort/pkg/logger"
	"GoShort/pkg/mail"
	"GoShort/pkg/mail/template"
	"GoShort/pkg/token"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"

	"GoShort/pkg/security"
)

type IAuthService interface {
	GetProfileByID(ctx context.Context, id uuid.UUID) (*ProfileResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req UpdateProfileRequest) (*ProfileResponse, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, req UpdatePasswordRequest) error

	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
	ResendVerificationEmail(ctx context.Context, req ResendVerificationRequest) error
	ForgotPasswordToken(ctx context.Context, req ForgotPasswordTokenRequest) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, userUUID uuid.UUID, req ResetPasswordRequest) error
}

type Service struct {
	repo     datastore.Querier
	jwtMaker *token.JWTMaker
	log      *logger.Logger
	mail     mail.IGoogleSMTPService
}

func (s *Service) ResetPassword(ctx context.Context, userUUID uuid.UUID, req ResetPasswordRequest) error {
	user, err := s.repo.GetUser(ctx, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrUserNotFound
		}
		s.log.Error("failed to retrieve user by ID", "user_id", userUUID, "error", err)
		return err
	}

	if !user.IsActive {
		return commons.ErrUserNotActive
	}

	hashedPassword, err := security.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("failed to hash new password", "error", err)
		return err
	}

	_, err = s.repo.UpdateUser(ctx, datastore.UpdateUserParams{
		ID:           user.ID,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		s.log.Error("failed to update user password", "user_id", user.ID, "error", err)
		return err
	}

	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrUserNotFound
		}
		s.log.Error("failed to retrieve user by email", "email", req.Email, "error", err)
		return err
	}

	if !user.IsActive {
		return commons.ErrUserNotActive
	}

	tokenRecord, err := s.repo.GetLatestTokenByUserIDAndType(ctx, datastore.GetLatestTokenByUserIDAndTypeParams{
		UserID: user.ID,
		Type:   datastore.TokenTypePasswordReset,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrTokenNotFound
		}
		s.log.Error("failed to retrieve token for user", "user_id", user.ID, "error", err)
		return err
	}

	if tokenRecord.ExpiresAt.Time.Before(time.Now()) {
		_ = s.repo.DeleteTokenByID(ctx, tokenRecord.ID)
		return commons.ErrTokenExpired
	}

	if tokenRecord.Attempts >= 3 {
		_ = s.repo.DeleteTokenByID(ctx, tokenRecord.ID)
		s.log.Warn("token attempt limit reached, token deleted", "user_id", user.ID)
		return commons.ErrInvalidToken
	}

	if !security.CheckPassword(req.Token, tokenRecord.TokenHash) {
		_ = s.repo.IncrementTokenAttempts(ctx, tokenRecord.ID)
		return commons.ErrInvalidToken
	}

	hashedPassword, err := security.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("failed to hash new password", "error", err)
		return err
	}

	_, err = s.repo.UpdateUser(ctx, datastore.UpdateUserParams{
		ID:           user.ID,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		s.log.Error("failed to update user password", "user_id", user.ID, "error", err)
		return err
	}

	err = s.repo.DeleteTokenByID(ctx, tokenRecord.ID)
	if err != nil {
		s.log.Error("failed to delete used token", "token_id", tokenRecord.ID, "error", err)
	}

	return nil
}

func (s *Service) ForgotPasswordToken(ctx context.Context, req ForgotPasswordTokenRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrUserNotFound
		}
		s.log.Error("failed to retrieve user by email", "email", req.Email, "error", err)
		return err
	}

	if !user.IsActive {
		return commons.ErrUserNotActive
	}

	// Generate the password reset code for the email
	resetCode, err := token.GenerateVerificationCode()
	if err != nil {
		s.log.Error("failed to generate password reset code", "error", err)
		return err
	}

	// Prepare the data for the HTML template
	emailData := template.ResetPasswordData{
		Username: user.Username,
		Token:    resetCode,
	}

	// Generate the HTML body from the embedded template file
	htmlBody, err := template.GenerateResetPasswordHTML(emailData)
	if err != nil {
		s.log.Error("failed to generate password reset email HTML", "error", err)
		return err
	}

	// Define the email subject
	subject := "GoShort Password Reset Request"

	// Use the SendEmail method - this will call the Google SMTP service
	err = s.mail.SendEmail(user.Email, subject, htmlBody)
	if err != nil {
		s.log.Error("failed to send password reset email", "user_email", user.Email, "error", err)
		return err
	}

	s.log.Info("password reset email sent successfully", "user_email", user.Email)

	// Save token to the database
	hashToken, err := security.HashPassword(resetCode)
	if err != nil {
		s.log.Error("failed to hash password reset code", "error", err)
		return err
	}

	tokenID, err := uuid.NewV7()
	if err != nil {
		s.log.Error("failed to generate token ID", "error", err)
		return err
	}

	_, err = s.repo.CreateToken(ctx, datastore.CreateTokenParams{
		ID:        tokenID,
		UserID:    user.ID,
		TokenHash: hashToken,
		Type:      datastore.TokenTypePasswordReset,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	})
	if err != nil {
		s.log.Error("failed to save password reset token", "error", err)
	}

	return nil
}

func (s *Service) ResendVerificationEmail(ctx context.Context, req ResendVerificationRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrUserNotFound
		}
		s.log.Error("failed to retrieve user by email", "email", req.Email, "error", err)
		return err
	}

	if user.IsActive {
		return commons.ErrUserAlreadyActive
	}

	// Generate the verification code for the email
	verificationCode, err := token.GenerateVerificationCode()
	if err != nil {
		s.log.Error("failed to generate verification code", "error", err)
		return err
	}

	// Prepare the data for the HTML template
	emailData := template.RegistrationData{
		Username: user.Username,
		Token:    verificationCode,
	}

	// Generate the HTML body from the embedded template file
	htmlBody, err := template.GenerateRegistrationHTML(emailData)
	if err != nil {
		s.log.Error("failed to generate registration email HTML", "error", err)
		return err
	}

	// Define the email subject
	subject := "Welcome to GoShort! Please Verify Your Account."

	// Use the SendEmail method - this will call the Google SMTP service
	err = s.mail.SendEmail(user.Email, subject, htmlBody)
	if err != nil {
		s.log.Error("failed to send verification email", "user_email", user.Email, "error", err)
		return err
	}

	s.log.Info("verification email sent successfully", "user_email", user.Email)

	// Save token to the database
	hashToken, err := security.HashPassword(verificationCode)
	if err != nil {
		s.log.Error("failed to hash verification code", "error", err)
		return err
	}

	tokenID, err := uuid.NewV7()
	if err != nil {
		s.log.Error("failed to generate token ID", "error", err)
		return err
	}

	_, err = s.repo.CreateToken(ctx, datastore.CreateTokenParams{
		ID:        tokenID,
		UserID:    user.ID,
		TokenHash: hashToken,
		Type:      datastore.TokenTypeRegistrationVerification,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	})
	if err != nil {
		s.log.Error("failed to save verification", "error", err)
	}

	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrUserNotFound
		}
		s.log.Error("failed to retrieve user by email", "email", req.Email, "error", err)
		return err
	}

	if user.IsActive {
		return commons.ErrUserAlreadyActive
	}

	tokenRecord, err := s.repo.GetLatestTokenByUserIDAndType(ctx, datastore.GetLatestTokenByUserIDAndTypeParams{
		UserID: user.ID,
		Type:   datastore.TokenTypeRegistrationVerification,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return commons.ErrTokenNotFound
		}
		s.log.Error("failed to retrieve token for user", "user_id", user.ID, "error", err)
		return err
	}

	if tokenRecord.ExpiresAt.Time.Before(time.Now()) {
		_ = s.repo.DeleteTokenByID(ctx, tokenRecord.ID)
		return commons.ErrTokenExpired
	}

	if tokenRecord.Attempts >= 3 {
		_ = s.repo.DeleteTokenByID(ctx, tokenRecord.ID)
		s.log.Warn("token attempt limit reached, token deleted", "user_id", user.ID)
		return commons.ErrInvalidToken
	}

	if !security.CheckPassword(req.Token, tokenRecord.TokenHash) {
		_ = s.repo.IncrementTokenAttempts(ctx, tokenRecord.ID)
		return commons.ErrInvalidToken
	}

	_, err = s.repo.UpdateUser(ctx, datastore.UpdateUserParams{
		ID:       user.ID,
		IsActive: true,
	})
	if err != nil {
		s.log.Error("failed to activate user account", "user_id", user.ID, "error", err)
		return err
	}

	err = s.repo.DeleteTokenByID(ctx, tokenRecord.ID)
	if err != nil {
		s.log.Error("failed to delete used token", "token_id", tokenRecord.ID, "error", err)
	}

	return nil
}

func NewService(repo datastore.Querier, jwtMaker *token.JWTMaker, log *logger.Logger, mail mail.IGoogleSMTPService) IAuthService {
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

	// In your Register function:
	// ... (code to create the user) ...

	// Send verification code email in a background task
	go func(username, email string, userID uuid.UUID) {
		// Generate the verification code for the email
		verificationCode, err := token.GenerateVerificationCode()
		if err != nil {
			s.log.Error("background: failed to generate verification code", "error", err)
			return // Exit the goroutine
		}

		// Prepare the data for the HTML template
		emailData := template.RegistrationData{
			Username: username,
			Token:    verificationCode,
		}

		// Generate the HTML body from the embedded template file
		htmlBody, err := template.GenerateRegistrationHTML(emailData)
		if err != nil {
			s.log.Error("background: failed to generate registration email HTML", "error", err)
			return // Exit the goroutine
		}

		// Define the email subject
		subject := "Welcome to GoShort! Please Verify Your Account."

		// Use the SendEmail method - this will call the Google SMTP service
		err = s.mail.SendEmail(email, subject, htmlBody)
		if err != nil {
			s.log.Error("background: failed to send verification email", "user_email", email, "error", err)
			return // <-- IMPORTANT: Exit if email fails, so we don't save a token that was never sent
		}

		s.log.Info("background: verification email sent successfully", "user_email", email)

		// Save token to the database
		hashToken, err := security.HashPassword(verificationCode)
		if err != nil {
			s.log.Error("background: failed to hash verification code", "error", err)
			return
		}

		tokenID, err := uuid.NewV7()
		if err != nil {
			s.log.Error("background: failed to generate token ID", "error", err)
			return
		}

		_, err = s.repo.CreateToken(context.Background(), datastore.CreateTokenParams{
			ID:        tokenID,
			UserID:    userID, // Use the userID passed into the goroutine
			TokenHash: hashToken,
			Type:      datastore.TokenTypeRegistrationVerification,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
		})
		if err != nil {
			s.log.Error("background: failed to save verification token", "error", err)
		}

	}(user.Username, user.Email, user.ID) // Pass all required user data into the goroutine

	return &RegisterResponse{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}
