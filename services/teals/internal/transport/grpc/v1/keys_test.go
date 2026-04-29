package v1

import (
	"context"
	"testing"

	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
)

// --- mocks ---

type mockKeyService struct {
	RegisterProducerKeyFunc func(ctx context.Context, producerID uuid.UUID, key []byte) (string, error)
}

func (m *mockKeyService) RegisterProducerKey(ctx context.Context, producerID uuid.UUID, key []byte) (string, error) {
	if m.RegisterProducerKeyFunc != nil {
		return m.RegisterProducerKeyFunc(ctx, producerID, key)
	}
	return "", nil
}

// --- RegisterKey: input validation errors ---

func TestRegisterKey_InvalidProducerID_ReturnsInvalidArgument(t *testing.T) {
	s := NewKeyRegistrationServiceServer(&mockKeyService{})

	_, err := s.RegisterKey(context.Background(), &auditv1.RegisterKeyRequest{
		ProducerId: "not-a-uuid",
		PublicKey:  []byte("key"),
	})

	assertGRPCCode(t, err, codes.InvalidArgument)
}

// --- RegisterKey: ledgerService error mapping ---

func TestRegisterKey_ServiceErrors(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{
			name:     "invalid public key",
			svcErr:   svcerrors.ErrInvalidPublicKey,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "duplicate key",
			svcErr:   svcerrors.ErrDuplicateProducerKey,
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "producer not found",
			svcErr:   svcerrors.ErrProducerNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "unexpected internal error",
			svcErr:   svcerrors.ErrKeyRegistrationFailed,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockKeyService{
				RegisterProducerKeyFunc: func(_ context.Context, _ uuid.UUID, _ []byte) (string, error) {
					return "", tc.svcErr
				},
			}
			s := NewKeyRegistrationServiceServer(svc)

			_, err := s.RegisterKey(context.Background(), &auditv1.RegisterKeyRequest{
				ProducerId: validProducerID,
				PublicKey:  []byte("key"),
			})

			assertGRPCCode(t, err, tc.wantCode)
		})
	}
}

// --- RegisterKey: happy path ---

func TestRegisterKey_Success_ResponseContainsKeyID(t *testing.T) {
	const expectedKID = "kid-abc123"
	svc := &mockKeyService{
		RegisterProducerKeyFunc: func(_ context.Context, _ uuid.UUID, _ []byte) (string, error) {
			return expectedKID, nil
		},
	}
	s := NewKeyRegistrationServiceServer(svc)

	resp, err := s.RegisterKey(context.Background(), &auditv1.RegisterKeyRequest{
		ProducerId: validProducerID,
		PublicKey:  []byte("key"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.KeyId != expectedKID {
		t.Errorf("KeyId: got %q, want %q", resp.KeyId, expectedKID)
	}
}

func TestRegisterKey_Success_ProducerIDAndKeyForwardedToService(t *testing.T) {
	expectedID := uuid.MustParse(validProducerID)
	expectedKey := []byte("public-key-bytes")

	var capturedID uuid.UUID
	var capturedKey []byte

	svc := &mockKeyService{
		RegisterProducerKeyFunc: func(_ context.Context, id uuid.UUID, key []byte) (string, error) {
			capturedID = id
			capturedKey = key
			return "kid", nil
		},
	}
	s := NewKeyRegistrationServiceServer(svc)

	_, err := s.RegisterKey(context.Background(), &auditv1.RegisterKeyRequest{
		ProducerId: validProducerID,
		PublicKey:  expectedKey,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != expectedID {
		t.Errorf("producerID: got %v, want %v", capturedID, expectedID)
	}
	if string(capturedKey) != string(expectedKey) {
		t.Errorf("publicKey: got %v, want %v", capturedKey, expectedKey)
	}
}
