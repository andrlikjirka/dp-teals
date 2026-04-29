package v1

import (
	"encoding/json"
	"testing"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const validEventID = "550e8400-e29b-41d4-a716-446655440000"
const validProducerID = "550e8400-e29b-41d4-a716-446655440000"

var validPayload = json.RawMessage(`{"action":"ACCESS"}`)

// assertGRPCCode is a helper function that checks if the provided error is a gRPC status error with the expected code.
func assertGRPCCode(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %v, got nil", want)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != want {
		t.Errorf("gRPC code: got %v, want %v", st.Code(), want)
	}
}

// validAppendRequest returns a minimal AppendRequest that passes all mapping validations.
func validAppendRequest() *auditv1.AppendRequest {
	return &auditv1.AppendRequest{
		Event: &auditv1.AuditEvent{
			Id:        validEventID,
			Timestamp: timestamppb.New(time.Now()),
			Actor:     &auditv1.Actor{Type: auditv1.Actor_TYPE_USER, Id: "actor-1"},
			Subject:   &auditv1.Subject{Id: "subject-1"},
			Action:    auditv1.Action_ACTION_ACCESS,
			Resource:  &auditv1.Resource{Id: "res-1", Name: "resource-name"},
			Result:    &auditv1.Result{Status: auditv1.Result_STATUS_SUCCESS},
		},
	}
}

// successResult returns a realistic IngestAuditEventResult.
func successResult() *svcmodel.IngestAuditEventResult {
	return &svcmodel.IngestAuditEventResult{
		EventID:    uuid.MustParse(validEventID),
		LedgerSize: 10,
		IngestedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}
