package service

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// --- logger ---

func newTestLogger() *logger.Logger {
	return &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

// --- TransactionProvider ---

// mockTx is a simple implementation of the TransactionProvider interface.
type mockTx struct {
	repos ports.Repositories
	err   error
}

func (m *mockTx) Transact(_ context.Context, fn func(ports.Repositories) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(m.repos)
}

// --- Serializer ---

type mockSerializer struct {
	SerializeAuditEventFunc            func(*svcmodel.AuditEvent) (json.RawMessage, error)
	DeserializeAuditEventFunc          func(json.RawMessage) (*svcmodel.AuditEvent, error)
	SerializeProtectedAuditEventFunc   func(*svcmodel.ProtectedAuditEvent) (json.RawMessage, error)
	DeserializeProtectedAuditEventFunc func(json.RawMessage) (*svcmodel.ProtectedAuditEvent, error)
}

func (m *mockSerializer) SerializeCanonicalAuditEvent(e *svcmodel.AuditEvent) (json.RawMessage, error) {
	if m.SerializeAuditEventFunc != nil {
		return m.SerializeAuditEventFunc(e)
	}
	return json.RawMessage(`{}`), nil
}

func (m *mockSerializer) DeserializeCanonicalAuditEvent(d json.RawMessage) (*svcmodel.AuditEvent, error) {
	if m.DeserializeAuditEventFunc != nil {
		return m.DeserializeAuditEventFunc(d)
	}
	return nil, nil
}

func (m *mockSerializer) SerializeCanonicalProtectedAuditEvent(e *svcmodel.ProtectedAuditEvent) (json.RawMessage, error) {
	if m.SerializeProtectedAuditEventFunc != nil {
		return m.SerializeProtectedAuditEventFunc(e)
	}
	return json.RawMessage(`{}`), nil
}

func (m *mockSerializer) DeserializeCanonicalProtectedAuditEvent(d json.RawMessage) (*svcmodel.ProtectedAuditEvent, error) {
	if m.DeserializeProtectedAuditEventFunc != nil {
		return m.DeserializeProtectedAuditEventFunc(d)
	}
	return nil, nil
}

// --- SignatureVerifier ---

type mockVerifier struct {
	VerifyFunc func(ctx context.Context, token string, payload []byte) (string, error)
}

func (m *mockVerifier) Verify(ctx context.Context, token string, payload []byte) (string, error) {
	if m.VerifyFunc != nil {
		return m.VerifyFunc(ctx, token, payload)
	}
	return "kid-default", nil
}

// --- MetadataProtector ---

type mockProtector struct {
	ProtectFunc func(secret []byte, metadata map[string]any) (*svcmodel.ProtectedMetadata, []byte, error)
	RevealFunc  func(secret []byte, pm *svcmodel.ProtectedMetadata) (map[string]any, error)
}

func (m *mockProtector) Protect(secret []byte, metadata map[string]any) (*svcmodel.ProtectedMetadata, []byte, error) {
	if m.ProtectFunc != nil {
		return m.ProtectFunc(secret, metadata)
	}
	return &svcmodel.ProtectedMetadata{}, []byte("salt"), nil
}

func (m *mockProtector) Reveal(secret []byte, pm *svcmodel.ProtectedMetadata) (map[string]any, error) {
	if m.RevealFunc != nil {
		return m.RevealFunc(secret, pm)
	}
	return nil, nil
}

// --- AuditLog ---

type mockAuditLog struct {
	StoreFunc func(ctx context.Context, eventID uuid.UUID, payload json.RawMessage, sigToken string, producerKeyID uuid.UUID, nodeID int64, salt []byte) error
	GetFunc   func(ctx context.Context, eventID uuid.UUID) (*svcmodel.AuditLogEntryRaw, error)
	ListFunc  func(ctx context.Context, filter *svcmodel.AuditEventFilter, cursor *int64, size int) ([]*svcmodel.AuditLogEntryRaw, error)
}

func (m *mockAuditLog) StoreAuditLogEntry(ctx context.Context, eventID uuid.UUID, payload json.RawMessage, sigToken string, producerKeyID uuid.UUID, nodeID int64, salt []byte) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, eventID, payload, sigToken, producerKeyID, nodeID, salt)
	}
	return nil
}

func (m *mockAuditLog) GetAuditLogEntryByEventID(ctx context.Context, eventID uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *mockAuditLog) ListAuditLogEntries(ctx context.Context, filter *svcmodel.AuditEventFilter, cursor *int64, size int) ([]*svcmodel.AuditLogEntryRaw, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter, cursor, size)
	}
	return nil, nil
}

// --- ProducerKeyRegistry ---

type mockProducerKeyRegistry struct {
	AddPublicKeyFunc        func(ctx context.Context, key *svcmodel.ProducerKey) error
	PublicKeyFunc           func(ctx context.Context, kid string) (ed25519.PublicKey, error)
	RevokeKeyFunc           func(ctx context.Context, kid string) error
	GetProducerKeyByKidFunc func(ctx context.Context, kid string) (*svcmodel.ProducerKey, error)
}

func (m *mockProducerKeyRegistry) AddPublicKey(ctx context.Context, key *svcmodel.ProducerKey) error {
	if m.AddPublicKeyFunc != nil {
		return m.AddPublicKeyFunc(ctx, key)
	}
	return nil
}

