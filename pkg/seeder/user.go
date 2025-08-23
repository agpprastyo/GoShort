package seeder

import (
	"GoShort/internal/datastore"
	"GoShort/pkg/security"

	"github.com/google/uuid"
)

// SeedUsers seeds the database with initial users
func (s *Seeder) SeedUsers() error {
	s.log.Info("Seeding users...")

	adminUUID, err := uuid.NewV7()
	if err != nil {
		s.log.Errorf("Failed to generate UUID for admin user: %v", err)
		return err
	}

	userUUID, err := uuid.NewV7()
	if err != nil {
		s.log.Errorf("Failed to generate UUID for user: %v", err)
		return err
	}

	plainPasswords := []string{"password", "password"} // Plain text passwords
	hashedPasswords := make([]string, len(plainPasswords))

	// Hash all passwords
	for i, pwd := range plainPasswords {
		hash, err := security.HashPassword(pwd)
		if err != nil {
			s.log.Errorf("Failed to hash password: %v", err)
			return err
		}
		hashedPasswords[i] = hash
	}

	// Example implementation based on your datastore methods
	users := []datastore.CreateUserParams{
		{
			ID:           adminUUID,
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: hashedPasswords[0],
			Role:         datastore.UserRoleAdmin,
		},
		{
			ID:           userUUID,
			Username:     "test",
			Email:        "test@example.com",
			PasswordHash: hashedPasswords[1],
			Role:         datastore.UserRoleUser,
		},
	}

	for _, user := range users {
		_, err := s.repo.CreateUser(s.ctx, user)
		if err != nil {
			s.log.Errorf("Failed to seed user %s: %v", user.Username, err)
			return err
		}
	}

	return nil
}
