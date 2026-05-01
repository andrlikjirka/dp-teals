package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// errTestAuditLogFailure is a sentinel for audit log errors that LedgerService propagates as-is.
var errTestAuditLogFailure = errors.New("audit log db error")

// defaultLedgerRepos returns a Repositories with succeeding mocks.
func defaultLedgerRepos() ports.Repositories {
	return ports.Repositories{
		Ledger:   &mockLedger{},
		AuditLog: &mockAuditLog{},
	}
}

// --- GetInclusionProof ---

func TestLedgerService_GetInclusionProof_Success(t *testing.T) {
	eventID, _ := uuid.NewV7()
	entry := newTestAuditLogEntryRaw(0) // LeafIndex 0 < currentSize 1
	entry.EventID = eventID

	proofData := &svcmodel.InclusionProofData{
		LeafIndex:  0,
		LedgerSize: 1,
		LeafHash:   []byte("leaf-hash"),
		RootHash:   []byte("root-hash"),
		Proof:      &mmr.InclusionProof{},
	}

	repos := defaultLedgerRepos()
	repos.AuditLog = &mockAuditLog{
		GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
			return entry, nil
		},
	}
	repos.Ledger = &mockLedger{
		GenerateInclusionProofFunc: func(_ context.Context, _ int64, _ int64) (*svcmodel.InclusionProofData, error) {
			return proofData, nil
		},
	}

	svc := NewLedgerService(&mockTx{repos: repos}, newTestLogger())

	result, err := svc.GetInclusionProof(context.Background(), eventID, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventID != eventID {
		t.Errorf("EventID: got %v, want %v", result.EventID, eventID)
	}
	if result.LeafIndex != proofData.LeafIndex {
		t.Errorf("LeafIndex: got %d, want %d", result.LeafIndex, proofData.LeafIndex)
	}
	if !bytes.Equal(result.LeafEventHash, proofData.LeafHash) {
		t.Errorf("LeafEventHash: got %x, want %x", result.LeafEventHash, proofData.LeafHash)
	}
	if !bytes.Equal(result.RootHash, proofData.RootHash) {
		t.Errorf("RootHash: got %x, want %x", result.RootHash, proofData.RootHash)
	}
}

func TestLedgerService_GetInclusionProof_SizeResolution(t *testing.T) {
	tests := []struct {
		name          string
		requestedSize int64
		currentSize   int64
		wantSize      int64
	}{
		{
			name:          "zero uses current ledger size",
			requestedSize: 0,
			currentSize:   5,
			wantSize:      5,
		},
		{
			name:          "explicit size within range is used as-is",
			requestedSize: 3,
			currentSize:   5,
			wantSize:      3,
		},
		{
			name:          "size exceeding current is capped to current",
			requestedSize: 10,
			currentSize:   5,
			wantSize:      5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedSize int64

			entry := newTestAuditLogEntryRaw(0)

			repos := defaultLedgerRepos()
			repos.Ledger = &mockLedger{
				SizeFunc: func(_ context.Context) (int64, error) {
					return tc.currentSize, nil
				},
				GenerateInclusionProofFunc: func(_ context.Context, _ int64, size int64) (*svcmodel.InclusionProofData, error) {
					capturedSize = size
					return &svcmodel.InclusionProofData{Proof: &mmr.InclusionProof{}}, nil
				},
			}
			repos.AuditLog = &mockAuditLog{
				GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
					return entry, nil
				},
			}

			svc := NewLedgerService(&mockTx{repos: repos}, newTestLogger())
			_, err := svc.GetInclusionProof(context.Background(), uuid.New(), tc.requestedSize)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedSize != tc.wantSize {
				t.Errorf("size passed to GenerateInclusionProof: got %d, want %d", capturedSize, tc.wantSize)
			}
		})
	}
}

