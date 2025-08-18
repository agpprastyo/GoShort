package server

import (
	_ "GoShort/docs"
	"GoShort/internal/admin"
	"GoShort/internal/auth"
	"GoShort/internal/commons"
	"GoShort/internal/datastore"
	"GoShort/internal/health"
	"GoShort/internal/middleware"
	"GoShort/internal/redirect"
	"GoShort/internal/shortlink"
	"GoShort/internal/stats"

	mail2 "GoShort/pkg/mail"
	"runtime"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	redisFiber "github.com/gofiber/storage/redis/v3"
	"github.com/gofiber/swagger"
)

// SetupRoutes registers all application routes

func SetupRoutes(app *App) {
	app.FiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     "https://goshort.agprastyo.me, https://goshort-api.agprastyo.me, http://localhost:5173, http://localhost:3000, http://127.0.0.1:5173",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Access-Control-Allow-Methods",
		AllowCredentials: true,
		MaxAge:           300,
	}))

	port, _ := strconv.Atoi(app.Config.Redis.Port)

	storage := redisFiber.New(
		redisFiber.Config{
			Host:      app.Config.Redis.Host,
			Port:      port,
			Password:  app.Config.Redis.Password,
			Database:  app.Config.Redis.DB,
			Reset:     false,
			TLSConfig: nil,
			PoolSize:  10 * runtime.GOMAXPROCS(0),
		})

	if app.Config.RateLimit.Enabled {
		app.FiberApp.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				if c.Path() == "/health" || c.Path() == "/metrics" || c.Path() == "/swagger/doc.json" {
					return true
				}
				return false
			},
			Max:        app.Config.RateLimit.MaxRequests,
			Expiration: app.Config.RateLimit.Expiration,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(commons.ErrorResponse{
					Error: "Rate limit exceeded. Please try again later.",
				})
			},
			Storage: storage,
		}))
	}

	healthHandler := health.NewHealthHandler(app.Logger, app.DB, app.Redis)

	app.FiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to GoShort API! ")
	})
	app.FiberApp.Get("/health", healthHandler.Check)
	app.FiberApp.Get("/metrics", middleware.BasicAuth(app.Config), monitor.New(monitor.Config{Title: "MyService Metrics Page"}))
	app.FiberApp.Get("/swagger/*", middleware.BasicAuth(app.Config), swagger.New(swagger.Config{
		URL: "/swagger/doc.json",
	}))

	redirectService := redirect.NewService(datastore.New(app.DB.DB), app.Logger)
	redirectHandler := redirect.NewRedirectHandler(redirectService, app.Logger)

	api := app.FiberApp.Group("/api/v1")

	registerAuthHandlers(api, app)
	registerAdminRoutes(api, app)
	registerUserRoutes(api, app)

	app.FiberApp.Get("/:code", redirectHandler.RedirectToOriginalURL)
}

// registerAuthHandlers sets up authentication routes
func registerAuthHandlers(router fiber.Router, app *App) {

	mailService := mail2.NewSendGridService(app.Config, app.Logger)
	authService := auth.NewService(app.Querier, app.JWTMaker, app.Logger, mailService)
	authHandler := auth.NewHandler(authService)

	authMiddleware := middleware.NewAuthMiddleware(app.JWTMaker, app.Logger)

	router.Post("/login", authHandler.Login)
	router.Post("/register", authHandler.Register)

	router.Get("/profile", authMiddleware.Authenticate(), authHandler.GetProfile)
	router.Patch("/profile", authMiddleware.Authenticate(), authHandler.UpdateProfile)

	router.Put("/profile/password", authMiddleware.Authenticate(), authHandler.UpdatePassword)
	router.Delete("/logout", authMiddleware.Authenticate(), authHandler.Logout)

}

// registerUserRoutes sets up routes for authenticated users to manage their short links
func registerUserRoutes(router fiber.Router, app *App) {

	queries := datastore.New(app.DB.DB)
	shortLinkService := shortlink.NewService(queries, app.Logger)
	shortLinkHandler := shortlink.NewShortLinkHandler(shortLinkService, app.Logger)

	authMiddleware := middleware.NewAuthMiddleware(app.JWTMaker, app.Logger)

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

	shortLinkStatsService := stats.NewShortLinksStatsService(queries, app.Logger)
	shortLinksStatsHandler := stats.NewShortLinksStatsHandler(shortLinkStatsService, app.Logger)

	// Stats and utilities
	userRoutes.Get("/stats", shortLinksStatsHandler.GetUserStats)
	//userRoutes.Get("/:id/stats", shortLinkHandler.GetLinkStats)
	//userRoutes.Get("/export", shortLinkHandler.ExportLinks)
	//userRoutes.Post("/import", shortLinkHandler.ImportLinks)
}

// registerAdminRoutes sets up routes for admin users to manage the application
func registerAdminRoutes(router fiber.Router, app *App) {
	queries := datastore.New(app.DB.DB)
	adminService := admin.NewService(queries, app.Logger)
	adminHandler := admin.NewHandler(adminService, app.Logger, app.validator)

	authMiddleware := middleware.NewAuthMiddleware(app.JWTMaker, app.Logger)

	roleMiddleware := middleware.NewRoleMiddleware()

	adminRoutes := router.Group("/admin")
	adminRoutes.Use(authMiddleware.Authenticate(), roleMiddleware.RequireAdmin())

	adminRoutes.Get("/links", adminHandler.ListAllLinks)
	adminRoutes.Get("/links/:id", adminHandler.GetLink)
	adminRoutes.Get("/users/:userId/links", adminHandler.ListUserLinks)
	adminRoutes.Patch("/links/:id/status", adminHandler.ToggleLinkStatus)
	adminRoutes.Get("/stats", adminHandler.GetSystemStats)
}
