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

// errTestVaultFailure is a sentinel used in tests that expect a raw storage error to be propagated.
var errTestVaultFailure = errors.New("vault error")

// newTestAuditEvent builds a minimal valid AuditEvent for use in tests.
func newTestAuditEvent(t *testing.T) *svcmodel.AuditEvent {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("generate uuid: %v", err)
	}
	e, err := svcmodel.NewAuditEvent(svcmodel.CreateAuditEventParams{
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
		t.Fatalf("NewAuditEvent: %v", err)
	}
	return e
}

// defaultRepos returns a Repositories with succeeding mocks for all deps used by AuditService.
func defaultRepos() ports.Repositories {
	return ports.Repositories{
		ProducerKeys:       &mockProducerKeyRegistry{},
		SubjectSecretStore: &mockSubjectSecretStore{},
		Ledger:             &mockLedger{},
		AuditLog:           &mockAuditLog{},
	}
}

// --- happy paths ---

func TestAuditService_IngestAuditEvent_SuccessWithoutMetadata(t *testing.T) {
	event := newTestAuditEvent(t)

	protectorCalled := false
	svc := NewAuditService(
		&mockTx{repos: defaultRepos()},
		&mockSerializer{},
		&mockVerifier{},
		&mockProtector{ProtectFunc: func(_ []byte, _ map[string]any) (*svcmodel.ProtectedMetadata, []byte, error) {
			protectorCalled = true
			return nil, nil, nil
		}},
		newTestLogger(),
	)

	result, err := svc.IngestAuditEvent(context.Background(), event, "sig-token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventID != event.ID {
		t.Errorf("EventID: got %v, want %v", result.EventID, event.ID)
	}
	if protectorCalled {
		t.Error("protector.Protect must NOT be called when event has no metadata")
	}
}

func TestAuditService_IngestAuditEvent_SuccessWithMetadata(t *testing.T) {
	event := newTestAuditEvent(t)
	event.Metadata = map[string]any{"email": "user@example.com"}

	protectorCalled := false
	svc := NewAuditService(
		&mockTx{repos: defaultRepos()},
		&mockSerializer{},
		&mockVerifier{},
		&mockProtector{ProtectFunc: func(_ []byte, _ map[string]any) (*svcmodel.ProtectedMetadata, []byte, error) {
			protectorCalled = true
			return &svcmodel.ProtectedMetadata{}, []byte("salt"), nil
		}},
		newTestLogger(),
	)

	_, err := svc.IngestAuditEvent(context.Background(), event, "sig-token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !protectorCalled {
		t.Error("protector.Protect must be called when event has metadata")
	}
}

// --- error paths ---

func TestAuditService_IngestAuditEvent_Errors(t *testing.T) {
	tests := []struct {
		name       string
		event      func(t *testing.T) *svcmodel.AuditEvent
		serializer *mockSerializer
		verifier   *mockVerifier
		protector  *mockProtector
		repos      func() ports.Repositories
		wantErr    error
	}{
		{
			name: "audit event serialization fails",
			serializer: &mockSerializer{
				SerializeAuditEventFunc: func(_ *svcmodel.AuditEvent) (json.RawMessage, error) {
					return nil, errors.New("marshal error")
				},
			},
			wantErr: svcerrors.ErrEventSerializationFailed,
		},
		{
			name: "signature verification fails",
			verifier: &mockVerifier{
				VerifyFunc: func(_ context.Context, _ string, _ []byte) (string, error) {
					return "", errors.New("bad signature")
				},
			},
			wantErr: svcerrors.ErrInvalidSignature,
		},
		{
			name: "producer key retrieval fails",
			repos: func() ports.Repositories {
				r := defaultRepos()
				r.ProducerKeys = &mockProducerKeyRegistry{
					GetProducerKeyByKidFunc: func(_ context.Context, _ string) (*svcmodel.ProducerKey, error) {
						return nil, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrProducerKeyRetrievalFailed,
		},
		{
			name: "subject secret retrieval fails",
			event: func(t *testing.T) *svcmodel.AuditEvent {
				e := newTestAuditEvent(t)
				e.Metadata = map[string]any{"key": "value"}
				return e
			},
			repos: func() ports.Repositories {
				r := defaultRepos()
				r.SubjectSecretStore = &mockSubjectSecretStore{
					GetOrCreateSecretFunc: func(_ context.Context, _ string) ([]byte, error) {
						return nil, errTestVaultFailure
					},
				}
				return r
			},
			wantErr: errTestVaultFailure,
		},
		{
			name: "metadata protection fails",
			event: func(t *testing.T) *svcmodel.AuditEvent {
				e := newTestAuditEvent(t)
				e.Metadata = map[string]any{"key": "value"}
				return e
			},
			protector: &mockProtector{
				ProtectFunc: func(_ []byte, _ map[string]any) (*svcmodel.ProtectedMetadata, []byte, error) {
					return nil, nil, errors.New("encrypt error")
				},
			},
			wantErr: svcerrors.ErrProtectionFailed,
		},
		{
			name: "protected event serialization fails",
			serializer: &mockSerializer{
				SerializeProtectedAuditEventFunc: func(_ *svcmodel.ProtectedAuditEvent) (json.RawMessage, error) {
					return nil, errors.New("marshal error")
				},
			},
			wantErr: svcerrors.ErrEventSerializationFailed,
		},
		{
			name: "ledger append fails",
			repos: func() ports.Repositories {
				r := defaultRepos()
				r.Ledger = &mockLedger{
					AppendLeafFunc: func(_ context.Context, _ []byte) (int64, int64, error) {
						return 0, 0, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrLedgerAppendFailed,
		},
		{
			name: "duplicate event ID",
			repos: func() ports.Repositories {
				r := defaultRepos()
				r.AuditLog = &mockAuditLog{
					StoreFunc: func(_ context.Context, _ uuid.UUID, _ json.RawMessage, _ string, _ uuid.UUID, _ int64, _ []byte) error {
						return svcerrors.ErrDuplicateEventID
					},
				}
				return r
			},
			wantErr: svcerrors.ErrDuplicateEventID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Defaults — overridden by any non-nil field in the test case.
			serializer := tc.serializer
			if serializer == nil {
				serializer = &mockSerializer{}
			}
			verifier := tc.verifier
			if verifier == nil {
				verifier = &mockVerifier{}
			}
			protector := tc.protector
			if protector == nil {
				protector = &mockProtector{}
			}
			repos := defaultRepos()
			if tc.repos != nil {
				repos = tc.repos()
			}

			event := newTestAuditEvent(t)
			if tc.event != nil {
				event = tc.event(t)
			}

			svc := NewAuditService(&mockTx{repos: repos}, serializer, verifier, protector, newTestLogger())

			_, err := svc.IngestAuditEvent(context.Background(), event, "sig-token")

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
