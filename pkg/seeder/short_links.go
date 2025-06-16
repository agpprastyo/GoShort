package seeder

import (
	"GoShort/internal/repository"
	"GoShort/pkg/security"
	"fmt"
	"github.com/google/uuid"
)

func (s *Seeder) SeedShortLinks() error {
	s.log.Info("Seeding user and 100 short links...")

	// Create a new user
	userUUID, err := uuid.NewV7()
	if err != nil {
		s.log.Errorf("Failed to generate UUID for user: %v", err)
		return err
	}
	passwordHash, err := security.HashPassword("shortlinkuser")
	if err != nil {
		s.log.Errorf("Failed to hash password: %v", err)
		return err
	}
	user := repository.CreateUserParams{
		ID:           userUUID,
		Username:     "shortlinkuser",
		Email:        "shortlinkuser@example.com",
		PasswordHash: passwordHash,
		Role:         repository.UserRoleUser,
	}
	_, err = s.repo.CreateUser(s.ctx, user)
	if err != nil {
		s.log.Errorf("Failed to create user: %v", err)
		return err
	}

	// Create 100 short links for the new user
	for i := 1; i <= 100; i++ {
		linkUUID, err := uuid.NewV7()
		if err != nil {
			s.log.Errorf("Failed to generate UUID for short link: %v", err)
			return err
		}
		params := repository.CreateShortLinkParams{
			ID:          linkUUID,
			UserID:      userUUID,
			ShortCode:   fmt.Sprintf("code%03d", i),
			OriginalUrl: fmt.Sprintf("https://example.com/page/%d", i),
			IsActive:    true,
		}
		_, err = s.repo.CreateShortLink(s.ctx, params)
		if err != nil {
			s.log.Errorf("Failed to create short link %d: %v", i, err)
			return err
		}
	}

	s.log.Info("Seeded 1 user and 100 short links successfully.")
	return nil
}
