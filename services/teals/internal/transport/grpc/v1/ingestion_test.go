package v1

import (
	"context"
	"testing"

	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/transport/grpc/interceptor"
	"google.golang.org/grpc/codes"
)

// --- mocks ---

type mockAuditService struct {
	IngestAuditEventFunc func(ctx context.Context, event *svcmodel.AuditEvent, signature string) (*svcmodel.IngestAuditEventResult, error)
}

func (m *mockAuditService) IngestAuditEvent(ctx context.Context, event *svcmodel.AuditEvent, signature string) (*svcmodel.IngestAuditEventResult, error) {
	if m.IngestAuditEventFunc != nil {
		return m.IngestAuditEventFunc(ctx, event, signature)
	}
	return nil, nil
}

// --- Append ---

func TestAppend_MissingSignatureInContext_ReturnsInternal(t *testing.T) {
	s := NewIngestionServiceServer(&mockAuditService{})

	_, err := s.Append(context.Background(), validAppendRequest())

	assertGRPCCode(t, err, codes.Internal)
}

func TestAppend_InvalidRequest_ReturnsInvalidArgument(t *testing.T) {
	s := NewIngestionServiceServer(&mockAuditService{})
	ctx := interceptor.ContextWithSignature(context.Background(), "test-token")

	invalidReq := validAppendRequest()
	invalidReq.Event.Id = "not-a-uuid"

	_, err := s.Append(ctx, invalidReq)

	assertGRPCCode(t, err, codes.InvalidArgument)
}

// --- Append: ledgerService error mapping ---

func TestAppend_ServiceErrors(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{
			name:     "duplicate event ID",
			svcErr:   svcerrors.ErrDuplicateEventID,
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "invalid signature",
			svcErr:   svcerrors.ErrInvalidSignature,
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "unexpected internal error",
			svcErr:   svcerrors.ErrLedgerAppendFailed,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		ctx := interceptor.ContextWithSignature(context.Background(), "test-token")

		t.Run(tc.name, func(t *testing.T) {
			svc := &mockAuditService{
				IngestAuditEventFunc: func(_ context.Context, _ *svcmodel.AuditEvent, _ string) (*svcmodel.IngestAuditEventResult, error) {
					return nil, tc.svcErr
				},
			}
			s := NewIngestionServiceServer(svc)

			_, err := s.Append(ctx, validAppendRequest())

			assertGRPCCode(t, err, tc.wantCode)
		})
	}
}

// --- Append: happy path ---

func TestAppend_Success_ResponseFieldsPopulated(t *testing.T) {
	ctx := interceptor.ContextWithSignature(context.Background(), "test-token")

	result := successResult()
	svc := &mockAuditService{
		IngestAuditEventFunc: func(_ context.Context, _ *svcmodel.AuditEvent, _ string) (*svcmodel.IngestAuditEventResult, error) {
			return result, nil
		},
	}
	s := NewIngestionServiceServer(svc)

	resp, err := s.Append(ctx, validAppendRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.EventId != result.EventID.String() {
		t.Errorf("EventId: got %q, want %q", resp.EventId, result.EventID.String())
	}
	if resp.LedgerSize != result.LedgerSize {
		t.Errorf("LedgerSize: got %d, want %d", resp.LedgerSize, result.LedgerSize)
	}
	if !resp.AppendedAt.AsTime().Equal(result.IngestedAt) {
		t.Errorf("AppendedAt: got %v, want %v", resp.AppendedAt.AsTime(), result.IngestedAt)
	}
}

func TestAppend_Success_SignatureTokenForwardedToService(t *testing.T) {
	const expectedToken = "jws-token-abc"
	var capturedToken string

	ctx := interceptor.ContextWithSignature(context.Background(), expectedToken)

	svc := &mockAuditService{
		IngestAuditEventFunc: func(_ context.Context, _ *svcmodel.AuditEvent, sig string) (*svcmodel.IngestAuditEventResult, error) {
			capturedToken = sig
			return successResult(), nil
		},
	}
	s := NewIngestionServiceServer(svc)

	_, err := s.Append(ctx, validAppendRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedToken != expectedToken {
		t.Errorf("signature token: got %q, want %q", capturedToken, expectedToken)
	}
}
