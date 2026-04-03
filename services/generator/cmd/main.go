package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	ingestionv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"github.com/andrlikjirka/dp-teals/services/generator/internal/generator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logger.New("")

	count := flag.Int("count", 1, "number of events to generate")
	addr := flag.String("addr", "localhost:50051", "teals-server gRPC address")
	delayMs := flag.Int("delay", 10, "delay in milliseconds between event generation")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to create client for teals-server", "addr", *addr, "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error("failed to close gRPC connection", "error", err)
		}
	}()

	if err := waitForReady(conn, 3*time.Second); err != nil {
		log.Error("failed to connect to teals-server", "addr", *addr, "error", err)
		os.Exit(1)
	}

	client := ingestionv1.NewIngestionServiceClient(conn)
	sender := generator.NewGrpcSender(client)
	gen := generator.NewGenerator(sender, log)
	ctx := context.Background()

	if err = gen.Run(ctx, *count, *delayMs); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info("generator stopped early")
			os.Exit(0)
		}
		log.Error("generator finished with errors", "error", err)
		os.Exit(1)
	}
	log.Info("done", "total", *count)
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
