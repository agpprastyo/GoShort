package health

import (
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"GoShort/pkg/redis"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler manages health check endpoints
type HealthHandler struct {
	logger *logger.Logger
	db     *database.Postgres
	redis  *redis.Redis
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(logger *logger.Logger, db *database.Postgres, redis *redis.Redis) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		db:     db,
		redis:  redis,
	}
}

// Status represents the health status of a service component
type Status struct {
	Status    string `json:"status"`
	Component string `json:"component"`
	Message   string `json:"message,omitempty"`
}

// Check handles the health check endpoint
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	var statuses []Status

	// Check database
	if dbStatus := h.checkDatabase(); dbStatus != nil {
		statuses = append(statuses, *dbStatus)
	}

	// Check Redis
	if redisStatus := h.checkRedis(); redisStatus != nil {
		statuses = append(statuses, *redisStatus)
	}

	// Overall health status
	healthy := true
	for _, s := range statuses {
		if s.Status != "healthy" {
			healthy = false
			break
		}
	}

	statusCode := fiber.StatusOK
	overallStatus := "healthy"
	if !healthy {
		statusCode = fiber.StatusServiceUnavailable
		overallStatus = "unhealthy"
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status":     overallStatus,
		"timestamp":  time.Now().Format(time.RFC3339),
		"components": statuses,
	})
}

func (h *HealthHandler) checkDatabase() *Status {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.db.DB.Ping(ctx)
	if err != nil {
		h.logger.Errorf("Database health check failed: %v", err)
		return &Status{
			Status:    "unhealthy",
			Component: "postgres",
			Message:   err.Error(),
		}
	}

	return &Status{
		Status:    "healthy",
		Component: "postgres",
	}
}

func (h *HealthHandler) checkRedis() *Status {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.redis.Client.Ping(ctx).Err()
	if err != nil {
		h.logger.Errorf("Redis health check failed: %v", err)
		return &Status{
			Status:    "unhealthy",
			Component: "redis",
			Message:   err.Error(),
		}
	}

	return &Status{
		Status:    "healthy",
		Component: "redis",
	}
}