func TestLedgerService_GetInclusionProof_Errors(t *testing.T) {
	tests := []struct {
		name          string
		requestedSize int64
		repos         func() ports.Repositories
		wantErr       error
	}{
		{
			name:          "negative size rejected before tx",
			requestedSize: -1,
			wantErr:       svcerrors.ErrInvalidInclusionProofLedgerSize,
		},
		{
			name: "ledger size fails",
			repos: func() ports.Repositories {
				r := defaultLedgerRepos()
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
			name: "audit log entry not found",
			repos: func() ports.Repositories {
				r := defaultLedgerRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return nil, svcerrors.ErrAuditLogEntryNotFound
					},
				}
				return r
			},
			wantErr: svcerrors.ErrAuditLogEntryNotFound,
		},
		{
			name: "audit log lookup fails with unexpected error",
			repos: func() ports.Repositories {
				r := defaultLedgerRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return nil, errTestAuditLogFailure
					},
				}
				return r
			},
			wantErr: errTestAuditLogFailure,
		},
		{
			name: "leaf index exceeds resolved tree size",
			repos: func() ports.Repositories {
				r := defaultLedgerRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return newTestAuditLogEntryRaw(1), nil
					},
				}
				return r
			},
			wantErr: svcerrors.ErrInvalidInclusionProofLedgerSize,
		},
		{
			name: "inclusion proof generation fails",
			repos: func() ports.Repositories {
				r := defaultLedgerRepos()
				r.AuditLog = &mockAuditLog{
					GetFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.AuditLogEntryRaw, error) {
						return newTestAuditLogEntryRaw(0), nil
					},
				}
				r.Ledger = &mockLedger{
					GenerateInclusionProofFunc: func(_ context.Context, _ int64, _ int64) (*svcmodel.InclusionProofData, error) {
						return nil, errors.New("mmr error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrInclusionProofFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repos := defaultLedgerRepos()
			if tc.repos != nil {
				repos = tc.repos()
			}

			svc := NewLedgerService(&mockTx{repos: repos}, newTestLogger())

			_, err := svc.GetInclusionProof(context.Background(), uuid.New(), tc.requestedSize)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// --- GetConsistencyProof ---

func TestLedgerService_GetConsistencyProof_Success(t *testing.T) {
	proof := &mmr.ConsistencyProof{}

	repos := defaultLedgerRepos()
	repos.Ledger = &mockLedger{
		GenerateConsistencyProofFunc: func(_ context.Context, _ int64, _ int64) (*mmr.ConsistencyProof, error) {
			return proof, nil
		},
	}

	svc := NewLedgerService(&mockTx{repos: repos}, newTestLogger())

	result, err := svc.GetConsistencyProof(context.Background(), 1, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Proof != proof {
		t.Error("result.Proof does not match the proof returned by the ledger")
	}
}

func TestLedgerService_GetConsistencyProof_Errors(t *testing.T) {
	tests := []struct {
		name     string
		fromSize int64
		toSize   int64
		proofFn  func(_ context.Context, _, _ int64) (*mmr.ConsistencyProof, error)
		wantErr  error
	}{
		{
			name:     "negative fromSize rejected before tx",
			fromSize: -1,
			toSize:   5,
			wantErr:  svcerrors.ErrInvalidConsistencyProofRange,
		},
		{
			name:     "fromSize greater than toSize rejected before tx",
			fromSize: 5,
			toSize:   3,
			wantErr:  svcerrors.ErrInvalidConsistencyProofRange,
		},
		{
			name:     "ledger returns invalid range error",
			fromSize: 1,
			toSize:   5,
			proofFn: func(_ context.Context, _, _ int64) (*mmr.ConsistencyProof, error) {
				return nil, svcerrors.ErrInvalidConsistencyProofRange
			},
			wantErr: svcerrors.ErrInvalidConsistencyProofRange,
		},
		{
			name:     "ledger returns unexpected error",
			fromSize: 1,
			toSize:   5,
			proofFn: func(_ context.Context, _, _ int64) (*mmr.ConsistencyProof, error) {
				return nil, errors.New("mmr error")
			},
			wantErr: svcerrors.ErrConsistencyProofFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repos := defaultLedgerRepos()
			if tc.proofFn != nil {
				repos.Ledger = &mockLedger{GenerateConsistencyProofFunc: tc.proofFn}
			}

			svc := NewLedgerService(&mockTx{repos: repos}, newTestLogger())

			_, err := svc.GetConsistencyProof(context.Background(), tc.fromSize, tc.toSize)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
