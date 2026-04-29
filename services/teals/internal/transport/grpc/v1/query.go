package v1

import (
	"context"
	"encoding/json"
	"errors"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/transport/grpc/v1/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
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

	event, err := eventPayloadToStruct(result.Payload)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert event payload to struct for event ID %s: %v", req.GetEventId(), err)
	}

	revealed, err := revealedMetadataToStruct(result.RevealedMetadata)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert revealed metadata for event ID %s: %v", req.GetEventId(), err)
	}

	return &auditv1.GetAuditEventResponse{
		Event:             event,
		RevealedMetadata:  revealed,
		LeafIndex:         result.LeafIndex,
		ProducerSignToken: result.SignatureToken,
	}, nil
}

// ListAuditEvents handles incoming ListAuditEventsRequest messages, retrieves a list of audit events based on the provided filter criteria using the service layer, and returns a ListAuditEventsResponse with the matching events and ledger size if successful. It returns an appropriate gRPC error status if there was an error during retrieval.
func (s *QueryServiceServer) ListAuditEvents(ctx context.Context, req *auditv1.ListAuditEventsRequest) (*auditv1.ListAuditEventsResponse, error) {
	filter := model.MapToAuditEventFilter(req.Filter)

	var cursor *int64
	if req.Cursor != nil {
		decoded, err := model.DecodeCursor(*req.Cursor)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid cursor: %v", err)
		}
		cursor = &decoded
	}

	result, err := s.service.ListAuditEvents(ctx, &filter, cursor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list audit events: %v", err)
	}

	items := make([]*auditv1.ListAuditEventsItem, len(result.Items))
	for i, item := range result.Items {
		event, err := eventPayloadToStruct(item.Payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert event payload to struct for event ID %s: %v", item.Event.ID, err)
		}

		revealed, err := revealedMetadataToStruct(item.RevealedMetadata)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert revealed metadata for event ID %s: %v", item.Event.ID, err)
		}

		items[i] = &auditv1.ListAuditEventsItem{
			Event:             event,
			RevealedMetadata:  revealed,
			LeafIndex:         item.LeafIndex,
			ProducerSignToken: item.SignatureToken,
		}
	}

	resp := &auditv1.ListAuditEventsResponse{
		Items:      items,
		LedgerSize: result.LedgerSize,
	}
	if result.NextCursor != nil {
		next := model.EncodeCursor(*result.NextCursor)
		resp.NextCursor = &next
	}

	return resp, nil
}

// eventPayloadToStruct converts a JSON payload from the audit event into a protobuf Struct, which can be used in the gRPC response. It returns an error if the payload cannot be unmarshaled into a Struct.
func eventPayloadToStruct(payload json.RawMessage) (*structpb.Struct, error) {
	out := &structpb.Struct{}
	if err := out.UnmarshalJSON(payload); err != nil {
		return nil, err
	}
	return out, nil
}

// mapToAuditEventFilter converts a gRPC request filter into the internal model used by the service layer to query audit events. This function maps the fields from the gRPC request to the corresponding fields in the service's filter model.
func revealedMetadataToStruct(m map[string]any) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}
	return structpb.NewStruct(m)
}
