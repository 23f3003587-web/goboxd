package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thesouldev/goboxd/internal/api"
	"github.com/thesouldev/goboxd/internal/api/handlers"
	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/runner"
)

// Service is the core application entry point for the sandbox service.
type Service struct {
	Addr string
}

func NewService() *Service {
	addr := os.Getenv("GOBOXD_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return &Service{Addr: addr}
}

func (s *Service) Run() error {
	if err := config.LoadLanguages("config/languages.yaml"); err != nil {
		return fmt.Errorf("load languages: %w", err)
	}
	if err := config.LoadConfig(); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	handlers.InitializeConcurrency()
	registry := runner.NewRegistry()
	if registry == nil {
		return fmt.Errorf("runner registry initialization failed")
	}

	router := api.NewRouter()
	srv := &http.Server{
		Addr:         s.Addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("✅ goboxd started on %s | Languages: %d\n", s.Addr, len(config.Languages))
		log.Printf("✅ runner registry initialized with %d supported language entries\n", len(registry.List()))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server failed to start: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		log.Println("⚠️ Shutting down goboxd server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
		log.Println("🛑 Server stopped cleanly.")
	}

	return nil
}
