package model

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapToAuditEvent converts an AppendRequest to an AuditEvent model.
func MapToAuditEvent(req *auditv1.AppendRequest) (*model.AuditEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("audit event request cannot be nil")
	}

	ev := req.GetEvent()
	if ev == nil {
		return nil, fmt.Errorf("event payload cannot be nil")
	}

	id, err := uuid.Parse(ev.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid event id %q: %w", ev.GetId(), err)
	}

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
		BaseEventParams: model.BaseEventParams{
			ID:          id,
			Timestamp:   timestamp,
			Environment: toEnvironment(ev.GetEnvironment()),
			Actor:       actor,
			Subject:     toSubject(ev.GetSubject()),
			Action:      action,
			Resource:    toResource(ev.GetResource()),
			Result:      result,
		},
		Metadata: metadata,
	})
}

func toEnvironment(env *auditv1.Environment) *model.Environment {
	if env == nil {
		return nil
	}

	return &model.Environment{
		Service: env.GetService(),
		TraceID: env.GetTraceId(),
		SpanID:  env.GetSpanId(),
	}
}

func toActor(actor *auditv1.Actor) (model.Actor, error) {
	if actor == nil {
		return model.Actor{}, nil
	}

	var actorType enum.ActorType
	switch actor.GetType() {
	case auditv1.Actor_TYPE_USER:
		actorType = enum.ActorTypeUser
	case auditv1.Actor_TYPE_SYSTEM:
		actorType = enum.ActorTypeSystem
	default:
		return model.Actor{}, fmt.Errorf("unsupported actor type: %v", actor.GetType())
	}

	return model.Actor{
		Type: actorType,
		ID:   actor.GetId(),
	}, nil
}

func toAction(action auditv1.Action) (enum.ActionType, error) {
	switch action {
	case auditv1.Action_ACTION_ACCESS:
		return enum.ActionTypeAccess, nil
	case auditv1.Action_ACTION_CREATE:
		return enum.ActionTypeCreate, nil
	case auditv1.Action_ACTION_UPDATE:
		return enum.ActionTypeUpdate, nil
	case auditv1.Action_ACTION_DELETE:
		return enum.ActionTypeDelete, nil
	case auditv1.Action_ACTION_SHARE:
		return enum.ActionTypeShare, nil
	case auditv1.Action_ACTION_EXPORT:
		return enum.ActionTypeExport, nil
	case auditv1.Action_ACTION_LOGIN:
		return enum.ActionTypeLogin, nil
	case auditv1.Action_ACTION_LOGOUT:
		return enum.ActionTypeLogout, nil
	case auditv1.Action_ACTION_AUTHORIZE:
		return enum.ActionTypeAuthorize, nil
	default:
		return "", fmt.Errorf("unsupported action type: %v", action)
	}
}

func toSubject(subject *auditv1.Subject) model.Subject {
	if subject == nil {
		return model.Subject{}
	}

	return model.Subject{
		ID: subject.GetId(),
	}
}

func toResource(resource *auditv1.Resource) model.Resource {
	if resource == nil {
		return model.Resource{}
	}

	return model.Resource{
		ID:     resource.GetId(),
		Name:   resource.GetName(),
		Fields: resource.GetFields(),
	}
}

func toResult(result *auditv1.Result) (model.Result, error) {
	if result == nil {
		return model.Result{}, nil
	}

	var status enum.ResultStatusType
	switch result.GetStatus() {
	case auditv1.Result_STATUS_SUCCESS:
		status = enum.ResultStatusSuccess
	case auditv1.Result_STATUS_FAILURE:
		status = enum.ResultStatusFailure
	default:
		return model.Result{}, fmt.Errorf("unsupported result status: %v", result.GetStatus())
	}

	return model.Result{
		Status: status,
		Reason: result.GetReason(),
	}, nil
}

// mapToProtoAuditEvent converts a service model AuditEvent to the proto AuditEvent message.
func mapToProtoAuditEvent(event *model.AuditEvent) *auditv1.AuditEvent {
	if event == nil {
		return nil
	}

	proto := &auditv1.AuditEvent{
		Id:        event.ID.String(),
		Timestamp: timestamppb.New(event.Timestamp),
		Actor: &auditv1.Actor{
			Type: fromActorType(event.Actor.Type),
			Id:   event.Actor.ID,
		},
		Subject: &auditv1.Subject{Id: event.Subject.ID},
		Action:  fromActionType(event.Action),
		Resource: &auditv1.Resource{
			Id:     event.Resource.ID,
			Name:   event.Resource.Name,
			Fields: event.Resource.Fields,
		},
	}

	if event.Environment != nil {
		proto.Environment = &auditv1.Environment{
			Service: event.Environment.Service,
			TraceId: event.Environment.TraceID,
			SpanId:  event.Environment.SpanID,
		}
	}

	result := &auditv1.Result{
		Status: fromResultStatus(event.Result.Status),
	}
	if event.Result.Reason != "" {
		result.Reason = &event.Result.Reason
	}
	proto.Result = result

	if len(event.Metadata) > 0 {
		if s, err := structpb.NewStruct(event.Metadata); err == nil {
			proto.Metadata = s
		}
	}

	return proto
}

