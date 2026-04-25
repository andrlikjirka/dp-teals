package generator

import (
	"context"
	"fmt"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/tools/generator/internal/model"
	"google.golang.org/grpc/metadata"
)

const signatureMetadataKey = "x-jws-event-signature"

type SendResult struct {
	EventID    string
	LedgerSize int64
	Timestamp  time.Time
}

// sender defines the interface for sending audit events to the ingestion service.
type sender interface {
	send(ctx context.Context, event *model.AuditEvent, token string) (*SendResult, error)
}

// GrpcSender implements the sender interface using gRPC to communicate with the ingestion service.
type GrpcSender struct {
	client auditv1.IngestionServiceClient
}

// NewGrpcSender creates a new instance of grpcSender with the provided gRPC client.
func NewGrpcSender(client auditv1.IngestionServiceClient) *GrpcSender {
	return &GrpcSender{client: client}
}

// send takes an auditEvent, converts it to the appropriate protobuf message, and sends it to the ingestion service using the gRPC client. It returns an error if the sending process fails.
func (s *GrpcSender) send(ctx context.Context, event *model.AuditEvent, token string) (*SendResult, error) {
	protoEvent, err := toProto(event)
	if err != nil {
		return nil, fmt.Errorf("error while mapping to protoEvent: %w", err)
	}

	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, signatureMetadataKey, token)
	}

	res, err := s.client.Append(ctx, &auditv1.AppendRequest{Event: protoEvent})
	if err != nil {
		return nil, fmt.Errorf("error while sending event: %w", err)
	}

	return &SendResult{
		EventID:    res.EventId,
		LedgerSize: res.LedgerSize,
		Timestamp:  res.AppendedAt.AsTime(),
	}, nil
}
