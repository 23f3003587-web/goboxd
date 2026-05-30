package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/handler"
)

func main() {
	// 1. Initialize Configuration
	if err := config.LoadLanguages("languages.yaml"); err != nil {
		log.Fatalf("❌ Failed to load languages: %v", err)
	}
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// 2. Setup Router & Middleware
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 3. Define Routes
	// Standardized naming convention; removed redundant wildcards/typo paths.
	r.Get("/health", handler.Healthz)
	r.Get("/healthz", handler.Healthz)
	r.Get("/ready", handler.Readyz)
	r.Get("/readyz", handler.Readyz)
	r.Get("/info", handler.Info)

	r.Post("/run", handler.Run)

	// 4. Configure HTTP Server with Timeouts
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Setup Graceful Shutdown Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 6. Start Server in a Goroutine
	go func() {
		log.Printf("✅ goboxd started on :8080 | Languages: %d\n", len(config.Languages))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Block until a shutdown signal is received
	<-ctx.Done()

	log.Println("⚠️ Shutting down goboxd server gracefully...")

	// 7. Enforce Maximum Shutdown Timeout Window
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("🛑 Server stopped cleanly.")
}
