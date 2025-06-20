package server

import (
	"GoShort/config"
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"GoShort/pkg/redis"
	"GoShort/pkg/token"
	"context"
	"errors"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

// App holds all application dependencies
type App struct {
	Config   *config.AppConfig
	Logger   *logger.Logger
	DB       *database.Postgres
	Redis    redis.RdsClient
	FiberApp *fiber.App
	JWTMaker *token.JWTMaker
}

func InitApp() *App {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg)

	// Initialize PostgreSQL
	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}

	// Initialize Redis
	redisClient, err := redis.NewRedis(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "GoShort",
		ErrorHandler: CustomErrorHandler(log),
	})

	// Initialize JWT Maker
	jwtMaker := token.NewJWTMaker(cfg)

	return &App{
		Config:   cfg,
		Logger:   log,
		DB:       db,
		Redis:    redisClient,
		FiberApp: fiberApp,
		JWTMaker: jwtMaker,
	}
}

func StartServer(app *App) {
	// Setup middleware
	SetupMiddleware(app)

	// Setup routes
	SetupRoutes(app.FiberApp, app.Logger, app.DB, app.Redis, app.JWTMaker, app.Config)

	// Start app
	app.Logger.Infof("Starting app on port %s...", app.Config.Server.Port)
	if err := app.FiberApp.Listen(":" + app.Config.Server.Port); err != nil {
		app.Logger.Fatalf("Error starting app: %v", err)
	}
}

func SetupMiddleware(app *App) {
	app.FiberApp.Use(fiberlog.New())
	app.FiberApp.Use(recover.New())

}

func Cleanup(app *App) {
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			app.Logger.Errorf("Error closing DB: %v", err)
		}
	}

	if app.Redis != nil {
		if err := app.Redis.Close(); err != nil {
			app.Logger.Errorf("Error closing Redis: %v", err)
		}
	}
}

func WaitForShutdown(app *App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down app...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.FiberApp.ShutdownWithContext(ctx); err != nil {
		app.Logger.Fatalf("Server shutdown failed: %v", err)
	}
}

func CustomErrorHandler(logger *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		logger.Errorf("Error 1: %v", err)

		var e *fiber.Error
		if errors.As(err, &e) {
			return c.Status(e.Code).JSON(fiber.Map{
				"error": e.Message,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})

	}
}
