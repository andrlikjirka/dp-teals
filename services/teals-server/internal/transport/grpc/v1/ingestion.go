package v1

import (
	"context"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/ingestion"
)

type IngestionServiceServer struct {
	ingestionv1.UnimplementedIngestionServiceServer
	service *ingestion.Service
}

func NewIngestionServiceServer(s *ingestion.Service) (*IngestionServiceServer, error) {
	return &IngestionServiceServer{
		service: s,
	}, nil
}

func (s *IngestionServiceServer) Append(ctx context.Context, req *ingestionv1.AppendRequest) (*ingestionv1.AppendResponse, error) {
	e, err := MapToAuditEvent(req)

	o, err := s.service.AppendEvent(ctx, e)
	if err != nil {
		return nil, err
	}

	return &ingestionv1.AppendResponse{EventId: o.String()}, nil
}