// mapToProtoProtectedAuditEvent converts a service model ProtectedAuditEvent to the proto ProtectedAuditEvent message, including the protected metadata if present.
func mapToProtoProtectedAuditEvent(event *model.ProtectedAuditEvent) *auditv1.ProtectedAuditEvent {
	if event == nil {
		return nil
	}

	proto := &auditv1.ProtectedAuditEvent{
		Id:        event.ID.String(),
		Timestamp: timestamppb.New(event.Timestamp),
		Actor: &auditv1.Actor{
			Type: fromActorType(event.Actor.Type),
			Id:   event.Actor.ID,
		},
		Subject: &auditv1.Subject{Id: event.Subject.ID},
		Action:  fromActionType(event.Action),
		Resource: &auditv1.Resource{
			Id:     event.Resource.ID,
			Name:   event.Resource.Name,
			Fields: event.Resource.Fields,
		},
	}

	if event.Environment != nil {
		proto.Environment = &auditv1.Environment{
			Service: event.Environment.Service,
			TraceId: event.Environment.TraceID,
			SpanId:  event.Environment.SpanID,
		}
	}

	if event.ProtectedMetadata != nil {
		proto.ProtectedMetadata = &auditv1.ProtectedMetadata{
			Ciphertext: event.ProtectedMetadata.Ciphertext,
			WrappedDek: event.ProtectedMetadata.WrappedDEK,
			Commitment: event.ProtectedMetadata.Commitment,
		}
	}

	result := &auditv1.Result{
		Status: fromResultStatus(event.Result.Status),
	}
	if event.Result.Reason != "" {
		result.Reason = &event.Result.Reason
	}
	proto.Result = result

	return proto
}

func fromActorType(t enum.ActorType) auditv1.Actor_Type {
	switch t {
	case enum.ActorTypeUser:
		return auditv1.Actor_TYPE_USER
	case enum.ActorTypeSystem:
		return auditv1.Actor_TYPE_SYSTEM
	default:
		return auditv1.Actor_TYPE_UNSPECIFIED
	}
}

func fromActionType(a enum.ActionType) auditv1.Action {
	switch a {
	case enum.ActionTypeAccess:
		return auditv1.Action_ACTION_ACCESS
	case enum.ActionTypeCreate:
		return auditv1.Action_ACTION_CREATE
	case enum.ActionTypeUpdate:
		return auditv1.Action_ACTION_UPDATE
	case enum.ActionTypeDelete:
		return auditv1.Action_ACTION_DELETE
	case enum.ActionTypeShare:
		return auditv1.Action_ACTION_SHARE
	case enum.ActionTypeExport:
		return auditv1.Action_ACTION_EXPORT
	case enum.ActionTypeLogin:
		return auditv1.Action_ACTION_LOGIN
	case enum.ActionTypeLogout:
		return auditv1.Action_ACTION_LOGOUT
	case enum.ActionTypeAuthorize:
		return auditv1.Action_ACTION_AUTHORIZE
	default:
		return auditv1.Action_ACTION_UNSPECIFIED
	}
}

func fromResultStatus(s enum.ResultStatusType) auditv1.Result_Status {
	switch s {
	case enum.ResultStatusSuccess:
		return auditv1.Result_STATUS_SUCCESS
	case enum.ResultStatusFailure:
		return auditv1.Result_STATUS_FAILURE
	default:
		return auditv1.Result_STATUS_UNSPECIFIED
	}
}

// MapToAuditEventFilter converts a proto AuditEventFilter message to the service model AuditEventFilter.
func MapToAuditEventFilter(f *auditv1.AuditEventFilter) model.AuditEventFilter {
	if f == nil {
		return model.AuditEventFilter{}
	}

	filter := model.AuditEventFilter{
		ActorID:      f.GetActorId(),
		SubjectID:    f.GetSubjectId(),
		ResourceID:   f.GetResourceId(),
		ResourceName: f.GetResourceName(),
	}

	for _, a := range f.GetActions() {
		if at, err := toAction(a); err == nil {
			filter.Actions = append(filter.Actions, at)
		}
	}

	for _, t := range f.GetActorTypes() {
		switch t {
		case auditv1.Actor_TYPE_USER:
			filter.ActorTypes = append(filter.ActorTypes, enum.ActorTypeUser)
		case auditv1.Actor_TYPE_SYSTEM:
			filter.ActorTypes = append(filter.ActorTypes, enum.ActorTypeSystem)
		}
	}

	for _, s := range f.GetResultStatuses() {
		switch s {
		case auditv1.Result_STATUS_SUCCESS:
			filter.ResultStatuses = append(filter.ResultStatuses, enum.ResultStatusSuccess)
		case auditv1.Result_STATUS_FAILURE:
			filter.ResultStatuses = append(filter.ResultStatuses, enum.ResultStatusFailure)
		}
	}

	if ts := f.GetTimestampFrom(); ts != nil {
		t := ts.AsTime()
		filter.TimestampFrom = &t
	}

	if ts := f.GetTimestampTo(); ts != nil {
		t := ts.AsTime()
		filter.TimestampTo = &t
	}

	return filter
}

// EncodeCursor encodes an int64 ID into a base64 string for use as a pagination cursor.
func EncodeCursor(id int64) string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor decodes a base64 string cursor back into an int64 ID. It returns an error if the input is not a valid base64 string or if the decoded byte length is not 8.
func DecodeCursor(s string) (int64, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor: %w", err)
	}
	if len(b) != 8 {
		return 0, fmt.Errorf("invalid cursor: unexpected length %d", len(b))
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}
