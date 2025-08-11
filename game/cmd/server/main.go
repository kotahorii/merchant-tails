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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/merchant-tails/game/internal/infrastructure/logging"
	"github.com/yourusername/merchant-tails/game/internal/infrastructure/monitoring"
	"github.com/yourusername/merchant-tails/game/internal/infrastructure/server"
)

var (
	port        = flag.String("port", "8080", "Server port")
	metricsPort = flag.String("metrics-port", "9090", "Metrics port for Prometheus")
	saveDir     = flag.String("save-dir", "./saves", "Save directory")
	logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	env         = flag.String("env", "development", "Environment (development, production)")
)

func main() {
	flag.Parse()

	// Initialize structured logging
	logConfig := &logging.LoggerConfig{
		Level:      parseLogLevel(*logLevel),
		Console:    true,
		JSON:       *env == "production",
		TimeFormat: time.RFC3339,
		Context: map[string]interface{}{
			"environment": *env,
			"service":     "merchant-tails-server",
		},
	}

	logManagerConfig := &logging.LogManagerConfig{
		LogDir:          "./logs",
		MaxFileSize:     100 * 1024 * 1024, // 100MB
		MaxBackups:      10,
		MaxAge:          30,
		Compress:        true,
		BufferSize:      1000,
		FlushInterval:   time.Second,
		RotationTime:    24 * time.Hour,
		FileNamePattern: "merchant-tails-%s.log",
	}

	if err := logging.Initialize(logConfig, logManagerConfig); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logging.Close()

	logging.Info("Starting Merchant Tails server")
	logging.WithFields(map[string]interface{}{
		"port":         *port,
		"metrics_port": *metricsPort,
		"environment":  *env,
	}).Info("Server configuration")

	// Create save directory if it doesn't exist
	if err := os.MkdirAll(*saveDir, 0755); err != nil {
		log.Fatalf("Failed to create save directory: %v", err)
	}

	// Initialize metrics collector
	metricsCollector := monitoring.NewMetricsCollector()

	// Start metrics server
	metricsPortInt := 9090
	if _, err := fmt.Sscanf(*metricsPort, "%d", &metricsPortInt); err != nil {
		log.Printf("Invalid metrics port, using default 9090: %v", err)
	}
	if err := metricsCollector.StartServer(metricsPortInt); err != nil {
		logging.WithError(err).Error("Failed to start metrics server")
	}
	logging.Infof("Metrics server started on port %d", metricsPortInt)

	// Start runtime metrics collector
	runtimeCollector := monitoring.NewRuntimeCollector(metricsCollector)
	runtimeCollector.Start(10 * time.Second)

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
		logging.Infof("Starting Merchant Tails server on %s (env: %s)", addr, *env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.WithError(err).Error("Server forced to shutdown")
	}

	// Stop metrics collector
	if metricsCollector != nil {
		if err := metricsCollector.StopServer(); err != nil {
			logging.WithError(err).Error("Error stopping metrics server")
		}
	}

	logging.Info("Server exited")
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
	mux.Handle("/metrics", promhttp.Handler())

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

func parseLogLevel(level string) logging.LogLevel {
	switch level {
	case "debug":
		return logging.DebugLevel
	case "info":
		return logging.InfoLevel
	case "warn":
		return logging.WarnLevel
	case "error":
		return logging.ErrorLevel
	default:
		return logging.InfoLevel
	}
}
