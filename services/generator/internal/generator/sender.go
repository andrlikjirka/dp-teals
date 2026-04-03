package generator

import (
	"context"
	"fmt"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/generator/internal/model"
)

// sender defines the interface for sending audit events to the ingestion service.
type sender interface {
	send(ctx context.Context, event *model.AuditEvent) error
}

// GrpcSender implements the sender interface using gRPC to communicate with the ingestion service.
type GrpcSender struct {
	client ingestionv1.IngestionServiceClient
}

// NewGrpcSender creates a new instance of grpcSender with the provided gRPC client.
func NewGrpcSender(client ingestionv1.IngestionServiceClient) *GrpcSender {
	return &GrpcSender{client: client}
}

// send takes an auditEvent, converts it to the appropriate protobuf message, and sends it to the ingestion service using the gRPC client. It returns an error if the sending process fails.
func (s *GrpcSender) send(ctx context.Context, event *model.AuditEvent) error {
	proto, err := toProto(event)
	if err != nil {
		return fmt.Errorf("error while mapping to proto: %w", err)
	}

	_, err = s.client.Append(ctx, &ingestionv1.AppendRequest{Event: proto})
	if err != nil {
		return fmt.Errorf("error while sending event: %w", err)
	}

	return nil
}
