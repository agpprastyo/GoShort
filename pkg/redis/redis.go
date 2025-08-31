package redis

import (
	"GoShort/config"
	"GoShort/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RdsClient interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Close() error
}

type Redis struct {
	Client *redis.Client
	Config *config.AppConfig
	logger *logger.Logger
}

var _ RdsClient = (*Redis)(nil)

func NewRedis(cfg *config.AppConfig, logger *logger.Logger) (*Redis, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	logger.Infof("Initializing connection to Redis at %s", addr)

	redisClient := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolTimeout:  cfg.Redis.PoolTimeout,
	})

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

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *Redis) Ping(ctx context.Context) error {
	if err := r.Client.Ping(ctx).Err(); err != nil {
		r.logger.Errorf("Error pinging Redis: %v", err)
		return fmt.Errorf("error pinging Redis: %w", err)
	}
	r.logger.Info("Redis connection is alive")
	return nil
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *Redis) Close() error {
	if err := r.Client.Close(); err != nil {
		r.logger.Errorf("Error closing Redis connection: %v", err)
		return fmt.Errorf("error closing Redis connection: %w", err)
	}
	r.logger.Info("Redis connection closed")
	return nil
}
