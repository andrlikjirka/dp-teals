package service

import (
	"context"
	"errors"
	"testing"

	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
)

// --- happy path ---

func TestSubjectService_ForgetSubject_Success(t *testing.T) {
	const subjectID = "subject-1"

	repos := defaultRepos()
	svc := NewSubjectService(&mockTx{repos: repos}, newTestLogger())

	result, err := svc.ForgetSubject(context.Background(), subjectID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SubjectID != subjectID {
		t.Errorf("SubjectID: got %q, want %q", result.SubjectID, subjectID)
	}
	if result.ForgottenAt.IsZero() {
		t.Error("ForgottenAt must not be zero")
	}
}

// --- error paths  ---

func TestSubjectService_ForgetSubject_Errors(t *testing.T) {
	tests := []struct {
		name       string
		subjectID  string
		deleteFunc func(_ context.Context, _ string) error
		wantErr    error
	}{
		{
			name:      "empty subject ID",
			subjectID: "",
			wantErr:   svcerrors.ErrMissingSubjectID,
		},
		{
			name:      "subject secret not found",
			subjectID: "subject-1",
			deleteFunc: func(_ context.Context, _ string) error {
				return svcerrors.ErrSubjectSecretNotFound
			},
			wantErr: svcerrors.ErrSubjectSecretNotFound,
		},
		{
			name:      "unexpected deletion error",
			subjectID: "subject-1",
			deleteFunc: func(_ context.Context, _ string) error {
				return errors.New("db error")
			},
			wantErr: svcerrors.ErrSubjectSecretDeletionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repos := defaultRepos()
			repos.SubjectSecretStore = &mockSubjectSecretStore{
				DeleteSecretBySubjectIDFunc: tc.deleteFunc,
			}

			svc := NewSubjectService(&mockTx{repos: repos}, newTestLogger())

			_, err := svc.ForgetSubject(context.Background(), tc.subjectID)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSubjectService_ForgetSubject_EmptyID_DoesNotCallRegistry(t *testing.T) {
	deleteCalled := false
	repos := defaultRepos()
	repos.SubjectSecretStore = &mockSubjectSecretStore{
		DeleteSecretBySubjectIDFunc: func(_ context.Context, _ string) error {
			deleteCalled = true
			return nil
		},
	}

	svc := NewSubjectService(&mockTx{repos: repos}, newTestLogger())
	svc.ForgetSubject(context.Background(), "")

	if deleteCalled {
		t.Error("DeleteSecretBySubjectId must not be called when subject ID is empty")
	}
}

func TestSubjectService_ForgetSubject_PassesSubjectIDToStore(t *testing.T) {
	const subjectID = "subject-abc"
	var capturedID string

	repos := defaultRepos()
	repos.SubjectSecretStore = &mockSubjectSecretStore{
		DeleteSecretBySubjectIDFunc: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}

	svc := NewSubjectService(&mockTx{repos: repos}, newTestLogger())
	svc.ForgetSubject(context.Background(), subjectID)

	if capturedID != subjectID {
		t.Errorf("store received subjectID %q, want %q", capturedID, subjectID)
	}
}
