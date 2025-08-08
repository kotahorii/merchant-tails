package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/merchant-tails/game/internal/infrastructure/server"
)

var (
	port     = flag.String("port", "8080", "Server port")
	saveDir  = flag.String("save-dir", "./saves", "Save directory")
	logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	env      = flag.String("env", "development", "Environment (development, production)")
)

func main() {
	flag.Parse()

	// Initialize logger
	initLogger(*logLevel)

	// Create save directory if it doesn't exist
	if err := os.MkdirAll(*saveDir, 0755); err != nil {
		log.Fatalf("Failed to create save directory: %v", err)
	}

	// Create game server
	gameServer := server.NewGameServer(*saveDir)

	// Create HTTP server with Connect-RPC handler
	addr := fmt.Sprintf(":%s", *port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      createHandler(gameServer),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Merchant Tails server on %s (env: %s)", addr, *env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func createHandler(gameServer *server.GameServer) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Metrics endpoint (for Prometheus)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement Prometheus metrics
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "# HELP merchant_tails_up Is the server up")
		fmt.Fprintln(w, "# TYPE merchant_tails_up gauge")
		fmt.Fprintln(w, "merchant_tails_up 1")
	})

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"version":"1.0.0","api_version":"v1","build_time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Add CORS middleware for development
	if *env == "development" {
		return corsMiddleware(mux)
	}

	return mux
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func initLogger(level string) {
	// TODO: Implement proper logging with levels
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Logger initialized with level: %s", level)
}
