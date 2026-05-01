package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"

	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/google/uuid"
)

// newTestPublicKey generates a fresh ed25519 public key for use in tests.
func newTestPublicKey(t *testing.T) ed25519.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}
	return pub
}

// --- happy path ---

func TestKeyService_RegisterProducerKey_Success(t *testing.T) {
	pub := newTestPublicKey(t)
	producerID := uuid.New()

	var stored *svcmodel.ProducerKey
	svc := NewKeyService(
		&mockProducerKeyRegistry{
			AddPublicKeyFunc: func(_ context.Context, key *svcmodel.ProducerKey) error {
				stored = key
				return nil
			},
		},
		newTestLogger(),
	)

	kid, err := svc.RegisterProducerKey(context.Background(), producerID, pub)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kid == "" {
		t.Error("expected non-empty kid")
	}
	if stored == nil {
		t.Fatal("expected AddPublicKey to be called")
	}
	if stored.KeyID != kid {
		t.Errorf("stored KeyID %q does not match returned kid %q", stored.KeyID, kid)
	}
	if stored.ProducerID != producerID {
		t.Errorf("stored ProducerID %v, want %v", stored.ProducerID, producerID)
	}
}

// --- error paths ---

func TestKeyService_RegisterProducerKey_Errors(t *testing.T) {
	validPub := newTestPublicKey(t)

	tests := []struct {
		name       string
		key        []byte
		registryFn func(_ context.Context, _ *svcmodel.ProducerKey) error
		wantErr    error
	}{
		{
			name:    "empty key",
			key:     []byte{},
			wantErr: svcerrors.ErrInvalidPublicKey,
		},
		{
			name:    "key too short",
			key:     make([]byte, ed25519.PublicKeySize-1),
			wantErr: svcerrors.ErrInvalidPublicKey,
		},
		{
			name:    "key too long",
			key:     make([]byte, ed25519.PublicKeySize+1),
			wantErr: svcerrors.ErrInvalidPublicKey,
		},
		{
			name: "duplicate key",
			key:  validPub,
			registryFn: func(_ context.Context, _ *svcmodel.ProducerKey) error {
				return svcerrors.ErrDuplicateProducerKey
			},
			wantErr: svcerrors.ErrDuplicateProducerKey,
		},
		{
			name: "producer not found",
			key:  validPub,
			registryFn: func(_ context.Context, _ *svcmodel.ProducerKey) error {
				return svcerrors.ErrProducerNotFound
			},
			wantErr: svcerrors.ErrProducerNotFound,
		},
		{
			name: "unexpected registry error",
			key:  validPub,
			registryFn: func(_ context.Context, _ *svcmodel.ProducerKey) error {
				return errors.New("db error")
			},
			wantErr: svcerrors.ErrKeyRegistrationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewKeyService(
				&mockProducerKeyRegistry{AddPublicKeyFunc: tc.registryFn},
				newTestLogger(),
			)

			_, err := svc.RegisterProducerKey(context.Background(), uuid.New(), tc.key)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
