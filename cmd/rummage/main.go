// Package main provides the entry point for the Rummage application.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ncecere/rummage/pkg/api"
	"github.com/ncecere/rummage/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the API router
	router, err := api.NewRouter(api.RouterOptions{
		BaseURL:  cfg.BaseURL,
		RedisURL: cfg.RedisURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	// Configure the server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		log.Printf("API base URL: %s", cfg.BaseURL)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown or server errors.
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Server is shutting down... (Signal: %v)", sig)

		// Create a deadline to wait for.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
	}

	fmt.Println("Server stopped")
}
