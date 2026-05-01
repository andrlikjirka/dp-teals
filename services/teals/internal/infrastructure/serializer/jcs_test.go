package serializer

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
)

// --- helpers ---

var fixedID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
var fixedTime = time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)

func newAuditEvent(t *testing.T, withEnv bool, metadata map[string]any) *svcmodel.AuditEvent {
	t.Helper()
	params := svcmodel.CreateAuditEventParams{
		BaseEventParams: svcmodel.BaseEventParams{
			ID:        fixedID,
			Timestamp: fixedTime,
			Actor:     svcmodel.Actor{Type: enum.ActorTypeUser, ID: "user-1"},
			Subject:   svcmodel.Subject{ID: "subject-1"},
			Action:    enum.ActionTypeCreate,
			Resource:  svcmodel.Resource{ID: "res-1", Name: "record", Fields: []string{"email", "name"}},
			Result:    svcmodel.Result{Status: enum.ResultStatusSuccess, Reason: "ok"},
		},
		Metadata: metadata,
	}
	if withEnv {
		params.Environment = &svcmodel.Environment{Service: "svc", TraceID: "trace-1", SpanID: "span-1"}
	}
	e, err := svcmodel.NewAuditEvent(params)
	if err != nil {
		t.Fatalf("newAuditEvent: %v", err)
	}
	return e
}

func newProtectedAuditEvent(t *testing.T, withEnv bool, pm *svcmodel.ProtectedMetadata) *svcmodel.ProtectedAuditEvent {
	t.Helper()
	params := svcmodel.CreateProtectedAuditEventParams{
		BaseEventParams: svcmodel.BaseEventParams{
			ID:        fixedID,
			Timestamp: fixedTime,
			Actor:     svcmodel.Actor{Type: enum.ActorTypeUser, ID: "user-1"},
			Subject:   svcmodel.Subject{ID: "subject-1"},
			Action:    enum.ActionTypeCreate,
			Resource:  svcmodel.Resource{ID: "res-1", Name: "record", Fields: []string{"email", "name"}},
			Result:    svcmodel.Result{Status: enum.ResultStatusSuccess, Reason: "ok"},
		},
		ProtectedMetadata: pm,
	}
	if withEnv {
		params.Environment = &svcmodel.Environment{Service: "svc", TraceID: "trace-1", SpanID: "span-1"}
	}
	e, err := svcmodel.NewProtectedAuditEvent(params)
	if err != nil {
		t.Fatalf("newProtectedAuditEvent: %v", err)
	}
	return e
}

// assertBaseFieldsEqual checks the fields common to both AuditEvent and ProtectedAuditEvent.
func assertBaseFieldsEqual(t *testing.T, gotID uuid.UUID, gotTS time.Time, gotActor svcmodel.Actor,
	gotSubject svcmodel.Subject, gotAction enum.ActionType, gotResource svcmodel.Resource,
	gotResult svcmodel.Result, gotEnv *svcmodel.Environment,
	wantID uuid.UUID, wantTS time.Time, wantActor svcmodel.Actor,
	wantSubject svcmodel.Subject, wantAction enum.ActionType, wantResource svcmodel.Resource,
	wantResult svcmodel.Result, wantEnv *svcmodel.Environment) {
	t.Helper()

	if gotID != wantID {
		t.Errorf("ID: got %v, want %v", gotID, wantID)
	}
	if !gotTS.Equal(wantTS) {
		t.Errorf("Timestamp: got %v, want %v", gotTS, wantTS)
	}
	if gotActor != wantActor {
		t.Errorf("Actor: got %+v, want %+v", gotActor, wantActor)
	}
	if gotSubject != wantSubject {
		t.Errorf("Subject: got %+v, want %+v", gotSubject, wantSubject)
	}
	if gotAction != wantAction {
		t.Errorf("Action: got %v, want %v", gotAction, wantAction)
	}
	if gotResource.ID != wantResource.ID || gotResource.Name != wantResource.Name {
		t.Errorf("Resource: got %+v, want %+v", gotResource, wantResource)
	}
	if !reflect.DeepEqual(gotResource.Fields, wantResource.Fields) {
		t.Errorf("Resource.Fields: got %v, want %v", gotResource.Fields, wantResource.Fields)
	}
	if gotResult != wantResult {
		t.Errorf("Result: got %+v, want %+v", gotResult, wantResult)
	}

	if wantEnv == nil && gotEnv != nil {
		t.Errorf("Environment: expected nil, got %+v", gotEnv)
	}
	if wantEnv != nil {
		if gotEnv == nil {
			t.Fatal("Environment: expected non-nil, got nil")
		}
		if *gotEnv != *wantEnv {
			t.Errorf("Environment: got %+v, want %+v", *gotEnv, *wantEnv)
		}
	}
}

