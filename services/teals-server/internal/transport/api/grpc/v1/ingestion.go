package v1

import (
	"context"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service"
)

type IngestionServiceServer struct {
	ingestionv1.UnimplementedIngestionServiceServer
	service *service.Service
}

func NewIngestionServiceServer(s *service.Service) (*IngestionServiceServer, error) {
	return &IngestionServiceServer{
		service: s,
	}, nil
}

func (s *IngestionServiceServer) Append(ctx context.Context, req *ingestionv1.AppendRequest) (*ingestionv1.AppendResponse, error) {
	o, err := s.service.AppendEvent(ctx, toAppendEventInput(req))
	if err != nil {
		return nil, err
	}

	return toAppendResponse(o), nil
}