func (m *mockProducerKeyRegistry) PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	if m.PublicKeyFunc != nil {
		return m.PublicKeyFunc(ctx, kid)
	}
	return nil, nil
}

func (m *mockProducerKeyRegistry) RevokeKey(ctx context.Context, kid string) error {
	if m.RevokeKeyFunc != nil {
		return m.RevokeKeyFunc(ctx, kid)
	}
	return nil
}

func (m *mockProducerKeyRegistry) GetProducerKeyByKid(ctx context.Context, kid string) (*svcmodel.ProducerKey, error) {
	if m.GetProducerKeyByKidFunc != nil {
		return m.GetProducerKeyByKidFunc(ctx, kid)
	}
	return &svcmodel.ProducerKey{ID: uuid.New()}, nil
}

// --- Ledger ---

type mockLedger struct {
	AppendLeafFunc               func(ctx context.Context, payload []byte) (int64, int64, error)
	SizeFunc                     func(ctx context.Context) (int64, error)
	RootHashFunc                 func(ctx context.Context) ([]byte, error)
	GenerateInclusionProofFunc   func(ctx context.Context, leafIndex int64, size int64) (*svcmodel.InclusionProofData, error)
	GenerateConsistencyProofFunc func(ctx context.Context, fromSize int64, toSize int64) (*mmr.ConsistencyProof, error)
}

func (m *mockLedger) AppendLeaf(ctx context.Context, payload []byte) (int64, int64, error) {
	if m.AppendLeafFunc != nil {
		return m.AppendLeafFunc(ctx, payload)
	}
	return 0, 1, nil
}

func (m *mockLedger) Size(ctx context.Context) (int64, error) {
	if m.SizeFunc != nil {
		return m.SizeFunc(ctx)
	}
	return 1, nil
}

func (m *mockLedger) RootHash(ctx context.Context) ([]byte, error) {
	if m.RootHashFunc != nil {
		return m.RootHashFunc(ctx)
	}
	return []byte("root"), nil
}

func (m *mockLedger) GenerateInclusionProof(ctx context.Context, leafIndex int64, size int64) (*svcmodel.InclusionProofData, error) {
	if m.GenerateInclusionProofFunc != nil {
		return m.GenerateInclusionProofFunc(ctx, leafIndex, size)
	}
	return nil, nil
}

func (m *mockLedger) GenerateConsistencyProof(ctx context.Context, fromSize int64, toSize int64) (*mmr.ConsistencyProof, error) {
	if m.GenerateConsistencyProofFunc != nil {
		return m.GenerateConsistencyProofFunc(ctx, fromSize, toSize)
	}
	return nil, nil
}

// --- SubjectSecretStore ---

type mockSubjectSecretStore struct {
	GetOrCreateSecretFunc       func(ctx context.Context, subjectID string) ([]byte, error)
	GetSecretBySubjectIDFunc    func(ctx context.Context, subjectID string) ([]byte, error)
	DeleteSecretBySubjectIDFunc func(ctx context.Context, subjectID string) error
}

func (m *mockSubjectSecretStore) GetOrCreateSecret(ctx context.Context, subjectID string) ([]byte, error) {
	if m.GetOrCreateSecretFunc != nil {
		return m.GetOrCreateSecretFunc(ctx, subjectID)
	}
	return []byte("secret"), nil
}

func (m *mockSubjectSecretStore) GetSecretBySubjectId(ctx context.Context, subjectID string) ([]byte, error) {
	if m.GetSecretBySubjectIDFunc != nil {
		return m.GetSecretBySubjectIDFunc(ctx, subjectID)
	}
	return []byte("secret"), nil
}

func (m *mockSubjectSecretStore) DeleteSecretBySubjectId(ctx context.Context, subjectID string) error {
	if m.DeleteSecretBySubjectIDFunc != nil {
		return m.DeleteSecretBySubjectIDFunc(ctx, subjectID)
	}
	return nil
}

// --- CheckpointStore ---

type mockCheckpointStore struct {
	StoreFunc     func(ctx context.Context, checkpoint *svcmodel.SignedCheckpoint) error
	GetLatestFunc func(ctx context.Context) (*svcmodel.SignedCheckpoint, error)
}

func (m *mockCheckpointStore) StoreCheckpoint(ctx context.Context, checkpoint *svcmodel.SignedCheckpoint) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, checkpoint)
	}
	return nil
}

func (m *mockCheckpointStore) GetLatestSignedCheckpoint(ctx context.Context) (*svcmodel.SignedCheckpoint, error) {
	if m.GetLatestFunc != nil {
		return m.GetLatestFunc(ctx)
	}
	return nil, nil
}

// --- CheckpointSigner ---

type mockCheckpointSigner struct {
	SignFunc       func(payload []byte) (string, error)
	KidValue       string
	PublicKeyValue []byte
}

func (m *mockCheckpointSigner) Sign(payload []byte) (string, error) {
	if m.SignFunc != nil {
		return m.SignFunc(payload)
	}
	return "sig-token", nil
}

func (m *mockCheckpointSigner) Kid() string {
	return m.KidValue
}

func (m *mockCheckpointSigner) PublicKey() []byte {
	return m.PublicKeyValue
}
