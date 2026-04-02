package v1

import (
	"context"
	"errors"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service"
	svcerrors "github.com/andrlijirka/dp-teals/services/teals-server/internal/service/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IngestionServiceServer struct {
	ingestionv1.UnimplementedIngestionServiceServer
	service *service.AuditService
}

func NewIngestionServiceServer(s *service.AuditService) (*IngestionServiceServer, error) {
	return &IngestionServiceServer{
		service: s,
	}, nil
}

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
