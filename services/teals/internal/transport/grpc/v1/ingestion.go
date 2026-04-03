package v1

import (
	"context"
	"errors"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals/internal/service"
	svcerrors "github.com/andrlijirka/dp-teals/services/teals/internal/service/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IngestionServiceServer implements the gRPC server for the IngestionService defined in the protobuf.
type IngestionServiceServer struct {
	ingestionv1.UnimplementedIngestionServiceServer
	service *service.AuditService
}

// NewIngestionServiceServer creates a new instance of IngestionServiceServer with the provided AuditService. This allows the gRPC server to delegate the actual ingestion logic to the service layer, keeping the transport layer focused on handling gRPC requests and responses.
func NewIngestionServiceServer(s *service.AuditService) *IngestionServiceServer {
	return &IngestionServiceServer{
		service: s,
	}
}

// Append handles incoming AppendRequest messages, maps them to the internal audit event model, and calls the service layer to ingest the event. It returns an AppendResponse with the event ID if successful, or an appropriate gRPC error status if the request is invalid or if there was an error during ingestion.
func (s *IngestionServiceServer) Append(ctx context.Context, req *ingestionv1.AppendRequest) (*ingestionv1.AppendResponse, error) {
	e, err := MapToAuditEvent(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	o, err := s.service.IngestAuditEvent(ctx, e)
	if err != nil {
		if errors.Is(err, svcerrors.ErrDuplicateEventID) {
			return nil, status.Errorf(codes.AlreadyExists, "audit event with ID %s already exists", e.ID)
		}
		return nil, status.Errorf(codes.Internal, "failed to append the audit event with ID %s", e.ID)
	}

	return &ingestionv1.AppendResponse{EventId: o.String()}, nil
}
