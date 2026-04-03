package generator

import (
	"fmt"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/generator/internal/model"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProto(event *model.AuditEvent) (*ingestionv1.AuditEvent, error) {
	var env *ingestionv1.Environment
	if event.Environment != nil {
		env = &ingestionv1.Environment{
			Service: event.Environment.Service,
			TraceId: event.Environment.TraceID,
			SpanId:  event.Environment.SpanID,
		}
	}

	var metadata *structpb.Struct
	if event.Metadata != nil {
		var err error
		metadata, err = structpb.NewStruct(event.Metadata)
		if err != nil {
			return nil, fmt.Errorf("toProto: convert metadata: %w", err)
		}
	}

	return &ingestionv1.AuditEvent{
		Id:          event.ID,
		Timestamp:   timestamppb.New(event.Timestamp),
		Environment: env,
		Actor:       &ingestionv1.Actor{Type: event.Actor.Type, Id: event.Actor.ID},
		Subject:     &ingestionv1.Subject{Id: event.Subject.ID},
		Action:      event.Action,
		Resource:    &ingestionv1.Resource{Id: event.Resource.ID, Name: event.Resource.Name, Fields: event.Resource.Fields},
		Result:      &ingestionv1.Result{Status: event.Result.Status, Reason: event.Result.Reason},
		Metadata:    metadata,
	}, nil

}
