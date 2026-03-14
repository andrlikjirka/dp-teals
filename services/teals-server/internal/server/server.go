package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/andrlikjirka/logger"
)

// Server encapsulates the HTTP server and its dependencies.
type Server struct {
	server *http.Server
	logger *slog.Logger
	config Config
	wg     sync.WaitGroup // to wait for background tasks to finish
}

// New creates a new Server instance with the given configuration
func New(cfg Config, log *logger.Logger, handler http.Handler) *Server {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		server: httpServer,
		logger: log.Logger,
		config: cfg,
	}
}

// Run starts the HTTP server and listens for incoming requests.
func (s *Server) Run() error {
	s.logger.Info("Server listening", slog.Int("port", s.config.Port))

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server crashed: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server, allowing ongoing requests to complete.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Initiating graceful shutdown...")

	// 1. Stop accepting new requests and finish ongoing ones
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.logger.Info("Server stopped. Waiting for background tasks...")

	// 2. Wait for any background goroutines to complete
	s.wg.Wait()

	s.logger.Info("All background tasks completed. Server fully stopped.")
	return nil
}
