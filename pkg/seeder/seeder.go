package seeder

import (
	"GoShort/internal/repository"
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"context"
)

// Seeder handles database seeding operations
type Seeder struct {
	db   *database.Postgres
	log  *logger.Logger
	repo *repository.Queries
	ctx  context.Context
}

// NewSeeder creates a new database seeder
func NewSeeder(db *database.Postgres, log *logger.Logger) *Seeder {
	ctx := context.Background()
	repo := repository.New(db.DB)

	return &Seeder{
		db:   db,
		log:  log,
		repo: repo,
		ctx:  ctx,
	}
}

// SeedAll runs all seeding functions
func (s *Seeder) SeedAll() error {
	s.log.Info("Starting database seeding...")

	//if err := s.SeedUsers(); err != nil {
	//	return err
	//}

	if err := s.SeedShortLinks(); err != nil {
		return err
	}

	s.log.Info("Database seeding completed successfully")
	return nil
}
