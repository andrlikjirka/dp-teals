package worker

import (
	"context"
	"time"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service"
)

// CheckpointWorker is responsible for periodically creating signed checkpoints of the ledger state. It runs in a separate goroutine and uses a ticker to trigger checkpoint creation at regular intervals. The worker handles errors gracefully, logging any issues encountered during checkpoint creation without crashing the application.
type CheckpointWorker struct {
	service  *service.CheckpointService
	interval time.Duration
	logger   *logger.Logger
}

// NewCheckpointWorker creates a new instance of CheckpointWorker with the provided CheckpointService, interval for checkpoint creation, and Logger. This worker will periodically create signed checkpoints of the ledger state at the specified interval, logging any errors encountered during the process.
func NewCheckpointWorker(service *service.CheckpointService, interval time.Duration, logger *logger.Logger) *CheckpointWorker {
	return &CheckpointWorker{
		service:  service,
		interval: interval,
		logger:   logger,
	}
}

// Start begins the checkpoint worker loop, which creates checkpoints at regular intervals until the provided context is canceled. It uses a ticker to trigger checkpoint creation and listens for context cancellation to gracefully stop the worker when needed.
func (w *CheckpointWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, _ = w.service.CreateCheckpoint(ctx)
		case <-ctx.Done():
			w.logger.Info("checkpoint worker stopped")
			return nil
		}
	}
}
