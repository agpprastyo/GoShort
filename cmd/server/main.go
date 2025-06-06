package main

import _ "GoShort/docs"

// @title GoShort API
// @version 1.0
// @description A URL shortener service API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email your-email@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
func main() {
	app := initApp()
	defer cleanup(app)

	// Start server
	go startServer(app)

	// Wait for interrupt signal
	waitForShutdown(app)
}
