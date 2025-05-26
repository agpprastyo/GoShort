// cmd/seeder/main.go
package main

import (
	"GoShort/config"
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"GoShort/pkg/seeder"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg)

	// Initialize PostgreSQL
	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	defer db.Close()

	// Create and run seeder
	s := seeder.NewSeeder(db, log)
	if err := s.SeedAll(); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Info("Seeding completed successfully")
}
