package v1

import (
	"fmt"
	"time"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/ingestion/model"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/ingestion/model/enum"
	"github.com/google/uuid"
)

// MapToAuditEvent converts an AppendRequest to an AuditEvent model
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

	return model.NewAuditEvent(model.CreateAuditEventParams{
		ID:          id,
		Timestamp:   timestamp,
		Environment: toEnvironment(ev.GetEnvironment()),
		Actor:       toActor(ev.GetActor()),
		Subject:     toSubject(ev.GetSubject()),
		Action:      toAction(ev.GetAction()),
		Resource:    toResource(ev.GetResource()),
		Result:      toResult(ev.GetResult()),
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

func toActor(actor *ingestionv1.Actor) model.Actor {
	if actor == nil {
		return model.Actor{} // Domain will catch empty ID and Type
	}

	var actorType enum.ActorType
	switch actor.GetType() {
	case ingestionv1.Actor_TYPE_USER:
		actorType = enum.ActorTypeUser
	case ingestionv1.Actor_TYPE_SYSTEM:
		actorType = enum.ActorTypeSystem
	default:
		return model.Actor{}
	}

	return model.Actor{
		Type: actorType,
		ID:   actor.GetId(),
	}
}

func toAction(action ingestionv1.Action) enum.ActionType {
	switch action {
	case ingestionv1.Action_ACTION_ACCESS:
		return enum.ActionTypeAccess
	case ingestionv1.Action_ACTION_CREATE:
		return enum.ActionTypeCreate
	case ingestionv1.Action_ACTION_UPDATE:
		return enum.ActionTypeUpdate
	case ingestionv1.Action_ACTION_DELETE:
		return enum.ActionTypeDelete
	case ingestionv1.Action_ACTION_SHARE:
		return enum.ActionTypeShare
	case ingestionv1.Action_ACTION_EXPORT:
		return enum.ActionTypeExport
	case ingestionv1.Action_ACTION_LOGIN:
		return enum.ActionTypeLogin
	case ingestionv1.Action_ACTION_LOGOUT:
		return enum.ActionTypeLogout
	default:
		return ""
	}
}

func toSubject(subject *ingestionv1.Subject) model.Subject {
	if subject == nil {
		return model.Subject{} // Domain will catch empty ID
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

func toResult(result *ingestionv1.Result) model.Result {
	if result == nil {
		return model.Result{}
	}

	var resultStatusType enum.ResultStatusType
	switch result.GetStatus() {
	case ingestionv1.Result_STATUS_SUCCESS:
		resultStatusType = enum.ResultTypeSuccess
	case ingestionv1.Result_STATUS_FAILURE:
		resultStatusType = enum.ResultTypeFailure
	default:
		return model.Result{}
	}

	return model.Result{
		Status: resultStatusType,
		Reason: result.GetReason(),
	}
}
