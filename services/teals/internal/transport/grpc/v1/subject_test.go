package v1

import (
	"context"
	"testing"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"google.golang.org/grpc/codes"
)

// --- mock ---

type mockSubjectService struct {
	ForgetSubjectFunc func(ctx context.Context, subjectID string) (svcmodel.ForgetSubjectResult, error)
}

func (m *mockSubjectService) ForgetSubject(ctx context.Context, subjectID string) (svcmodel.ForgetSubjectResult, error) {
	if m.ForgetSubjectFunc != nil {
		return m.ForgetSubjectFunc(ctx, subjectID)
	}
	return svcmodel.ForgetSubjectResult{}, nil
}

// --- ForgetSubject: service error mapping ---

func TestForgetSubject_ServiceErrors(t *testing.T) {
	tests := []struct {
		name     string
		svcErr   error
		wantCode codes.Code
	}{
		{
			name:     "missing subject ID",
			svcErr:   svcerrors.ErrMissingSubjectID,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "subject secret not found",
			svcErr:   svcerrors.ErrSubjectSecretNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "unexpected internal error",
			svcErr:   svcerrors.ErrSubjectSecretDeletionFailed,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockSubjectService{
				ForgetSubjectFunc: func(_ context.Context, _ string) (svcmodel.ForgetSubjectResult, error) {
					return svcmodel.ForgetSubjectResult{}, tc.svcErr
				},
			}
			s := NewDataSubjectServiceServer(svc)

			_, err := s.ForgetSubject(context.Background(), &auditv1.ForgetSubjectRequest{
				SubjectId: "subject-1",
			})

			assertGRPCCode(t, err, tc.wantCode)
		})
	}
}

// --- ForgetSubject: happy path ---

func TestForgetSubject_Success_ResponseFieldsPopulated(t *testing.T) {
	const expectedSubjectID = "subject-42"
	expectedForgottenAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	svc := &mockSubjectService{
		ForgetSubjectFunc: func(_ context.Context, _ string) (svcmodel.ForgetSubjectResult, error) {
			return svcmodel.ForgetSubjectResult{
				SubjectID:   expectedSubjectID,
				ForgottenAt: expectedForgottenAt,
			}, nil
		},
	}
	s := NewDataSubjectServiceServer(svc)

	resp, err := s.ForgetSubject(context.Background(), &auditv1.ForgetSubjectRequest{
		SubjectId: expectedSubjectID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SubjectId != expectedSubjectID {
		t.Errorf("SubjectId: got %q, want %q", resp.SubjectId, expectedSubjectID)
	}
	if !resp.ForgottenAt.AsTime().Equal(expectedForgottenAt) {
		t.Errorf("ForgottenAt: got %v, want %v", resp.ForgottenAt.AsTime(), expectedForgottenAt)
	}
}

func TestForgetSubject_Success_SubjectIDForwardedToService(t *testing.T) {
	const expectedSubjectID = "subject-42"
	var capturedSubjectID string

	svc := &mockSubjectService{
		ForgetSubjectFunc: func(_ context.Context, subjectID string) (svcmodel.ForgetSubjectResult, error) {
			capturedSubjectID = subjectID
			return svcmodel.ForgetSubjectResult{SubjectID: subjectID, ForgottenAt: time.Now()}, nil
		},
	}
	s := NewDataSubjectServiceServer(svc)

	_, err := s.ForgetSubject(context.Background(), &auditv1.ForgetSubjectRequest{
		SubjectId: expectedSubjectID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSubjectID != expectedSubjectID {
		t.Errorf("subjectID forwarded: got %q, want %q", capturedSubjectID, expectedSubjectID)
	}
}
