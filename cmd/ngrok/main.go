package main

import (
	"GoShort/internal/server"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

func main() {
	app := server.InitApp()
	defer server.Cleanup(app)

	// Start app
	go server.StartServer(app)

	// Start ngrok tunnel
	go setupNgrok(app)

	// Wait for interrupt signal
	server.WaitForShutdown(app)
}

func setupNgrok(app *server.App) {
	ctx := context.Background()

	// Get port from app config or use default 8080
	port := 8080
	if app.Config != nil && app.Config.Server.Port != "" {
		if p, err := strconv.Atoi(app.Config.Server.Port); err == nil {
			port = p
		}
	}

	// Create the ngrok tunnel
	listener, err := ngrok.Listen(ctx,
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		log.Printf("Failed to start ngrok tunnel: %v", err)
		return
	}

	log.Println("Ingress established at:", listener.URL())

	// Create a proxy handler that forwards requests to local app
	proxyHandler := func(w http.ResponseWriter, r *http.Request) {
		localURL := fmt.Sprintf("http://localhost:%d%s", port, r.URL.Path)

		// Forward request to local server
		resp, err := http.Get(localURL)
		if err != nil {
			http.Error(w, "Error forwarding request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy headers and status code
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		_, _ = io.Copy(w, http.MaxBytesReader(w, resp.Body, 1<<20))
	}

	if err := http.Serve(listener, http.HandlerFunc(proxyHandler)); err != nil {
		log.Printf("Ngrok serving failed: %v", err)
	}
}
