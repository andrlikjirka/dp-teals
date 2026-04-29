package model

import (
	"testing"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// validID is a fixed UUID used throughout the tests.
const validID = "550e8400-e29b-41d4-a716-446655440000"

// minValidRequest returns a minimal AppendRequest that satisfies all model
// validations so tests can override individual fields.
func minValidRequest() *auditv1.AppendRequest {
	return &auditv1.AppendRequest{
		Event: &auditv1.AuditEvent{
			Id:        validID,
			Timestamp: timestamppb.New(time.Now()),
			Actor:     &auditv1.Actor{Type: auditv1.Actor_TYPE_USER, Id: "actor-1"},
			Subject:   &auditv1.Subject{Id: "subject-1"},
			Action:    auditv1.Action_ACTION_ACCESS,
			Resource:  &auditv1.Resource{Id: "res-1", Name: "resource-name"},
			Result:    &auditv1.Result{Status: auditv1.Result_STATUS_SUCCESS},
		},
	}
}

// --- MapToAuditEvent ---

func TestMapToAuditEvent_NilRequest(t *testing.T) {
	_, err := MapToAuditEvent(nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestMapToAuditEvent_NilEvent(t *testing.T) {
	_, err := MapToAuditEvent(&auditv1.AppendRequest{Event: nil})
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestMapToAuditEvent_InvalidUUID(t *testing.T) {
	req := minValidRequest()
	req.Event.Id = "not-a-uuid"
	_, err := MapToAuditEvent(req)
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestMapToAuditEvent_UnsupportedActorType(t *testing.T) {
	req := minValidRequest()
	req.Event.Actor = &auditv1.Actor{Type: auditv1.Actor_TYPE_UNSPECIFIED, Id: "actor-1"}
	_, err := MapToAuditEvent(req)
	if err == nil {
		t.Fatal("expected error for unsupported actor type")
	}
}

func TestMapToAuditEvent_UnsupportedAction(t *testing.T) {
	req := minValidRequest()
	req.Event.Action = auditv1.Action_ACTION_UNSPECIFIED
	_, err := MapToAuditEvent(req)
	if err == nil {
		t.Fatal("expected error for unsupported action")
	}
}

func TestMapToAuditEvent_UnsupportedResultStatus(t *testing.T) {
	req := minValidRequest()
	req.Event.Result = &auditv1.Result{Status: auditv1.Result_STATUS_UNSPECIFIED}
	_, err := MapToAuditEvent(req)
	if err == nil {
		t.Fatal("expected error for unsupported result status")
	}
}

func TestMapToAuditEvent_ValidMapping(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	reason := "denied"
	req := &auditv1.AppendRequest{
		Event: &auditv1.AuditEvent{
			Id:        validID,
			Timestamp: timestamppb.New(ts),
			Environment: &auditv1.Environment{
				Service: "svc",
				TraceId: "trace-123",
				SpanId:  "span-456",
			},
			Actor:   &auditv1.Actor{Type: auditv1.Actor_TYPE_SYSTEM, Id: "sys-1"},
			Subject: &auditv1.Subject{Id: "subj-42"},
			Action:  auditv1.Action_ACTION_DELETE,
			Resource: &auditv1.Resource{
				Id:     "r-1",
				Name:   "patient-record",
				Fields: []string{"ssn", "dob"},
			},
			Result:   &auditv1.Result{Status: auditv1.Result_STATUS_FAILURE, Reason: &reason},
			Metadata: mustStruct(t, map[string]any{"ip": "1.2.3.4"}),
		},
	}

	e, err := MapToAuditEvent(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.ID.String() != validID {
		t.Errorf("ID: got %v, want %v", e.ID, validID)
	}
	if !e.Timestamp.Equal(ts) {
		t.Errorf("Timestamp: got %v, want %v", e.Timestamp, ts)
	}
	if e.Environment == nil || e.Environment.Service != "svc" || e.Environment.TraceID != "trace-123" || e.Environment.SpanID != "span-456" {
		t.Errorf("Environment: %+v", e.Environment)
	}
	if e.Actor.Type != enum.ActorTypeSystem || e.Actor.ID != "sys-1" {
		t.Errorf("Actor: %+v", e.Actor)
	}
	if e.Subject.ID != "subj-42" {
		t.Errorf("Subject: got %v", e.Subject.ID)
	}
	if e.Action != enum.ActionTypeDelete {
		t.Errorf("Action: got %v", e.Action)
	}
	if e.Resource.ID != "r-1" || e.Resource.Name != "patient-record" || len(e.Resource.Fields) != 2 {
		t.Errorf("Resource: %+v", e.Resource)
	}
	if e.Result.Status != enum.ResultStatusFailure || e.Result.Reason != reason {
		t.Errorf("Result: %+v", e.Result)
	}
	if e.Metadata["ip"] != "1.2.3.4" {
		t.Errorf("Metadata: %+v", e.Metadata)
	}
}

func TestMapToAuditEvent_NilOptionalFields(t *testing.T) {
	req := minValidRequest()
	req.Event.Environment = nil
	req.Event.Metadata = nil
	req.Event.Resource = &auditv1.Resource{Id: "r-1", Name: "res"}

	e, err := MapToAuditEvent(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Environment != nil {
		t.Errorf("expected nil Environment, got %+v", e.Environment)
	}
	if len(e.Metadata) != 0 {
		t.Errorf("expected empty Metadata, got %+v", e.Metadata)
	}
}

// --- toActor ---

func TestToActor_Nil(t *testing.T) {
	a, err := toActor(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Type != "" || a.ID != "" {
		t.Errorf("expected zero Actor, got %+v", a)
	}
}

func TestToActor_Variants(t *testing.T) {
	tests := []struct {
		name     string
		proto    auditv1.Actor_Type
		wantType enum.ActorType
		wantErr  bool
	}{
		{"user", auditv1.Actor_TYPE_USER, enum.ActorTypeUser, false},
		{"system", auditv1.Actor_TYPE_SYSTEM, enum.ActorTypeSystem, false},
		{"unspecified", auditv1.Actor_TYPE_UNSPECIFIED, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toActor(&auditv1.Actor{Type: tc.proto, Id: "id"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if !tc.wantErr && got.Type != tc.wantType {
				t.Errorf("ActorType: got %v, want %v", got.Type, tc.wantType)
			}
		})
	}
}

// --- toAction ---

func TestToAction_AllVariants(t *testing.T) {
	tests := []struct {
		proto   auditv1.Action
		want    enum.ActionType
		wantErr bool
	}{
		{auditv1.Action_ACTION_ACCESS, enum.ActionTypeAccess, false},
		{auditv1.Action_ACTION_CREATE, enum.ActionTypeCreate, false},
		{auditv1.Action_ACTION_UPDATE, enum.ActionTypeUpdate, false},
		{auditv1.Action_ACTION_DELETE, enum.ActionTypeDelete, false},
		{auditv1.Action_ACTION_SHARE, enum.ActionTypeShare, false},
		{auditv1.Action_ACTION_EXPORT, enum.ActionTypeExport, false},
		{auditv1.Action_ACTION_LOGIN, enum.ActionTypeLogin, false},
		{auditv1.Action_ACTION_LOGOUT, enum.ActionTypeLogout, false},
		{auditv1.Action_ACTION_AUTHORIZE, enum.ActionTypeAuthorize, false},
		{auditv1.Action_ACTION_UNSPECIFIED, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.proto.String(), func(t *testing.T) {
			got, err := toAction(tc.proto)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("ActionType: got %v, want %v", got, tc.want)
			}
		})
	}
}

// --- toResult ---

func TestToResult_Nil(t *testing.T) {
	r, err := toResult(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != "" || r.Reason != "" {
		t.Errorf("expected zero Result, got %+v", r)
	}
}

func TestToResult_Variants(t *testing.T) {
	reason := "forbidden"
	tests := []struct {
		name    string
		input   *auditv1.Result
		want    svcmodel.Result
		wantErr bool
	}{
		{
			name:  "success",
			input: &auditv1.Result{Status: auditv1.Result_STATUS_SUCCESS},
			want:  svcmodel.Result{Status: enum.ResultStatusSuccess},
		},
		{
			name:  "failure with reason",
			input: &auditv1.Result{Status: auditv1.Result_STATUS_FAILURE, Reason: &reason},
			want:  svcmodel.Result{Status: enum.ResultStatusFailure, Reason: reason},
		},
		{
			name:    "unspecified",
			input:   &auditv1.Result{Status: auditv1.Result_STATUS_UNSPECIFIED},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toResult(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if !tc.wantErr {
				if got.Status != tc.want.Status {
					t.Errorf("Status: got %v, want %v", got.Status, tc.want.Status)
				}
				if got.Reason != tc.want.Reason {
					t.Errorf("Reason: got %v, want %v", got.Reason, tc.want.Reason)
				}
			}
		})
	}
}

// --- MapToAuditEventFilter ---

func TestMapToAuditEventFilter_Nil(t *testing.T) {
	f := MapToAuditEventFilter(nil)
	if f.ActorID != "" || f.SubjectID != "" || len(f.Actions) != 0 {
		t.Errorf("expected zero filter for nil input, got %+v", f)
	}
}

func TestMapToAuditEventFilter_ScalarFields(t *testing.T) {
	f := MapToAuditEventFilter(&auditv1.AuditEventFilter{
		ActorId:      "actor-99",
		SubjectId:    "subject-99",
		ResourceId:   "res-99",
		ResourceName: "records",
	})

	if f.ActorID != "actor-99" {
		t.Errorf("ActorID: got %v", f.ActorID)
	}
	if f.SubjectID != "subject-99" {
		t.Errorf("SubjectID: got %v", f.SubjectID)
	}
	if f.ResourceID != "res-99" {
		t.Errorf("ResourceID: got %v", f.ResourceID)
	}
	if f.ResourceName != "records" {
		t.Errorf("ResourceName: got %v", f.ResourceName)
	}
}

func TestMapToAuditEventFilter_ActorTypes(t *testing.T) {
	tests := []struct {
		name  string
		input []auditv1.Actor_Type
		want  []enum.ActorType
	}{
		{"user only", []auditv1.Actor_Type{auditv1.Actor_TYPE_USER}, []enum.ActorType{enum.ActorTypeUser}},
		{"system only", []auditv1.Actor_Type{auditv1.Actor_TYPE_SYSTEM}, []enum.ActorType{enum.ActorTypeSystem}},
		{"both", []auditv1.Actor_Type{auditv1.Actor_TYPE_USER, auditv1.Actor_TYPE_SYSTEM}, []enum.ActorType{enum.ActorTypeUser, enum.ActorTypeSystem}},
		{"unspecified ignored", []auditv1.Actor_Type{auditv1.Actor_TYPE_UNSPECIFIED}, []enum.ActorType{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := MapToAuditEventFilter(&auditv1.AuditEventFilter{ActorTypes: tc.input})
			if len(f.ActorTypes) != len(tc.want) {
				t.Fatalf("ActorTypes len: got %d, want %d", len(f.ActorTypes), len(tc.want))
			}
			for i, at := range tc.want {
				if f.ActorTypes[i] != at {
					t.Errorf("[%d] ActorType: got %v, want %v", i, f.ActorTypes[i], at)
				}
			}
		})
	}
}

func TestMapToAuditEventFilter_ResultStatuses(t *testing.T) {
	tests := []struct {
		name  string
		input []auditv1.Result_Status
		want  []enum.ResultStatusType
	}{
		{"success", []auditv1.Result_Status{auditv1.Result_STATUS_SUCCESS}, []enum.ResultStatusType{enum.ResultStatusSuccess}},
		{"failure", []auditv1.Result_Status{auditv1.Result_STATUS_FAILURE}, []enum.ResultStatusType{enum.ResultStatusFailure}},
		{"both", []auditv1.Result_Status{auditv1.Result_STATUS_SUCCESS, auditv1.Result_STATUS_FAILURE}, []enum.ResultStatusType{enum.ResultStatusSuccess, enum.ResultStatusFailure}},
		{"unspecified ignored", []auditv1.Result_Status{auditv1.Result_STATUS_UNSPECIFIED}, []enum.ResultStatusType{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := MapToAuditEventFilter(&auditv1.AuditEventFilter{ResultStatuses: tc.input})
			if len(f.ResultStatuses) != len(tc.want) {
				t.Fatalf("ResultStatuses len: got %d, want %d", len(f.ResultStatuses), len(tc.want))
			}
			for i, rs := range tc.want {
				if f.ResultStatuses[i] != rs {
					t.Errorf("[%d] ResultStatus: got %v, want %v", i, f.ResultStatuses[i], rs)
				}
			}
		})
	}
}

func TestMapToAuditEventFilter_Actions(t *testing.T) {
	input := []auditv1.Action{
		auditv1.Action_ACTION_ACCESS,
		auditv1.Action_ACTION_CREATE,
		auditv1.Action_ACTION_UNSPECIFIED, // should be silently ignored
	}
	f := MapToAuditEventFilter(&auditv1.AuditEventFilter{Actions: input})
	if len(f.Actions) != 2 {
		t.Fatalf("expected 2 actions (UNSPECIFIED ignored), got %d", len(f.Actions))
	}
	if f.Actions[0] != enum.ActionTypeAccess || f.Actions[1] != enum.ActionTypeCreate {
		t.Errorf("Actions: got %v", f.Actions)
	}
}

func TestMapToAuditEventFilter_Timestamps(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	f := MapToAuditEventFilter(&auditv1.AuditEventFilter{
		TimestampFrom: timestamppb.New(from),
		TimestampTo:   timestamppb.New(to),
	})

	if f.TimestampFrom == nil || !f.TimestampFrom.Equal(from) {
		t.Errorf("TimestampFrom: got %v, want %v", f.TimestampFrom, from)
	}
	if f.TimestampTo == nil || !f.TimestampTo.Equal(to) {
		t.Errorf("TimestampTo: got %v, want %v", f.TimestampTo, to)
	}
}

func TestMapToAuditEventFilter_NilTimestamps(t *testing.T) {
	f := MapToAuditEventFilter(&auditv1.AuditEventFilter{})
	if f.TimestampFrom != nil || f.TimestampTo != nil {
		t.Errorf("expected nil timestamps, got from=%v to=%v", f.TimestampFrom, f.TimestampTo)
	}
}

// --- EncodeCursor / DecodeCursor ---

func TestCursor_RoundTrip(t *testing.T) {
	tests := []int64{0, 1, 42, 1<<32 - 1, 1<<62 - 1}
	for _, id := range tests {
		encoded := EncodeCursor(id)
		decoded, err := DecodeCursor(encoded)
		if err != nil {
			t.Errorf("DecodeCursor(%v) error: %v", encoded, err)
			continue
		}
		if decoded != id {
			t.Errorf("round-trip %d: got %d", id, decoded)
		}
	}
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, err := DecodeCursor("not!!valid==base64")
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}

func TestDecodeCursor_WrongLength(t *testing.T) {
	shortCursor := "AA"
	_, err := DecodeCursor(shortCursor)
	if err == nil {
		t.Fatal("expected error for decoded length != 8")
	}
}

func TestEncodeCursor_KnownValue(t *testing.T) {
	// 1 encodes to 8 big-endian bytes: 00 00 00 00 00 00 00 01
	encoded := EncodeCursor(1)
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded != 1 {
		t.Errorf("got %d, want 1", decoded)
	}
}

// --- mapToProtoAuditEvent ---

func TestMapToProtoAuditEvent_Nil(t *testing.T) {
	if mapToProtoAuditEvent(nil) != nil {
		t.Error("expected nil proto for nil input")
	}
}

func TestMapToProtoAuditEvent_FieldMapping(t *testing.T) {
	e := buildValidAuditEvent(t)
	e.Metadata = map[string]any{"key": "val"}

	proto := mapToProtoAuditEvent(e)
	if proto == nil {
		t.Fatal("expected non-nil proto")
	}
	if proto.Id != e.ID.String() {
		t.Errorf("Id: got %v, want %v", proto.Id, e.ID)
	}
	if proto.Actor.Id != e.Actor.ID {
		t.Errorf("Actor.Id: got %v", proto.Actor.Id)
	}
	if proto.Metadata == nil {
		t.Error("expected Metadata to be set")
	}
}

func TestMapToProtoAuditEvent_WithEnvironment(t *testing.T) {
	e := buildValidAuditEvent(t)
	e.Environment = &svcmodel.Environment{Service: "svc", TraceID: "t", SpanID: "s"}

	proto := mapToProtoAuditEvent(e)
	if proto.Environment == nil {
		t.Fatal("expected Environment in proto")
	}
	if proto.Environment.Service != "svc" {
		t.Errorf("Environment.Service: got %v", proto.Environment.Service)
	}
}

func TestMapToProtoAuditEvent_ResultReasonOmittedWhenEmpty(t *testing.T) {
	e := buildValidAuditEvent(t)
	proto := mapToProtoAuditEvent(e)
	if proto.Result == nil {
		t.Fatal("expected Result in proto")
	}
	if proto.Result.Reason != nil {
		t.Errorf("expected nil Reason for empty string, got %v", *proto.Result.Reason)
	}
}

// --- mapToProtoProtectedAuditEvent ---

func TestMapToProtoProtectedAuditEvent_Nil(t *testing.T) {
	if mapToProtoProtectedAuditEvent(nil) != nil {
		t.Error("expected nil proto for nil input")
	}
}

func TestMapToProtoProtectedAuditEvent_WithProtectedMetadata(t *testing.T) {
	e := buildValidProtectedAuditEvent(t)
	e.ProtectedMetadata = &svcmodel.ProtectedMetadata{
		Ciphertext: []byte("ct"),
		WrappedDEK: []byte("dek"),
		Commitment: []byte("com"),
	}

	proto := mapToProtoProtectedAuditEvent(e)
	if proto.ProtectedMetadata == nil {
		t.Fatal("expected ProtectedMetadata in proto")
	}
	if string(proto.ProtectedMetadata.Ciphertext) != "ct" {
		t.Errorf("Ciphertext: got %v", proto.ProtectedMetadata.Ciphertext)
	}
	if string(proto.ProtectedMetadata.WrappedDek) != "dek" {
		t.Errorf("WrappedDek: got %v", proto.ProtectedMetadata.WrappedDek)
	}
	if string(proto.ProtectedMetadata.Commitment) != "com" {
		t.Errorf("Commitment: got %v", proto.ProtectedMetadata.Commitment)
	}
}

func TestMapToProtoProtectedAuditEvent_WithoutProtectedMetadata(t *testing.T) {
	e := buildValidProtectedAuditEvent(t)
	e.ProtectedMetadata = nil

	proto := mapToProtoProtectedAuditEvent(e)
	if proto.ProtectedMetadata != nil {
		t.Errorf("expected nil ProtectedMetadata in proto, got %+v", proto.ProtectedMetadata)
	}
}

// --- helpers ---

func buildValidAuditEvent(t *testing.T) *svcmodel.AuditEvent {
	t.Helper()
	e, err := svcmodel.NewAuditEvent(svcmodel.CreateAuditEventParams{
		BaseEventParams: svcmodel.BaseEventParams{
			ID:        uuid.MustParse(validID),
			Timestamp: time.Now(),
			Actor:     svcmodel.Actor{Type: enum.ActorTypeUser, ID: "actor-1"},
			Subject:   svcmodel.Subject{ID: "subject-1"},
			Action:    enum.ActionTypeAccess,
			Resource:  svcmodel.Resource{ID: "r-1", Name: "resource"},
			Result:    svcmodel.Result{Status: enum.ResultStatusSuccess},
		},
	})
	if err != nil {
		t.Fatalf("build AuditEvent: %v", err)
	}
	return e
}

func buildValidProtectedAuditEvent(t *testing.T) *svcmodel.ProtectedAuditEvent {
	t.Helper()
	e, err := svcmodel.NewProtectedAuditEvent(svcmodel.CreateProtectedAuditEventParams{
		BaseEventParams: svcmodel.BaseEventParams{
			ID:        uuid.MustParse(validID),
			Timestamp: time.Now(),
			Actor:     svcmodel.Actor{Type: enum.ActorTypeUser, ID: "actor-1"},
			Subject:   svcmodel.Subject{ID: "subject-1"},
			Action:    enum.ActionTypeAccess,
			Resource:  svcmodel.Resource{ID: "r-1", Name: "resource"},
			Result:    svcmodel.Result{Status: enum.ResultStatusSuccess},
		},
	})
	if err != nil {
		t.Fatalf("build ProtectedAuditEvent: %v", err)
	}
	return e
}

func mustStruct(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("mustStruct: %v", err)
	}
	return s
}
