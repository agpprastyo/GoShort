package database

import (
	"GoShort/config"
	"GoShort/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres represents a PostgreSQL database instance
type Postgres struct {
	DB     *pgxpool.Pool // Changed from *sql.DB to *pgxpool.Pool
	Config *config.AppConfig
	log    *logger.Logger
}

// NewPostgres creates and configures a new PostgreSQL database connection
func NewPostgres(cfg *config.AppConfig, log *logger.Logger) (*Postgres, error) {
	// Build PostgreSQL connection string (pgx format)
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode,
	)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL connection string: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.Database.MaxOpenConn)

	// Create connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL database")

	return &Postgres{
		DB:     pool,
		Config: cfg,
		log:    log,
	}, nil
}

// Close closes the database connection
func (p *Postgres) Close() error {
	p.DB.Close()
	p.log.Info("PostgreSQL connection closed")
	return nil
}
