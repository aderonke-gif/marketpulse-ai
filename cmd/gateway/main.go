package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/olamitunde/marketpulse-ai/internal/api"
	"github.com/olamitunde/marketpulse-ai/internal/cache"
	"github.com/olamitunde/marketpulse-ai/internal/client"
)

func main() {
	port := env("PORT", "8080")
	mlURL := env("ML_SERVICE_URL", "http://localhost:8000")
	ml := client.NewMLClient(mlURL, 3*time.Second)
	handler := api.New(ml, cache.New(10*time.Minute))
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("gateway listening", "port", port, "ml_service", mlURL)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
