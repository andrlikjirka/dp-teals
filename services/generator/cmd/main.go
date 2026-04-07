package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	ingestionv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/pkg/jws"
	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"github.com/andrlikjirka/dp-teals/services/generator/internal/generator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}

func run() error {
	log := logger.New("")
	count := flag.Int("count", 1, "number of events to generate")
	addr := flag.String("addr", "localhost:50051", "teals-server gRPC address")
	delayMs := flag.Int("delay", 10, "delay in milliseconds between event generation")
	privKeyB64 := flag.String("key", "", "base64-encoded Ed25519 private key for JWS signing")
	kid := flag.String("kid", "", "key ID (JWK thumbprint) matching the registered public key")
	flag.Parse()

	signer, err := buildSigner(*privKeyB64, *kid, log)
	if err != nil {
		log.Error("failed to create signer", "error", err)
		return err
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to create client for teals-server", "addr", *addr, "error", err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error("failed to close gRPC connection", "error", err)
		}
	}()

	if err := waitForReady(conn, 3*time.Second); err != nil {
		log.Error("failed to connect to teals-server", "addr", *addr, "error", err)
		return err
	}

	client := ingestionv1.NewIngestionServiceClient(conn)
	sender := generator.NewGrpcSender(client)
	eventSigner := generator.NewEventSigner(signer)
	gen := generator.NewGenerator(eventSigner, sender, log)

	if err = gen.Run(context.Background(), *count, *delayMs); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info("generator stopped early")
			return nil
		}
		log.Error("generator finished with errors", "error", err)
		return err
	}

	log.Info("done", "total", *count)
	return nil
}

func waitForReady(conn *grpc.ClientConn, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn.Connect()
	for {
		if conn.GetState() == connectivity.Ready {
			return nil
		}
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			return ctx.Err()
		}
	}
}

func buildSigner(privKeyB64, kid string, log *logger.Logger) (jws.Signer, error) {
	if privKeyB64 == "" && kid == "" {
		log.Info("no signing key provided — events will be sent unsigned")
		return nil, nil
	}
	if privKeyB64 == "" || kid == "" {
		return nil, fmt.Errorf("both -key and -kid must be provided together")
	}
	privBytes, err := base64.StdEncoding.DecodeString(privKeyB64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 key: %w", err)
	}
	signer, err := jws.NewEd25519Signer(privBytes, kid)
	if err != nil {
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}
	log.Info("JWS signing enabled", "kid", kid)
	return signer, nil
}
