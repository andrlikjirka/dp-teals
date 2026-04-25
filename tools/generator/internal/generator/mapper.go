package generator

import (
	"fmt"

	ingestionv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/tools/generator/internal/model"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProto(event *model.AuditEvent) (*ingestionv1.AuditEvent, error) {
	actorType, err := toProtoActorType(event.Actor.Type)
	if err != nil {
		return nil, err
	}
	action, err := toProtoAction(event.Action)
	if err != nil {
		return nil, err
	}
	resultStatus, err := toProtoResultStatus(event.Result.Status)
	if err != nil {
		return nil, err
	}

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
		metadata, err = structpb.NewStruct(event.Metadata)
		if err != nil {
			return nil, fmt.Errorf("toProto: convert metadata: %w", err)
		}
	}

	return &ingestionv1.AuditEvent{
		Id:          event.ID.String(),
		Timestamp:   timestamppb.New(event.Timestamp),
		Environment: env,
		Actor:       &ingestionv1.Actor{Type: actorType, Id: event.Actor.ID},
		Subject:     &ingestionv1.Subject{Id: event.Subject.ID},
		Action:      action,
		Resource:    &ingestionv1.Resource{Id: event.Resource.ID, Name: event.Resource.Name, Fields: event.Resource.Fields},
		Result:      &ingestionv1.Result{Status: resultStatus, Reason: event.Result.Reason},
		Metadata:    metadata,
	}, nil
}

func toProtoActorType(t model.ActorType) (ingestionv1.Actor_Type, error) {
	switch t {
	case model.ActorTypeUser:
		return ingestionv1.Actor_TYPE_USER, nil
	case model.ActorTypeSystem:
		return ingestionv1.Actor_TYPE_SYSTEM, nil
	default:
		return 0, fmt.Errorf("toProto: unsupported actor type: %q", t)
	}
}

func toProtoAction(a model.ActionType) (ingestionv1.Action, error) {
	switch a {
	case model.ActionAccess:
		return ingestionv1.Action_ACTION_ACCESS, nil
	case model.ActionCreate:
		return ingestionv1.Action_ACTION_CREATE, nil
	case model.ActionUpdate:
		return ingestionv1.Action_ACTION_UPDATE, nil
	case model.ActionDelete:
		return ingestionv1.Action_ACTION_DELETE, nil
	case model.ActionShare:
		return ingestionv1.Action_ACTION_SHARE, nil
	case model.ActionExport:
		return ingestionv1.Action_ACTION_EXPORT, nil
	case model.ActionLogin:
		return ingestionv1.Action_ACTION_LOGIN, nil
	case model.ActionLogout:
		return ingestionv1.Action_ACTION_LOGOUT, nil
	case model.ActionAuthorize:
		return ingestionv1.Action_ACTION_AUTHORIZE, nil
	default:
		return 0, fmt.Errorf("toProto: unsupported action: %q", a)
	}
}

func toProtoResultStatus(s model.ResultStatus) (ingestionv1.Result_Status, error) {
	switch s {
	case model.ResultStatusSuccess:
		return ingestionv1.Result_STATUS_SUCCESS, nil
	case model.ResultStatusFailure:
		return ingestionv1.Result_STATUS_FAILURE, nil
	default:
		return 0, fmt.Errorf("toProto: unsupported result status: %q", s)
	}
}