func TestJcsSerializer_AuditEvent_RoundTrip(t *testing.T) {
	s := NewJcsSerializer()

	tests := []struct {
		name  string
		event *svcmodel.AuditEvent
	}{
		{
			name:  "all fields including environment and metadata",
			event: newAuditEvent(t, true, map[string]any{"key": "value", "flag": "true"}),
		},
		{
			name:  "without optional environment",
			event: newAuditEvent(t, false, map[string]any{"key": "value"}),
		},
		{
			name:  "without metadata",
			event: newAuditEvent(t, false, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := s.SerializeCanonicalAuditEvent(tc.event)
			if err != nil {
				t.Fatalf("serialize: %v", err)
			}
			if len(payload) == 0 {
				t.Fatal("serialize: empty output")
			}
			if !json.Valid(payload) {
				t.Fatalf("serialize: output is not valid JSON: %s", payload)
			}

			got, err := s.DeserializeCanonicalAuditEvent(payload)
			if err != nil {
				t.Fatalf("deserialize: %v", err)
			}

			want := tc.event
			assertBaseFieldsEqual(t,
				got.ID, got.Timestamp, got.Actor, got.Subject, got.Action, got.Resource, got.Result, got.Environment,
				want.ID, want.Timestamp, want.Actor, want.Subject, want.Action, want.Resource, want.Result, want.Environment,
			)
			if !reflect.DeepEqual(got.Metadata, want.Metadata) {
				t.Errorf("Metadata: got %v, want %v", got.Metadata, want.Metadata)
			}
		})
	}
}

func TestJcsSerializer_ProtectedAuditEvent_RoundTrip(t *testing.T) {
	s := NewJcsSerializer()

	pm := &svcmodel.ProtectedMetadata{
		Ciphertext: []byte("ciphertext-bytes"),
		WrappedDEK: []byte("wrapped-dek-bytes"),
		Commitment: []byte("commitment-bytes"),
	}

	tests := []struct {
		name  string
		event *svcmodel.ProtectedAuditEvent
	}{
		{
			name:  "all fields including environment and protected metadata",
			event: newProtectedAuditEvent(t, true, pm),
		},
		{
			name:  "without optional environment",
			event: newProtectedAuditEvent(t, false, pm),
		},
		{
			name:  "without protected metadata",
			event: newProtectedAuditEvent(t, false, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := s.SerializeCanonicalProtectedAuditEvent(tc.event)
			if err != nil {
				t.Fatalf("serialize: %v", err)
			}
			if len(payload) == 0 {
				t.Fatal("serialize: empty output")
			}
			if !json.Valid(payload) {
				t.Fatalf("serialize: output is not valid JSON: %s", payload)
			}

			got, err := s.DeserializeCanonicalProtectedAuditEvent(payload)
			if err != nil {
				t.Fatalf("deserialize: %v", err)
			}

			want := tc.event
			assertBaseFieldsEqual(t,
				got.ID, got.Timestamp, got.Actor, got.Subject, got.Action, got.Resource, got.Result, got.Environment,
				want.ID, want.Timestamp, want.Actor, want.Subject, want.Action, want.Resource, want.Result, want.Environment,
			)

			if want.ProtectedMetadata == nil && got.ProtectedMetadata != nil {
				t.Errorf("ProtectedMetadata: expected nil, got %+v", got.ProtectedMetadata)
			}
			if want.ProtectedMetadata != nil {
				if got.ProtectedMetadata == nil {
					t.Fatal("ProtectedMetadata: expected non-nil, got nil")
				}
				if !bytes.Equal(got.ProtectedMetadata.Ciphertext, want.ProtectedMetadata.Ciphertext) {
					t.Errorf("Ciphertext: got %x, want %x", got.ProtectedMetadata.Ciphertext, want.ProtectedMetadata.Ciphertext)
				}
				if !bytes.Equal(got.ProtectedMetadata.WrappedDEK, want.ProtectedMetadata.WrappedDEK) {
					t.Errorf("WrappedDEK: got %x, want %x", got.ProtectedMetadata.WrappedDEK, want.ProtectedMetadata.WrappedDEK)
				}
				if !bytes.Equal(got.ProtectedMetadata.Commitment, want.ProtectedMetadata.Commitment) {
					t.Errorf("Commitment: got %x, want %x", got.ProtectedMetadata.Commitment, want.ProtectedMetadata.Commitment)
				}
			}
		})
	}
}

