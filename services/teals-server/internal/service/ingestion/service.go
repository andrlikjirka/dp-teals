package ingestion

import (
	"context"
	"fmt"

	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/ingestion/model"
	"github.com/google/uuid"
)

type Service struct {
}

func NewIngestionService() *Service {
	return &Service{}
}

func (*Service) AppendEvent(ctx context.Context, event *model.AuditEvent) (uuid.UUID, error) {
	fmt.Printf("Appending event with ID: %s\n", event.Id)

	return event.Id, nil
}
