package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrlijirka/dp-teals/pkg/logger"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/api/grpc/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/application/ingestion"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/server"
	"golang.org/x/sync/errgroup"
)

func main() {
	// 1. Setup Server
	config := server.MustLoadConfig(".env")
	log := logger.New(config.Env)

	ingestionService := ingestion.NewIngestionService()
	ingestor, err := v1.NewIngestionServiceServer(ingestionService)
	if err != nil {
		log.Error("Failed to create gRPC service", "error", err)
		os.Exit(1)
	}

	server, err := server.NewServer(config, log, ingestor)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// 2. Create an errgroup with context
	// The ctx is canceled if any function in the group returns an error
	// or if we cancel it manually via signal.
	g, ctx := errgroup.WithContext(context.Background())

	// 3. Run the gRPC Server
	g.Go(func() error {
		log.Info("Starting gRPC server", "port", config.Port)
		return server.Run()
	})

	// 4. Listen for shutdown signals in a separate goroutine
	g.Go(func() error {
		// This will block until a shutdown signal is received or the context is canceled
		interceptSignals(ctx, log)

		return server.Stop(ctx)
	})

	// 5. Wait for everything to finish
	if err := g.Wait(); err != nil {
		log.Error("Application terminated with error", "error", err)
		os.Exit(1)
	}
}

func interceptSignals(ctx context.Context, log *logger.Logger) {
	sigc := make(chan os.Signal, 1)

	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case sig := <-sigc:
		log.Info("Shutdown signal received", "signal", sig.String())
	case <-ctx.Done():
		log.Info("Context cancelled, stopping signal listener")
	}
}
