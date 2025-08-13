// cmd/seeder/main.go
package main

import (
	"GoShort/config"
	"GoShort/internal/server"
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"GoShort/pkg/seeder"
)

func main() {
	// Load environment variables
	server.LoadEnv()

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
