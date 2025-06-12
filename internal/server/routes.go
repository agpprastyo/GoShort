package server

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
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	api.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173", // Allow all origins for development and localhost:5173
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Access-Control-Allow-Methods",
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight response for 5 minutes
	}))

	// auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtMaker, logger)

	registerAuthHandlers(api, db, jwtMaker, logger, authMiddleware)
	registerAdminRoutes(api, db, authMiddleware, logger)
	registerUserRoutes(api, db, authMiddleware, logger)
}

// registerAuthHandlers sets up authentication routes
func registerAuthHandlers(router fiber.Router, db *database.Postgres, jwtMaker *token.JWTMaker, log *logger.Logger, authMiddleware *middleware.AuthMiddleware) {
	// Create dependencies
	queries := repository.New(db.DB)
	authService := service.NewAuthService(queries, jwtMaker, log)
	authHandler := handler.NewAuthHandler(authService)

	// Auth routes directly on the router (no auth group)
	router.Post("/login", authHandler.Login)
	router.Post("/register", authHandler.Register)

	router.Get("/profile", authMiddleware.Authenticate(), authHandler.GetProfile)
	router.Delete("/logout", authMiddleware.Authenticate(), authHandler.Logout)

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
	// Create repository
	queries := repository.New(db.DB)

	// Create admin service
	adminService := service.NewAdminService(queries, log)

	// Create admin handler with service
	adminHandler := handler.NewAdminHandler(adminService, log)

	// Role middleware for admin checks
	roleMiddleware := middleware.NewRoleMiddleware()

	// Admin routes - all protected with auth + admin role check
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(authMiddleware.Authenticate(), roleMiddleware.RequireAdmin())

	// Test endpoint
	adminRoutes.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Admin route accessed",
		})
	})
	//
	//// Link management routes
	adminRoutes.Get("/links", adminHandler.ListAllLinks)
	adminRoutes.Get("/links/:id", adminHandler.GetLink)
	adminRoutes.Get("/users/:userId/links", adminHandler.ListUserLinks)

	adminRoutes.Patch("/links/:id/status", adminHandler.ToggleLinkStatus)
	//
	//// Stats route
	//adminRoutes.Get("/stats", adminHandler.GetStats)
}
