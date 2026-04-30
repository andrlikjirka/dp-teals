package v1

import (
	"context"
	"testing"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

// --- mocks ---

type mockLedgerProver struct {
	GetInclusionProofFunc   func(ctx context.Context, eventID uuid.UUID, size int64) (*svcmodel.InclusionProofResult, error)
	GetConsistencyProofFunc func(ctx context.Context, fromSize int64, toSize int64) (*svcmodel.ConsistencyProofResult, error)
}

func (m *mockLedgerProver) GetInclusionProof(ctx context.Context, eventID uuid.UUID, size int64) (*svcmodel.InclusionProofResult, error) {
	if m.GetInclusionProofFunc != nil {
		return m.GetInclusionProofFunc(ctx, eventID, size)
	}
	return nil, nil
}

func (m *mockLedgerProver) GetConsistencyProof(ctx context.Context, fromSize int64, toSize int64) (*svcmodel.ConsistencyProofResult, error) {
	if m.GetConsistencyProofFunc != nil {
		return m.GetConsistencyProofFunc(ctx, fromSize, toSize)
	}
	return nil, nil
}

type mockCheckpointProvider struct {
	GetLatestCheckpointFunc func(ctx context.Context) (*svcmodel.SignedCheckpoint, error)
	ServerPublicKeyVal      []byte
	ServerKidVal            string
}

func (m *mockCheckpointProvider) GetLatestCheckpoint(ctx context.Context) (*svcmodel.SignedCheckpoint, error) {
	if m.GetLatestCheckpointFunc != nil {
		return m.GetLatestCheckpointFunc(ctx)
	}
	return nil, nil
}

func (m *mockCheckpointProvider) ServerPublicKey() []byte { return m.ServerPublicKeyVal }
func (m *mockCheckpointProvider) ServerKid() string       { return m.ServerKidVal }

// --- GetInclusionProof ---

func TestGetInclusionProof_InvalidEventID_ReturnsInvalidArgument(t *testing.T) {
	s := NewProofServiceServer(&mockLedgerProver{}, &mockCheckpointProvider{})
	_, err := s.GetInclusionProof(context.Background(), &auditv1.GetInclusionProofRequest{
		EventId: "not-a-uuid",
	})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestGetInclusionProof_ServiceErrors(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{
			name:     "event not found",
			svcErr:   svcerrors.ErrAuditLogEntryNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "invalid ledger size",
			svcErr:   svcerrors.ErrInvalidInclusionProofLedgerSize,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "unexpected internal error",
			svcErr:   svcerrors.ErrInclusionProofFailed,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockLedgerProver{
				GetInclusionProofFunc: func(_ context.Context, _ uuid.UUID, _ int64) (*svcmodel.InclusionProofResult, error) {
					return nil, tc.svcErr
				},
			}
			s := NewProofServiceServer(svc, &mockCheckpointProvider{})

			_, err := s.GetInclusionProof(context.Background(), &auditv1.GetInclusionProofRequest{
				EventId: validProducerID,
			})
			assertGRPCCode(t, err, tc.wantCode)
		})
	}
}

