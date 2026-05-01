package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// --- helpers ---

// newTestProtectedAuditEvent builds a minimal valid ProtectedAuditEvent with no protected metadata.
func newTestProtectedAuditEvent(t *testing.T) *svcmodel.ProtectedAuditEvent {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("generate uuid: %v", err)
	}
	e, err := svcmodel.NewProtectedAuditEvent(svcmodel.CreateProtectedAuditEventParams{
		BaseEventParams: svcmodel.BaseEventParams{
			ID:        id,
			Timestamp: time.Now().UTC(),
			Actor:     svcmodel.Actor{Type: enum.ActorTypeUser, ID: "user-1"},
			Subject:   svcmodel.Subject{ID: "subject-1"},
			Action:    enum.ActionTypeCreate,
			Resource:  svcmodel.Resource{ID: "res-1", Name: "record"},
			Result:    svcmodel.Result{Status: enum.ResultStatusSuccess},
		},
	})
	if err != nil {
		t.Fatalf("NewProtectedAuditEvent: %v", err)
	}
	return e
}

// newTestAuditLogEntryRaw builds a raw audit log entry with the given integer ID.
func newTestAuditLogEntryRaw(id int64) *svcmodel.AuditLogEntryRaw {
	eventID, _ := uuid.NewV7()
	return &svcmodel.AuditLogEntryRaw{
		ID:             &id,
		EventID:        eventID,
		SignatureToken: "sig-token",
		LeafIndex:      id,
		Payload:        json.RawMessage(`{}`),
	}
}

// defaultQueryRepos returns a Repositories with succeeding mocks.
func defaultQueryRepos() ports.Repositories {
	return ports.Repositories{
		AuditLog:           &mockAuditLog{},
		Ledger:             &mockLedger{},
		SubjectSecretStore: &mockSubjectSecretStore{},
	}
}

// deserializerReturning returns a mockSerializer whose deserialization always returns the given event.
func deserializerReturning(event *svcmodel.ProtectedAuditEvent) *mockSerializer {
	return &mockSerializer{
		DeserializeProtectedAuditEventFunc: func(_ json.RawMessage) (*svcmodel.ProtectedAuditEvent, error) {
			return event, nil
		},
	}
}

// --- GetAuditEvent ---

func TestQueryService_GetAuditEvent_Success(t *testing.T) {
	entry := newTestAuditLogEntryRaw(7)
	event := newTestProtectedAuditEvent(t)

	repos := defaultQueryRepos()
	repos.AuditLog = &mockAuditLog{
		GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
			return entry, nil
		},
	}

	svc := NewQueryService(&mockTx{repos: repos}, deserializerReturning(event), &mockProtector{}, newTestLogger())

	result, err := svc.GetAuditEvent(context.Background(), entry.EventID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Event != event {
		t.Error("result.Event does not match deserialized event")
	}
	if string(result.Payload) != string(entry.Payload) {
		t.Errorf("Payload: got %s, want %s", result.Payload, entry.Payload)
	}
	if result.LeafIndex != entry.LeafIndex {
		t.Errorf("LeafIndex: got %d, want %d", result.LeafIndex, entry.LeafIndex)
	}
	if result.SignatureToken != entry.SignatureToken {
		t.Errorf("SignatureToken: got %q, want %q", result.SignatureToken, entry.SignatureToken)
	}
	if result.RevealedMetadata != nil {
		t.Error("RevealedMetadata must be nil when event has no protected metadata")
	}
}

func TestQueryService_GetAuditEvent_RevealsMetadataWhenPresent(t *testing.T) {
	entry := newTestAuditLogEntryRaw(1)
	event := newTestProtectedAuditEvent(t)
	event.ProtectedMetadata = &svcmodel.ProtectedMetadata{
		Ciphertext: []byte("enc"),
		WrappedDEK: []byte("dek"),
		Commitment: []byte("cmt"),
	}
	revealed := map[string]any{"email": "user@example.com"}

	repos := defaultQueryRepos()
	repos.AuditLog = &mockAuditLog{
		GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
			return entry, nil
		},
	}

	protector := &mockProtector{
		RevealFunc: func(_ []byte, _ *svcmodel.ProtectedMetadata) (map[string]any, error) {
			return revealed, nil
		},
	}

	svc := NewQueryService(&mockTx{repos: repos}, deserializerReturning(event), protector, newTestLogger())

	result, err := svc.GetAuditEvent(context.Background(), entry.EventID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RevealedMetadata == nil {
		t.Fatal("expected non-nil RevealedMetadata")
	}
	if result.RevealedMetadata["email"] != "user@example.com" {
		t.Errorf("RevealedMetadata[email]: got %v, want %q", result.RevealedMetadata["email"], "user@example.com")
	}
}

func TestQueryService_GetAuditEvent_MetadataReveal_GracefulDegradation(t *testing.T) {
	tests := []struct {
		name      string
		repos     func() ports.Repositories
		protector *mockProtector
	}{
		{
			name: "subject secret not found",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return newTestAuditLogEntryRaw(1), nil
					},
				}
				r.SubjectSecretStore = &mockSubjectSecretStore{
					GetSecretBySubjectIDFunc: func(_ context.Context, _ string) ([]byte, error) {
						return nil, svcerrors.ErrSubjectSecretNotFound
					},
				}
				return r
			},
		},
		{
			name: "reveal fails",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return newTestAuditLogEntryRaw(1), nil
					},
				}
				return r
			},
			protector: &mockProtector{
				RevealFunc: func(_ []byte, _ *svcmodel.ProtectedMetadata) (map[string]any, error) {
					return nil, errors.New("decrypt error")
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := newTestProtectedAuditEvent(t)
			event.ProtectedMetadata = &svcmodel.ProtectedMetadata{
				Ciphertext: []byte("enc"),
				WrappedDEK: []byte("dek"),
				Commitment: []byte("cmt"),
			}

			protector := tc.protector
			if protector == nil {
				protector = &mockProtector{}
			}

			svc := NewQueryService(&mockTx{repos: tc.repos()}, deserializerReturning(event), protector, newTestLogger())

			result, err := svc.GetAuditEvent(context.Background(), uuid.New())

			if err != nil {
				t.Fatalf("expected no error on graceful degradation, got: %v", err)
			}
			if result.RevealedMetadata != nil {
				t.Error("expected nil RevealedMetadata on degradation")
			}
		})
	}
}

