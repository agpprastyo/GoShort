package server

import (
	config "GoShort/config"
	_ "GoShort/docs" // Import generated Swagger docs
	"GoShort/internal/dto"
	"GoShort/internal/handler"
	"GoShort/internal/health"
	"GoShort/internal/middleware"
	"GoShort/internal/repository"
	"GoShort/internal/service"
	"GoShort/pkg/database"
	"GoShort/pkg/logger"
	"GoShort/pkg/redis"
	"GoShort/pkg/token"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	redisFiber "github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/swagger"
	"runtime"
	"strconv"
)

// SetupRoutes registers all application routes

func SetupRoutes(app *fiber.App, logger *logger.Logger, db *database.Postgres, redisClient redis.RdsClient, jwtMaker *token.JWTMaker, cfg *config.AppConfig) {
	port, _ := strconv.Atoi(cfg.Redis.Port)

	storage := redisFiber.New(
		redisFiber.Config{
			Host:      cfg.Redis.Host,
			Port:      port,
			Password:  cfg.Redis.Password,
			Database:  cfg.Redis.DB,
			Reset:     false,
			TLSConfig: nil,
			PoolSize:  10 * runtime.GOMAXPROCS(0),
		})

	if cfg.RateLimit.Enabled {
		app.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				// Skip rate limiting for health check and metrics endpoints
				if c.Path() == "/health" || c.Path() == "/metrics" || c.Path() == "/swagger/doc.json" {
					return true
				}
				return false
			},
			Max:        cfg.RateLimit.MaxRequests,
			Expiration: cfg.RateLimit.Expiration,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				// Handle rate limit exceeded response
				return c.Status(fiber.StatusTooManyRequests).JSON(dto.ErrorResponse{
					Error: "Rate limit exceeded. Please try again later.",
				})
			},
			Storage: storage,
		}))
	}

	healthHandler := health.NewHealthHandler(logger, db, redisClient)

	app.Get("/health", healthHandler.Check)
	app.Get("/metrics", middleware.BasicAuth(cfg), monitor.New(monitor.Config{Title: "MyService Metrics Page"}))
	app.Get("/swagger/*", middleware.BasicAuth(cfg), swagger.New(swagger.Config{
		URL: "/swagger/doc.json",
	}))

	redirectService := service.NewRedirectService(repository.New(db.DB), logger)
	redirectHandler := handler.NewRedirectHandler(redirectService, logger)

	app.Get("/:code", redirectHandler.RedirectToOriginalURL)

	api := app.Group("/api/v1")

	api.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Access-Control-Allow-Methods",
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight response for 5 minutes
	}))

	authMiddleware := middleware.NewAuthMiddleware(jwtMaker, logger)

	registerAuthHandlers(api, db, jwtMaker, logger, authMiddleware)
	registerAdminRoutes(api, db, authMiddleware, logger)
	registerUserRoutes(api, db, authMiddleware, logger)
}

// registerAuthHandlers sets up authentication routes
func registerAuthHandlers(router fiber.Router, db *database.Postgres, jwtMaker *token.JWTMaker, log *logger.Logger, authMiddleware *middleware.AuthMiddleware) {

	queries := repository.New(db.DB)
	authService := service.NewAuthService(queries, jwtMaker, log)
	authHandler := handler.NewAuthHandler(authService)

	router.Post("/login", authHandler.Login)
	router.Post("/register", authHandler.Register)

	router.Get("/profile", authMiddleware.Authenticate(), authHandler.GetProfile)
	router.Patch("/profile", authMiddleware.Authenticate(), authHandler.UpdateProfile)

	router.Put("/profile/password", authMiddleware.Authenticate(), authHandler.UpdatePassword)
	router.Delete("/logout", authMiddleware.Authenticate(), authHandler.Logout)

}

// registerUserRoutes sets up routes for authenticated users to manage their short links
func registerUserRoutes(router fiber.Router, db *database.Postgres, authMiddleware *middleware.AuthMiddleware, log *logger.Logger) {

	queries := repository.New(db.DB)
	shortLinkService := service.NewShortLinkService(queries, log)
	shortLinkHandler := handler.NewShortLinkHandler(shortLinkService, log)

	// User routes - all require authentication
	userRoutes := router.Group("/links")
	userRoutes.Use(authMiddleware.Authenticate())

	// Short link management routes
	userRoutes.Get("/", shortLinkHandler.GetUserLinks)
	userRoutes.Get("/:id", shortLinkHandler.GetUserLinkByID)
	userRoutes.Get("/code/:shortCode", shortLinkHandler.GetUserLinkByShortCode)
	userRoutes.Post("/", shortLinkHandler.CreateShortLink)
	userRoutes.Patch("/:id", shortLinkHandler.UpdateLink)
	userRoutes.Delete("/:id", shortLinkHandler.DeleteLink)
	userRoutes.Patch("/:id/status", shortLinkHandler.ToggleLinkStatus)

	// Bulk operations
	userRoutes.Post("/bulk", shortLinkHandler.CreateBulkShortLinks)
	userRoutes.Delete("/bulk", shortLinkHandler.DeleteBulkShortLinks)
	userRoutes.Delete("/", shortLinkHandler.DeleteAllLinks)

	shortLinkStatsService := service.NewShortLinksStatsService(queries, log)
	shortlinksStartshandler := handler.NewShortLinksStatsHandler(shortLinkStatsService, log)

	// Stats and utilities
	userRoutes.Get("/stats", shortlinksStartshandler.GetUserStats)
	//userRoutes.Get("/:id/stats", shortLinkHandler.GetLinkStats)
	//userRoutes.Get("/export", shortLinkHandler.ExportLinks)
	//userRoutes.Post("/import", shortLinkHandler.ImportLinks)
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

	//// Link management routes
	adminRoutes.Get("/links", adminHandler.ListAllLinks)
	adminRoutes.Get("/links/:id", adminHandler.GetLink)
	adminRoutes.Get("/users/:userId/links", adminHandler.ListUserLinks)
	adminRoutes.Patch("/links/:id/status", adminHandler.ToggleLinkStatus)
	adminRoutes.Get("/stats", adminHandler.GetSystemStats)

}