func TestGetInclusionProof_Success_ResponseFieldsPopulated(t *testing.T) {
	eventID := uuid.MustParse(validProducerID)
	result := &svcmodel.InclusionProofResult{
		EventID:       eventID,
		LeafIndex:     3,
		LeafEventHash: []byte("leaf-hash"),
		RootHash:      []byte("root-hash"),
		LedgerSize:    10,
		Proof:         &mmr.InclusionProof{Siblings: [][]byte{[]byte("sib")}, Left: []bool{true}},
	}

	svc := &mockLedgerProver{
		GetInclusionProofFunc: func(_ context.Context, _ uuid.UUID, _ int64) (*svcmodel.InclusionProofResult, error) {
			return result, nil
		},
	}
	s := NewProofServiceServer(svc, &mockCheckpointProvider{})

	resp, err := s.GetInclusionProof(context.Background(), &auditv1.GetInclusionProofRequest{
		EventId: validProducerID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.EventId != result.EventID.String() {
		t.Errorf("EventId: got %q, want %q", resp.EventId, result.EventID.String())
	}
	if resp.LeafIndex != result.LeafIndex {
		t.Errorf("LeafIndex: got %d, want %d", resp.LeafIndex, result.LeafIndex)
	}
	if resp.LedgerSize != result.LedgerSize {
		t.Errorf("LedgerSize: got %d, want %d", resp.LedgerSize, result.LedgerSize)
	}
	if string(resp.LeafHash) != string(result.LeafEventHash) {
		t.Errorf("LeafHash: got %v, want %v", resp.LeafHash, result.LeafEventHash)
	}
	if string(resp.RootHash) != string(result.RootHash) {
		t.Errorf("RootHash: got %v, want %v", resp.RootHash, result.RootHash)
	}
	if resp.Proof == nil {
		t.Fatal("expected Proof in response")
	}
	if len(resp.Proof.Siblings) != 1 {
		t.Errorf("Proof.Siblings length: got %d, want 1", len(resp.Proof.Siblings))
	}
}

func TestGetInclusionProof_LedgerSizeForwardedToService(t *testing.T) {
	expectedSize := int64(42)

	var capturedSize int64
	svc := &mockLedgerProver{
		GetInclusionProofFunc: func(_ context.Context, _ uuid.UUID, size int64) (*svcmodel.InclusionProofResult, error) {
			capturedSize = size
			return &svcmodel.InclusionProofResult{
				EventID: uuid.MustParse(validProducerID),
				Proof:   &mmr.InclusionProof{},
			}, nil
		},
	}
	s := NewProofServiceServer(svc, &mockCheckpointProvider{})

	_, err := s.GetInclusionProof(context.Background(), &auditv1.GetInclusionProofRequest{
		EventId:    validProducerID,
		LedgerSize: expectedSize,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSize != expectedSize {
		t.Errorf("LedgerSize forwarded: got %d, want %d", capturedSize, expectedSize)
	}
}

// --- GetConsistencyProof ---

func TestGetConsistencyProof_ServiceError_ReturnsInternal(t *testing.T) {
	svc := &mockLedgerProver{
		GetConsistencyProofFunc: func(_ context.Context, _, _ int64) (*svcmodel.ConsistencyProofResult, error) {
			return nil, svcerrors.ErrConsistencyProofFailed
		},
	}
	s := NewProofServiceServer(svc, &mockCheckpointProvider{})

	_, err := s.GetConsistencyProof(context.Background(), &auditv1.GetConsistencyProofRequest{
		FromSize: 1, ToSize: 10,
	})
	assertGRPCCode(t, err, codes.Internal)
}

func TestGetConsistencyProof_Success_ResponseFieldsPopulated(t *testing.T) {
	result := &svcmodel.ConsistencyProofResult{
		Proof: &mmr.ConsistencyProof{
			OldSize:        2,
			NewSize:        5,
			OldPeaksHashes: [][]byte{[]byte("old-peak")},
			ConsistencyPaths: []*mmr.ConsistencyPath{
				{Siblings: [][]byte{[]byte("sib")}, Left: []bool{false}},
			},
			RightPeaks: [][]byte{[]byte("right-peak")},
		},
	}

	svc := &mockLedgerProver{
		GetConsistencyProofFunc: func(_ context.Context, _, _ int64) (*svcmodel.ConsistencyProofResult, error) {
			return result, nil
		},
	}
	s := NewProofServiceServer(svc, &mockCheckpointProvider{})

	resp, err := s.GetConsistencyProof(context.Background(), &auditv1.GetConsistencyProofRequest{
		FromSize: 2, ToSize: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := resp.Proof
	if p == nil {
		t.Fatal("expected Proof in response")
	}
	if p.OldSize != int64(result.Proof.OldSize) {
		t.Errorf("OldSize: got %d, want %d", p.OldSize, result.Proof.OldSize)
	}
	if p.NewSize != int64(result.Proof.NewSize) {
		t.Errorf("NewSize: got %d, want %d", p.NewSize, result.Proof.NewSize)
	}
	if len(p.ConsistencyPaths) != 1 {
		t.Errorf("ConsistencyPaths length: got %d, want 1", len(p.ConsistencyPaths))
	}
	if len(p.OldPeaksHashes) != 1 {
		t.Errorf("OldPeaksHashes length: got %d, want 1", len(p.OldPeaksHashes))
	}
	if len(p.RightPeaks) != 1 {
		t.Errorf("RightPeaks length: got %d, want 1", len(p.RightPeaks))
	}
}

func TestGetConsistencyProof_FromAndToSizeForwardedToService(t *testing.T) {
	expectedFromSize, expectedToSize := int64(3), int64(7)

	var capturedFrom, capturedTo int64
	svc := &mockLedgerProver{
		GetConsistencyProofFunc: func(_ context.Context, from, to int64) (*svcmodel.ConsistencyProofResult, error) {
			capturedFrom, capturedTo = from, to
			return &svcmodel.ConsistencyProofResult{Proof: &mmr.ConsistencyProof{}}, nil
		},
	}
	s := NewProofServiceServer(svc, &mockCheckpointProvider{})

	_, err := s.GetConsistencyProof(context.Background(), &auditv1.GetConsistencyProofRequest{
		FromSize: expectedFromSize, ToSize: expectedToSize,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFrom != expectedFromSize || capturedTo != expectedToSize {
		t.Errorf("sizes forwarded: got from=%d to=%d, want from=%d to=%d", capturedFrom, capturedTo, expectedFromSize, expectedToSize)
	}
}

// --- GetLatestSignedCheckpoint ---

func TestGetLatestSignedCheckpoint_ServiceErrors(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{
			name:     "no checkpoint exists",
			svcErr:   svcerrors.ErrCheckpointNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "unexpected internal error",
			svcErr:   svcerrors.ErrGetCheckpointFailed,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs := &mockCheckpointProvider{
				GetLatestCheckpointFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
					return nil, tc.svcErr
				},
			}
			s := NewProofServiceServer(&mockLedgerProver{}, cs)

			_, err := s.GetLatestSignedCheckpoint(context.Background(), &auditv1.GetLatestSignedCheckpointRequest{})
			assertGRPCCode(t, err, tc.wantCode)
		})
	}
}

func TestGetLatestSignedCheckpoint_Success_ResponseFieldsPopulated(t *testing.T) {
	anchoredAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	checkpoint := &svcmodel.SignedCheckpoint{
		ID: uuid.MustParse(validProducerID),
		Checkpoint: svcmodel.Checkpoint{
			Size:       100,
			RootHash:   []byte("root"),
			AnchoredAt: anchoredAt,
		},
		Kid:            "kid-xyz",
		SignatureToken: "sig-token",
	}

	cs := &mockCheckpointProvider{
		GetLatestCheckpointFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
			return checkpoint, nil
		},
	}
	s := NewProofServiceServer(&mockLedgerProver{}, cs)

	resp, err := s.GetLatestSignedCheckpoint(context.Background(), &auditv1.GetLatestSignedCheckpointRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Checkpoint == nil {
		t.Fatal("expected Checkpoint in response")
	}
	if resp.Checkpoint.Id != checkpoint.ID.String() {
		t.Errorf("Id: got %q, want %q", resp.Checkpoint.Id, checkpoint.ID.String())
	}
	if resp.Checkpoint.Payload.Size != checkpoint.Checkpoint.Size {
		t.Errorf("Size: got %d, want %d", resp.Checkpoint.Payload.Size, checkpoint.Checkpoint.Size)
	}
	if string(resp.Checkpoint.Payload.RootHash) != string(checkpoint.Checkpoint.RootHash) {
		t.Errorf("RootHash: got %v, want %v", resp.Checkpoint.Payload.RootHash, checkpoint.Checkpoint.RootHash)
	}
	if resp.Signature.Kid != checkpoint.Kid {
		t.Errorf("Kid: got %q, want %q", resp.Signature.Kid, checkpoint.Kid)
	}
	if resp.Signature.SignatureToken != checkpoint.SignatureToken {
		t.Errorf("SignatureToken: got %q, want %q", resp.Signature.SignatureToken, checkpoint.SignatureToken)
	}
}

// --- GetServerPublicKey ---

func TestGetServerPublicKey_ReturnsPublicKeyAndKid(t *testing.T) {
	cs := &mockCheckpointProvider{
		ServerPublicKeyVal: []byte("public-key-bytes"),
		ServerKidVal:       "kid-server",
	}
	s := NewProofServiceServer(&mockLedgerProver{}, cs)

	resp, err := s.GetServerPublicKey(context.Background(), &auditv1.GetServerPublicKeyRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.PublicKey) != string(cs.ServerPublicKeyVal) {
		t.Errorf("PublicKey: got %v, want %v", resp.PublicKey, cs.ServerPublicKeyVal)
	}
	if resp.Kid != cs.ServerKidVal {
		t.Errorf("Kid: got %q, want %q", resp.Kid, cs.ServerKidVal)
	}
}
