package api

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

	//// session fiber middleware
	//storage := redisFiber.New(redisFiber.Config{
	//	Host: redisClient.Config.Redis.Host,
	//	Port: func() int {
	//		port, _ := strconv.Atoi(redisClient.Config.Redis.Port)
	//		return port
	//	}(),
	//	Password: redisClient.Config.Redis.Password,
	//	Database: redisClient.Config.Redis.DB,
	//	PoolSize: redisClient.Config.Redis.PoolSize,
	//})
	//store := session.New(session.Config{
	//	Storage: storage,
	//})

	app.Use(swagger.New(cfg))
	// Base API group
	api := app.Group("/api/v1")

	// Register all handlers

	registerAuthHandlers(api, db, jwtMaker, logger)
	registerAdminRoutes(api, db, middleware.NewAuthMiddleware(jwtMaker, logger), logger)
	registerUserRoutes(api, db, middleware.NewAuthMiddleware(jwtMaker, logger), logger)
}

// Add this function to routes.go
func registerUserRoutes(router fiber.Router, db *database.Postgres, authMiddleware *middleware.AuthMiddleware, log *logger.Logger) {
	// Create dependencies
	queries := repository.New(db.DB)
	shortLinkService := service.NewShortLinkService(queries)
	shortLinkHandler := handler.NewShortLinkHandler(shortLinkService)

	// User routes - all require authentication
	userRoutes := router.Group("/links")
	userRoutes.Use(authMiddleware.Authenticate())

	// Short link management routes
	userRoutes.Get("/", shortLinkHandler.GetUserLinks)
	userRoutes.Post("/", shortLinkHandler.CreateShortLink)
	userRoutes.Put("/:id", shortLinkHandler.UpdateLink)
	userRoutes.Delete("/:id", shortLinkHandler.DeleteLink)
	userRoutes.Post("/:id/:status", shortLinkHandler.ToggleLinkStatus)
}

func registerAuthHandlers(router fiber.Router, db *database.Postgres, jwtMaker *token.JWTMaker, log *logger.Logger) {
	// Create dependencies
	queries := repository.New(db.DB)
	authService := service.NewAuthService(queries, jwtMaker, log)
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(jwtMaker, log)

	// Auth routes directly on the router (no auth group)
	router.Post("/login", authHandler.Login)
	router.Post("/register", authHandler.Register)
	router.Delete("/logout", authHandler.Logout)

	// Protected routes with middleware
	router.Get("/profile", authMiddleware.Authenticate(), func(ctx *fiber.Ctx) error {
		userID := ctx.Locals("user_id")
		username := ctx.Locals("username")
		role := ctx.Locals("role")

		return ctx.JSON(fiber.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
			"message":  "Protected route accessed",
		})
	})

}

// Admin routes for managing short links
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
