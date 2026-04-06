package generator

import (
	"context"
	"fmt"

	ingestionv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/pkg/jws"
	"github.com/andrlikjirka/dp-teals/services/generator/internal/model"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const signatureMetadataKey = "x-jws-event-signature"

// sender defines the interface for sending audit events to the ingestion service.
type sender interface {
	send(ctx context.Context, event *model.AuditEvent) error
}

// GrpcSender implements the sender interface using gRPC to communicate with the ingestion service.
type GrpcSender struct {
	client ingestionv1.IngestionServiceClient
	signer jws.Signer
}

// NewGrpcSender creates a new instance of grpcSender with the provided gRPC client.
func NewGrpcSender(client ingestionv1.IngestionServiceClient, signer jws.Signer) *GrpcSender {
	return &GrpcSender{client: client, signer: signer}
}

// send takes an auditEvent, converts it to the appropriate protobuf message, and sends it to the ingestion service using the gRPC client. It returns an error if the sending process fails.
func (s *GrpcSender) send(ctx context.Context, event *model.AuditEvent) error {
	protoEvent, err := toProto(event)
	if err != nil {
		return fmt.Errorf("error while mapping to protoEvent: %w", err)
	}

	if s.signer != nil {
		payload, err := proto.MarshalOptions{Deterministic: true}.Marshal(protoEvent)
		if err != nil {
			return fmt.Errorf("error while marshaling event for signing: %w", err)
		}
		token, err := s.signer.Sign(payload)
		if err != nil {
			return fmt.Errorf("error while signing event: %w", err)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, signatureMetadataKey, token)
	}

	_, err = s.client.Append(ctx, &ingestionv1.AppendRequest{Event: protoEvent})
	if err != nil {
		return fmt.Errorf("error while sending event: %w", err)
	}

	return nil
}