func TestJcsSerializer_DeserializeAuditEvent_Errors(t *testing.T) {
	s := NewJcsSerializer()

	const validID = "550e8400-e29b-41d4-a716-446655440000"
	const validTS = "2024-01-15T10:30:45.123456789Z"

	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "malformed JSON",
			payload: `not-json`,
		},
		{
			name:    "invalid UUID",
			payload: `{"action":"CREATE","actor":{"id":"u1","type":"USER"},"id":"not-a-uuid","resource":{"id":"r1","name":"rec"},"result":{"status":"SUCCESS"},"subject":{"id":"s1"},"timestamp":"` + validTS + `"}`,
		},
		{
			name:    "invalid timestamp",
			payload: `{"action":"CREATE","actor":{"id":"u1","type":"USER"},"id":"` + validID + `","resource":{"id":"r1","name":"rec"},"result":{"status":"SUCCESS"},"subject":{"id":"s1"},"timestamp":"not-a-time"}`,
		},
		{
			name:    "invalid actor type",
			payload: `{"action":"CREATE","actor":{"id":"u1","type":"INVALID"},"id":"` + validID + `","resource":{"id":"r1","name":"rec"},"result":{"status":"SUCCESS"},"subject":{"id":"s1"},"timestamp":"` + validTS + `"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.DeserializeCanonicalAuditEvent(json.RawMessage(tc.payload))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestJcsSerializer_DeserializeProtectedAuditEvent_Errors(t *testing.T) {
	s := NewJcsSerializer()

	const validID = "550e8400-e29b-41d4-a716-446655440000"
	const validTS = "2024-01-15T10:30:45.123456789Z"

	validCiphertext := base64.StdEncoding.EncodeToString([]byte("enc"))
	validWrappedDEK := base64.StdEncoding.EncodeToString([]byte("dek"))
	validCommitment := hex.EncodeToString([]byte("cmt"))

	// Base payload with valid protected metadata — modified per row.
	baseWithPM := func(ciphertext, wrappedDEK, commitment string) string {
		return `{"action":"CREATE","actor":{"id":"u1","type":"USER"},"id":"` + validID + `","protected_metadata":{"ciphertext":"` +
			ciphertext + `","commitment":"` + commitment + `","wrapped_dek":"` + wrappedDEK +
			`"},"resource":{"id":"r1","name":"rec"},"result":{"status":"SUCCESS"},"subject":{"id":"s1"},"timestamp":"` + validTS + `"}`
	}

	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "malformed JSON",
			payload: `not-json`,
		},
		{
			name:    "invalid UUID",
			payload: `{"action":"CREATE","actor":{"id":"u1","type":"USER"},"id":"not-a-uuid","resource":{"id":"r1","name":"rec"},"result":{"status":"SUCCESS"},"subject":{"id":"s1"},"timestamp":"` + validTS + `"}`,
		},
		{
			name:    "invalid timestamp",
			payload: `{"action":"CREATE","actor":{"id":"u1","type":"USER"},"id":"` + validID + `","resource":{"id":"r1","name":"rec"},"result":{"status":"SUCCESS"},"subject":{"id":"s1"},"timestamp":"not-a-time"}`,
		},
		{
			name:    "invalid base64 ciphertext",
			payload: baseWithPM("not-base64!!", validWrappedDEK, validCommitment),
		},
		{
			name:    "invalid base64 wrapped_dek",
			payload: baseWithPM(validCiphertext, "not-base64!!", validCommitment),
		},
		{
			name:    "invalid hex commitment",
			payload: baseWithPM(validCiphertext, validWrappedDEK, "not-hex!!"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.DeserializeCanonicalProtectedAuditEvent(json.RawMessage(tc.payload))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
