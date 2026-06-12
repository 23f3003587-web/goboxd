package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thesouldev/goboxd/internal/api/handlers"
)

// NewRouter creates the HTTP router used by the service.
func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Core middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// RealIP is disabled due to security deprecation (IP spoofing risk)
	// Use only if you have proper proxy configuration in production
	// r.Use(middleware.RealIP)

	// Custom recovery middleware
	r.Use(handlers.RecoveryMiddleware)

	// Health and info endpoints
	r.Get("/health", handlers.Healthz)
	r.Get("/healthz", handlers.Healthz)
	r.Get("/ready", handlers.Readyz)
	r.Get("/readyz", handlers.Readyz)
	r.Get("/info", handlers.Info)

	// Main execution endpoint
	r.Post("/run", handlers.Run)

	return r
}
