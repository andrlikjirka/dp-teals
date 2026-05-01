package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pkgjws "github.com/andrlikjirka/dp-teals/pkg/jws"
	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/bootstrap"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/protector"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/serializer"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/transport/grpc/v1"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/transport/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}

func run() error {
	// 1. Setup Server
	if err := bootstrap.LoadEnvFile(".env"); err != nil {
		fmt.Printf("env file error: %v\n", err)
		return err
	}
	config, err := bootstrap.LoadConfig()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return err
	}
	log := logger.New(config.Env)

	signer, err := bootstrap.NewServerSigner(config)
	if err != nil {
		log.Error("failed to setup server signer", "error", err)
		return err
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), config.DBConnectTimeout)
	defer dbCancel()
	pool, err := bootstrap.NewPgxPool(dbCtx, config.DatabaseURL)
	if err != nil {
		log.Error("Failed to create database pool", "error", err)
		return err
	}
	defer pool.Close()

	masterKEK, err := base64.StdEncoding.DecodeString(config.MasterKEKB64)
	if err != nil {
		log.Error("failed to decode master KEK", "error", err)
		return err
	}

	// Infrastructure
	jcsSerializer := serializer.NewJcsSerializer()
	txProvider := repository.NewTransactionProvider(pool)
	keyRepo := repository.NewProducerKeyRepository(pool)
	protect, err := protector.NewAesGcmProtector(masterKEK)
	if err != nil {
		log.Error("failed to create metadata protector", "error", err)
		return err
	}

	// Services
	verifier := pkgjws.NewEd25519Verifier(keyRepo)
	auditService := service.NewAuditService(txProvider, jcsSerializer, verifier, protect, log)
	keyService := service.NewKeyService(keyRepo, log)
	ledgerService := service.NewLedgerService(txProvider, log)
	checkpointService := service.NewCheckpointService(txProvider, signer, log)
	queryService := service.NewQueryService(txProvider, jcsSerializer, protect, log)
	subjectService := service.NewSubjectService(txProvider, log)

	// Transport
	cpWorker := worker.NewCheckpointWorker(checkpointService, config.CheckpointInterval, log)
	ingestor := v1.NewIngestionServiceServer(auditService)
	keys := v1.NewKeyRegistrationServiceServer(keyService)
	proofer := v1.NewProofServiceServer(ledgerService, checkpointService)
	querier := v1.NewQueryServiceServer(queryService)
	subjecter := v1.NewDataSubjectServiceServer(subjectService)

	server, err := bootstrap.NewServer(config, log, ingestor, keys, proofer, querier, subjecter)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		return err
	}

	// 2. Create an errgroup with context
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()
	g, ctx := errgroup.WithContext(appCtx)

	// 3. Run the gRPC Server
	g.Go(func() error {
		log.Info("Starting gRPC server", "port", config.Port)
		return server.Run()
	})

	g.Go(func() error {
		return cpWorker.Start(ctx)
	})

	// 4. Listen for shutdown signals in a separate goroutine
	g.Go(func() error {
		interceptSignals(ctx, log)

		appCancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer shutdownCancel()

		return server.Stop(shutdownCtx)
	})

	// 5. Wait for everything to finish
	if err := g.Wait(); err != nil {
		log.Error("Application terminated with error", "error", err)
		return err
	}
	return nil
}

func interceptSignals(ctx context.Context, log *logger.Logger) {
	sigc := make(chan os.Signal, 1)
	defer signal.Stop(sigc)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigc:
		log.Info("Shutdown signal received", "signal", sig.String())
	case <-ctx.Done():
		log.Info("Context cancelled, stopping signal listener")
	}
}
