package main

import (
	_ "GoShort/docs"
	"GoShort/internal/server"
)

// @title GoShort API
// @version 1.0
// @description A URL shortener service API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email your-email@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in cookie
// @name access_token
func main() {
	app := server.InitApp()
	defer server.Cleanup(app)

	// Start app
	go server.StartServer(app)

	// Wait for interrupt signal
	server.WaitForShutdown(app)
}
