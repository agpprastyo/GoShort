package main

import (
	"GoShort/internal/handler"
	"GoShort/internal/health"
	"GoShort/internal/middleware"
	"GoShort/internal/repository"
	"GoShort/internal/service"
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"GoShort/pkg/redis"
	"GoShort/pkg/token"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

// SetupRoutes registers all application routes
func SetupRoutes(app *fiber.App, logger *logger.Logger, db *database.Postgres, redisClient *redis.Redis, jwtMaker *token.JWTMaker) {
	healthHandler := health.NewHealthHandler(logger, db, redisClient)
	app.Get("/health", healthHandler.Check)
	app.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))

	cfg := swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}
	app.Use(swagger.New(cfg))

	// Create redirect handler
	// In routes.go
	redirectService := service.NewRedirectService(repository.New(db.DB), logger)
	redirectHandler := handler.NewRedirectHandler(redirectService, logger)

	// Add redirect route at root level (must be before other routes)
	app.Get("/:code", redirectHandler.RedirectToOriginalURL)

	// Base API group
	api := app.Group("/api/v1")

	// Register all handlers

	registerAuthHandlers(api, db, jwtMaker, logger)
	registerAdminRoutes(api, db, middleware.NewAuthMiddleware(jwtMaker, logger), logger)
	registerUserRoutes(api, db, middleware.NewAuthMiddleware(jwtMaker, logger), logger)
}

// registerAuthHandlers sets up authentication routes
func registerAuthHandlers(router fiber.Router, db *database.Postgres, jwtMaker *token.JWTMaker, log *logger.Logger) {
	// Create dependencies
	queries := repository.New(db.DB)
	authService := service.NewAuthService(queries, jwtMaker, log)
	authHandler := handler.NewAuthHandler(authService)
	//authMiddleware := middleware.NewAuthMiddleware(jwtMaker, log)

	// Auth routes directly on the router (no auth group)
	router.Post("/login", authHandler.Login)
	router.Post("/register", authHandler.Register)
	router.Delete("/logout", authHandler.Logout)

}

// registerUserRoutes sets up routes for authenticated users to manage their short links
func registerUserRoutes(router fiber.Router, db *database.Postgres, authMiddleware *middleware.AuthMiddleware, log *logger.Logger) {
	// Create dependencies
	queries := repository.New(db.DB)
	shortLinkService := service.NewShortLinkService(queries, log)
	shortLinkHandler := handler.NewShortLinkHandler(shortLinkService, log)

	// User routes - all require authentication
	userRoutes := router.Group("/links")
	userRoutes.Use(authMiddleware.Authenticate())

	// Short link management routes
	userRoutes.Get("/", shortLinkHandler.GetUserLinks)
	userRoutes.Get("/:id", shortLinkHandler.GetUserLinkByID)
	userRoutes.Post("/", shortLinkHandler.CreateShortLink)
	userRoutes.Put("/:id", shortLinkHandler.UpdateLink)
	userRoutes.Delete("/:id", shortLinkHandler.DeleteLink)
	userRoutes.Patch("/:id/status", shortLinkHandler.ToggleLinkStatus)
}

// registerAdminRoutes sets up routes for admin users to manage the application
func registerAdminRoutes(router fiber.Router, db *database.Postgres, authMiddleware *middleware.AuthMiddleware, log *logger.Logger) {
	//// Create repository
	//queries := repository.New(db.DB)
	//
	//service := service.NewAdminService(queries, log)
	//
	//// Create admin handler with repository
	//adminHandler := handler.NewAdminHandler(service)

	roleMiddleware := middleware.NewRoleMiddleware()

	router.Get("/test-admin", authMiddleware.Authenticate(), roleMiddleware.RequireAdmin(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Admin route accessed",
		})
	})

	// Admin routes - all protected with auth + admin check
	//router.Get("/admin/links", authMiddleware.Authenticate(), adminMiddleware, adminHandler.ListAllLinks)
	//router.Get("/admin/links/:id", authMiddleware.Authenticate(), adminMiddleware, adminHandler.GetLink)
	//router.Get("/admin/users/:userId/links", authMiddleware.Authenticate(), adminMiddleware, adminHandler.ListUserLinks)
	//router.Put("/admin/links/:id", authMiddleware.Authenticate(), adminMiddleware, adminHandler.UpdateLink)
	//router.Delete("/admin/links/:id", authMiddleware.Authenticate(), adminMiddleware, adminHandler.DeleteLink)
	//router.Post("/admin/links/:id/deactivate", authMiddleware.Authenticate(), adminMiddleware, adminHandler.DeactivateLink)
	//router.Get("/admin/stats", authMiddleware.Authenticate(), adminMiddleware, adminHandler.GetStats)
}
