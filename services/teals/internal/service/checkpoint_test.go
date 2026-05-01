package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
)

// errTestStoreFailure is a sentinel for storage errors that CheckpointService propagates as-is.
var errTestStoreFailure = errors.New("store error")

// defaultCheckpointRepos returns a Repositories with succeeding mocks for Ledger and CheckpointStore.
func defaultCheckpointRepos() ports.Repositories {
	return ports.Repositories{
		Ledger:          &mockLedger{},
		CheckpointStore: &mockCheckpointStore{},
	}
}

// --- CreateCheckpoint: happy path ---

func TestCheckpointService_CreateCheckpoint_Success(t *testing.T) {
	const expectedKid = "server-kid-v1"
	expectedRoot := []byte("root")

	svc := NewCheckpointService(
		&mockTx{repos: defaultCheckpointRepos()},
		&mockCheckpointSigner{KidValue: expectedKid},
		newTestLogger(),
	)

	result, err := svc.CreateCheckpoint(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Checkpoint.Size != 1 { // default mockLedger.Size returns 1
		t.Errorf("Size: got %d, want 1", result.Checkpoint.Size)
	}
	if !bytes.Equal(result.Checkpoint.RootHash, expectedRoot) {
		t.Errorf("RootHash: got %x, want %x", result.Checkpoint.RootHash, expectedRoot)
	}
	if result.Kid != expectedKid {
		t.Errorf("Kid: got %q, want %q", result.Kid, expectedKid)
	}
	if result.SignatureToken == "" {
		t.Error("SignatureToken must not be empty")
	}
}

func TestCheckpointService_CreateCheckpoint_SkipsWhenLedgerUnchanged(t *testing.T) {
	repos := defaultCheckpointRepos()
	repos.CheckpointStore = &mockCheckpointStore{
		GetLatestFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
			return &svcmodel.SignedCheckpoint{
				Checkpoint: svcmodel.Checkpoint{Size: 1}, // same as default mockLedger.Size
			}, nil
		},
	}

	svc := NewCheckpointService(&mockTx{repos: repos}, &mockCheckpointSigner{}, newTestLogger())

	_, err := svc.CreateCheckpoint(context.Background())

	if !errors.Is(err, svcerrors.ErrCheckpointEmptyLedger) {
		t.Errorf("got %v, want ErrCheckpointEmptyLedger", err)
	}
}

// --- CreateCheckpoint: error paths ---

func TestCheckpointService_CreateCheckpoint_Errors(t *testing.T) {
	tests := []struct {
		name    string
		repos   func() ports.Repositories
		signer  *mockCheckpointSigner
		wantErr error
	}{
		{
			name: "ledger size fails",
			repos: func() ports.Repositories {
				r := defaultCheckpointRepos()
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
			name: "empty ledger",
			repos: func() ports.Repositories {
				r := defaultCheckpointRepos()
				r.Ledger = &mockLedger{
					SizeFunc: func(_ context.Context) (int64, error) { return 0, nil },
				}
				return r
			},
			wantErr: svcerrors.ErrCheckpointEmptyLedger,
		},
		{
			name: "get latest checkpoint fails",
			repos: func() ports.Repositories {
				r := defaultCheckpointRepos()
				r.CheckpointStore = &mockCheckpointStore{
					GetLatestFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
						return nil, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrGetCheckpointFailed,
		},
		{
			name: "root hash fails",
			repos: func() ports.Repositories {
				r := defaultCheckpointRepos()
				r.Ledger = &mockLedger{
					RootHashFunc: func(_ context.Context) ([]byte, error) {
						return nil, errors.New("db error")
					},
				}
				return r
			},
			wantErr: svcerrors.ErrGetCheckpointFailed,
		},
		{
			name: "signing fails",
			signer: &mockCheckpointSigner{
				SignFunc: func(_ []byte) (string, error) {
					return "", errors.New("key error")
				},
			},
			wantErr: svcerrors.ErrSignCheckpointFailed,
		},
		{
			name: "store checkpoint fails",
			repos: func() ports.Repositories {
				r := defaultCheckpointRepos()
				r.CheckpointStore = &mockCheckpointStore{
					StoreFunc: func(_ context.Context, _ *svcmodel.SignedCheckpoint) error {
						return errTestStoreFailure // propagated as-is
					},
				}
				return r
			},
			wantErr: errTestStoreFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repos := defaultCheckpointRepos()
			if tc.repos != nil {
				repos = tc.repos()
			}
			signer := tc.signer
			if signer == nil {
				signer = &mockCheckpointSigner{}
			}

			svc := NewCheckpointService(&mockTx{repos: repos}, signer, newTestLogger())

			_, err := svc.CreateCheckpoint(context.Background())

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// --- GetLatestCheckpoint  ---

func TestCheckpointService_GetLatestCheckpoint(t *testing.T) {
	stored := &svcmodel.SignedCheckpoint{Kid: "kid-1", SignatureToken: "tok-1"}

	tests := []struct {
		name      string
		storeFunc func(_ context.Context) (*svcmodel.SignedCheckpoint, error)
		wantErr   error
		wantNil   bool
	}{
		{
			name: "success",
			storeFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
				return stored, nil
			},
		},
		{
			name: "not found",
			storeFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
				return nil, svcerrors.ErrCheckpointNotFound
			},
			wantErr: svcerrors.ErrCheckpointNotFound,
			wantNil: true,
		},
		{
			name: "unexpected storage error",
			storeFunc: func(_ context.Context) (*svcmodel.SignedCheckpoint, error) {
				return nil, errTestStoreFailure
			},
			wantErr: errTestStoreFailure,
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repos := defaultCheckpointRepos()
			repos.CheckpointStore = &mockCheckpointStore{GetLatestFunc: tc.storeFunc}

			svc := NewCheckpointService(&mockTx{repos: repos}, &mockCheckpointSigner{}, newTestLogger())

			result, err := svc.GetLatestCheckpoint(context.Background())

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("error: got %v, want %v", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantNil && result != nil {
				t.Error("expected nil result")
			}
			if !tc.wantNil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

// --- ServerKid / ServerPublicKey ---

func TestCheckpointService_ServerKid(t *testing.T) {
	const kid = "server-key-v1"
	svc := NewCheckpointService(&mockTx{}, &mockCheckpointSigner{KidValue: kid}, newTestLogger())

	if got := svc.ServerKid(); got != kid {
		t.Errorf("got %q, want %q", got, kid)
	}
}

func TestCheckpointService_ServerPublicKey(t *testing.T) {
	pub := []byte("ed25519-public-key-bytes")
	svc := NewCheckpointService(&mockTx{}, &mockCheckpointSigner{PublicKeyValue: pub}, newTestLogger())

	if got := svc.ServerPublicKey(); !bytes.Equal(got, pub) {
		t.Errorf("got %x, want %x", got, pub)
	}
}
