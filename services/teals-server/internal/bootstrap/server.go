package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/pkg/logger"
	v1 "github.com/andrlijirka/dp-teals/services/teals-server/internal/transport/api/grpc/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server encapsulates the HTTP server and its dependencies.
type Server struct {
	grpcSrv      *grpc.Server
	listener     net.Listener
	logger       *slog.Logger
	healthServer *health.Server
	config       Config
}

// NewServer creates a new Server instance with the given configuration
func NewServer(cfg Config, log *logger.Logger, ingestor *v1.IngestionServiceServer) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Port, err)
	}

	grpcSrv := grpc.NewServer()
	ingestionv1.RegisterIngestionServiceServer(grpcSrv, ingestor)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthServer)
	// go checkDatabaseHealth -> 1 sec interval -> halth server status update

	if cfg.EnableReflection {
		reflection.Register(grpcSrv)
	}

	return &Server{
		grpcSrv:      grpcSrv,
		listener:     listener,
		config:       cfg,
		logger:       log.Logger,
		healthServer: healthServer,
	}, nil
}

// Run starts the HTTP server and listens for incoming requests.
func (s *Server) Run() error {
	if s.listener == nil {
		return fmt.Errorf("server listener is not initialized")
	}
	s.logger.Info("Server listening", slog.Int("port", s.config.Port))

	if err := s.grpcSrv.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server, allowing ongoing requests to complete.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Initiating graceful shutdown...")

	s.healthServer.Shutdown()

	// Create a channel to signal when GracefulStop is done
	done := make(chan struct{})
	go func() {
		s.grpcSrv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Server stopped gracefully.")
		return nil
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout reached, forcing stop.")
		s.grpcSrv.Stop() // Force close connections
		return ctx.Err()
	}
}
