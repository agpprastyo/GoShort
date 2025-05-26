package main

func main() {
	app := initApp()
	defer cleanup(app)

	// Start server
	go startServer(app)

	// Wait for interrupt signal
	waitForShutdown(app)
}
