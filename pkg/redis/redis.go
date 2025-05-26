package redis

import (
	"GoShort/config"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
)

// Redis represents a Redis client instance
type Redis struct {
	Client *redis.Client
	Config *config.AppConfig
	logger *logrus.Logger
}

// NewRedis creates and configures a new Redis client
func NewRedis(cfg *config.AppConfig, logger *logrus.Logger) (*Redis, error) {
	// Build Redis connection configuration
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis")

	return &Redis{
		Client: redisClient,
		Config: cfg,
		logger: logger,
	}, nil
}

// Close closes the Redis client connection
func (r *Redis) Close() error {
	if err := r.Client.Close(); err != nil {
		return fmt.Errorf("error closing Redis connection: %w", err)
	}
	r.logger.Info("Redis connection closed")
	return nil
}
