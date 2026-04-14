package v1

import (
	"context"
	"errors"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryServiceServer implements the gRPC server for the QueryService defined in the protobuf. It provides methods to retrieve audit events by ID and to list audit events based on filters. The service layer is responsible for the actual retrieval logic, while the transport layer focuses on handling gRPC requests and responses.
type QueryServiceServer struct {
	auditv1.UnimplementedQueryServiceServer
	service *service.QueryService
}

// NewQueryServiceServer creates a new instance of QueryService with the provided QueryService. This allows the gRPC server to delegate the actual query logic to the service layer, keeping the transport layer focused on handling gRPC requests and responses.
func NewQueryServiceServer(s *service.QueryService) *QueryServiceServer {
	return &QueryServiceServer{
		service: s,
	}
}

// GetAuditEvent handles incoming GetAuditEventRequest messages, parses the event ID, and calls the service layer to retrieve the specified audit event. It returns a GetAuditEventResponse with the event details if successful, or an appropriate gRPC error status if the request is invalid or if there was an error during retrieval.
func (s *QueryServiceServer) GetAuditEvent(ctx context.Context, req *auditv1.GetAuditEventRequest) (*auditv1.GetAuditEventResponse, error) {
	id, err := uuid.Parse(req.GetEventId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	result, err := s.service.GetAuditEvent(ctx, id)
	if err != nil {
		if errors.Is(err, svcerrors.ErrAuditLogEntryNotFound) {
			return nil, status.Errorf(codes.NotFound, "audit event with ID %s not found", req.GetEventId())
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve audit event with ID %s: %v", req.GetEventId(), err)
	}

	return &auditv1.GetAuditEventResponse{
		Event:             mapToProtoAuditEvent(result.Event),
		LeafIndex:         result.LeafIndex,
		ProducerSignToken: result.SignatureToken,
	}, nil
}

// ListAuditEvents handles incoming ListAuditEventsRequest messages, retrieves a list of audit events based on the provided filter criteria using the service layer, and returns a ListAuditEventsResponse with the matching events and ledger size if successful. It returns an appropriate gRPC error status if there was an error during retrieval.
func (s *QueryServiceServer) ListAuditEvents(ctx context.Context, req *auditv1.ListAuditEventsRequest) (*auditv1.ListAuditEventsResponse, error) {
	filter := mapToAuditEventFilter(req.Filter)

	result, err := s.service.ListAuditEvents(ctx, &filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list audit events: %v", err)
	}

	items := make([]*auditv1.ListAuditEventsItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = &auditv1.ListAuditEventsItem{
			Event:             mapToProtoAuditEvent(item.Event),
			LeafIndex:         item.LeafIndex,
			ProducerSignToken: item.SignatureToken,
		}
	}

	return &auditv1.ListAuditEventsResponse{
		Items:      items,
		LedgerSize: result.LedgerSize,
	}, nil
}
