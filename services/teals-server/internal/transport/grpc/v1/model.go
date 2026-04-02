package v1

import (
	"fmt"
	"time"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/model"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/model/enum"
	"github.com/google/uuid"
)

// MapToAuditEvent converts an AppendRequest to an AuditEvent model.
func MapToAuditEvent(req *ingestionv1.AppendRequest) (*model.AuditEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("audit event request cannot be nil")
	}

	ev := req.GetEvent()
	if ev == nil {
		return nil, fmt.Errorf("event payload cannot be nil")
	}

	id, _ := uuid.Parse(ev.GetId())

	var timestamp time.Time
	if ts := ev.GetTimestamp(); ts != nil {
		timestamp = ts.AsTime()
	}

	var metadata map[string]any
	if md := ev.GetMetadata(); md != nil {
		metadata = md.AsMap()
	}

	actor, err := toActor(ev.GetActor())
	if err != nil {
		return nil, err
	}

	action, err := toAction(ev.GetAction())
	if err != nil {
		return nil, err
	}

	result, err := toResult(ev.GetResult())
	if err != nil {
		return nil, err
	}

	return model.NewAuditEvent(model.CreateAuditEventParams{
		ID:          id,
		Timestamp:   timestamp,
		Environment: toEnvironment(ev.GetEnvironment()),
		Actor:       actor,
		Subject:     toSubject(ev.GetSubject()),
		Action:      action,
		Resource:    toResource(ev.GetResource()),
		Result:      result,
		Metadata:    metadata,
	})
}

func toEnvironment(env *ingestionv1.Environment) *model.Environment {
	if env == nil {
		return nil
	}

	return &model.Environment{
		Service: env.GetService(),
		TraceID: env.GetTraceId(),
		SpanID:  env.GetSpanId(),
	}
}

func toActor(actor *ingestionv1.Actor) (model.Actor, error) {
	if actor == nil {
		return model.Actor{}, nil
	}

	var actorType enum.ActorType
	switch actor.GetType() {
	case ingestionv1.Actor_TYPE_USER:
		actorType = enum.ActorTypeUser
	case ingestionv1.Actor_TYPE_SYSTEM:
		actorType = enum.ActorTypeSystem
	default:
		return model.Actor{}, fmt.Errorf("unsupported actor type: %v", actor.GetType())
	}

	return model.Actor{
		Type: actorType,
		ID:   actor.GetId(),
	}, nil
}

func toAction(action ingestionv1.ActionType) (enum.ActionType, error) {
	switch action {
	case ingestionv1.ActionType_ACTION_ACCESS:
		return enum.ActionTypeAccess, nil
	case ingestionv1.ActionType_ACTION_CREATE:
		return enum.ActionTypeCreate, nil
	case ingestionv1.ActionType_ACTION_UPDATE:
		return enum.ActionTypeUpdate, nil
	case ingestionv1.ActionType_ACTION_DELETE:
		return enum.ActionTypeDelete, nil
	case ingestionv1.ActionType_ACTION_SHARE:
		return enum.ActionTypeShare, nil
	case ingestionv1.ActionType_ACTION_EXPORT:
		return enum.ActionTypeExport, nil
	case ingestionv1.ActionType_ACTION_LOGIN:
		return enum.ActionTypeLogin, nil
	case ingestionv1.ActionType_ACTION_LOGOUT:
		return enum.ActionTypeLogout, nil
	default:
		return "", fmt.Errorf("unsupported action type: %v", action)
	}
}

func toSubject(subject *ingestionv1.Subject) model.Subject {
	if subject == nil {
		return model.Subject{}
	}

	return model.Subject{
		ID: subject.GetId(),
	}
}

func toResource(resource *ingestionv1.Resource) model.Resource {
	if resource == nil {
		return model.Resource{}
	}

	return model.Resource{
		ID:     resource.GetId(),
		Name:   resource.GetName(),
		Fields: resource.GetFields(),
	}
}

func toResult(result *ingestionv1.Result) (model.Result, error) {
	if result == nil {
		return model.Result{}, nil
	}

	var status enum.ResultStatusType
	switch result.GetStatus() {
	case ingestionv1.Result_STATUS_SUCCESS:
		status = enum.ResultStatusSuccess
	case ingestionv1.Result_STATUS_FAILURE:
		status = enum.ResultStatusFailure
	default:
		return model.Result{}, fmt.Errorf("unsupported result status: %v", result.GetStatus())
	}

	return model.Result{
		Status: status,
		Reason: result.GetReason(),
	}, nil
}
