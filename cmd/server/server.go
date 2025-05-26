package main

import (
	"GoShort/api"
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
	Redis    *redis.Redis
	FiberApp *fiber.App
	JWTMaker *token.JWTMaker
}

func initApp() *App {
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
	redisClient, err := redis.NewRedis(cfg, log.Logger)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "GoShort",
		ErrorHandler: customErrorHandler(log),
	})

	// Initialize JWT Maker
	jwtMaker := token.NewJWTMaker(&cfg.JWT)

	return &App{
		Config:   cfg,
		Logger:   log,
		DB:       db,
		Redis:    redisClient,
		FiberApp: fiberApp,
		JWTMaker: jwtMaker,
	}
}

func startServer(app *App) {
	// Setup middleware
	setupMiddleware(app)

	// Setup routes
	api.SetupRoutes(app.FiberApp, app.Logger, app.DB, app.Redis, app.JWTMaker)

	// Start server
	app.Logger.Infof("Starting server on port %s...", app.Config.Server.Port)
	if err := app.FiberApp.Listen(":" + app.Config.Server.Port); err != nil {
		app.Logger.Fatalf("Error starting server: %v", err)
	}
}

func setupMiddleware(app *App) {
	app.FiberApp.Use(fiberlog.New())
	app.FiberApp.Use(recover.New())

}

func cleanup(app *App) {
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

func waitForShutdown(app *App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.FiberApp.ShutdownWithContext(ctx); err != nil {
		app.Logger.Fatalf("Server shutdown failed: %v", err)
	}
}

func customErrorHandler(logger *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		logger.Errorf("Error: %v", err)

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