func TestQueryService_GetAuditEvent_Errors(t *testing.T) {
	tests := []struct {
		name    string
		repos   func() ports.Repositories
		wantErr error
	}{
		{
			name: "audit log entry not found",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return nil, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrAuditLogEntryNotFound,
		},
		{
			name: "deserialization fails",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return newTestAuditLogEntryRaw(1), nil
					},
				}
				return r
			},
			wantErr: svcerrors.ErrEventDeserializationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serializer := &mockSerializer{
				DeserializeProtectedAuditEventFunc: func(_ json.RawMessage) (*svcmodel.ProtectedAuditEvent, error) {
					return nil, errors.New("unmarshal error")
				},
			}

			svc := NewQueryService(&mockTx{repos: tc.repos()}, serializer, &mockProtector{}, newTestLogger())

			_, err := svc.GetAuditEvent(context.Background(), uuid.New())

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// --- ListAuditEvents ---

func TestQueryService_ListAuditEvents_Success(t *testing.T) {
	entries := []*svcmodel.AuditLogEntryRaw{
		newTestAuditLogEntryRaw(1),
		newTestAuditLogEntryRaw(2),
		newTestAuditLogEntryRaw(3),
	}
	event := newTestProtectedAuditEvent(t)

	repos := defaultQueryRepos()
	repos.AuditLog = &mockAuditLog{
		ListFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64, _ int) ([]*svcmodel.AuditLogEntryRaw, error) {
			return entries, nil
		},
	}

	svc := NewQueryService(&mockTx{repos: repos}, deserializerReturning(event), &mockProtector{}, newTestLogger())

	result, err := svc.ListAuditEvents(context.Background(), nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LedgerSize != 1 { // default mockLedger.Size returns 1
		t.Errorf("LedgerSize: got %d, want 1", result.LedgerSize)
	}
	if len(result.Items) != len(entries) {
		t.Errorf("Items count: got %d, want %d", len(result.Items), len(entries))
	}
	if result.NextCursor != nil {
		t.Errorf("NextCursor: expected nil, got %d", *result.NextCursor)
	}
}

func TestQueryService_ListAuditEvents_Pagination(t *testing.T) {
	entries := make([]*svcmodel.AuditLogEntryRaw, auditEventPageSize+1)
	for i := range entries {
		entries[i] = newTestAuditLogEntryRaw(int64(i + 1))
	}
	event := newTestProtectedAuditEvent(t)

	repos := defaultQueryRepos()
	repos.AuditLog = &mockAuditLog{
		ListFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64, _ int) ([]*svcmodel.AuditLogEntryRaw, error) {
			return entries, nil
		},
	}

	svc := NewQueryService(&mockTx{repos: repos}, deserializerReturning(event), &mockProtector{}, newTestLogger())

	result, err := svc.ListAuditEvents(context.Background(), nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != auditEventPageSize {
		t.Errorf("Items count: got %d, want %d", len(result.Items), auditEventPageSize)
	}
	if result.NextCursor == nil {
		t.Fatal("expected non-nil NextCursor")
	}
	if *result.NextCursor != *entries[auditEventPageSize-1].ID {
		t.Errorf("NextCursor: got %d, want %d", *result.NextCursor, *entries[auditEventPageSize-1].ID)
	}
}

func TestQueryService_ListAuditEvents_Errors(t *testing.T) {
	tests := []struct {
		name    string
		repos   func() ports.Repositories
		wantErr error
	}{
		{
			name: "ledger size fails",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.Ledger = &mockLedger{
					SizeFunc: func(_ context.Context) (int64, error) {
						return 0, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrLedgerSizeFailed,
		},
		{
			name: "list audit log entries fails",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.AuditLog = &mockAuditLog{
					ListFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64, _ int) ([]*svcmodel.AuditLogEntryRaw, error) {
						return nil, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrAuditLogEntryNotFound,
		},
		{
			name: "deserialization of an entry fails",
			repos: func() ports.Repositories {
				r := defaultQueryRepos()
				r.AuditLog = &mockAuditLog{
					ListFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64, _ int) ([]*svcmodel.AuditLogEntryRaw, error) {
						return []*svcmodel.AuditLogEntryRaw{newTestAuditLogEntryRaw(1)}, nil
					},
				}
				return r
			},
			wantErr: svcerrors.ErrEventDeserializationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serializer := &mockSerializer{
				DeserializeProtectedAuditEventFunc: func(_ json.RawMessage) (*svcmodel.ProtectedAuditEvent, error) {
					return nil, errors.New("unmarshal error")
				},
			}

			svc := NewQueryService(&mockTx{repos: tc.repos()}, serializer, &mockProtector{}, newTestLogger())

			_, err := svc.ListAuditEvents(context.Background(), nil, nil)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
