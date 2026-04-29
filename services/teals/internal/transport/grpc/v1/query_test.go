package v1

import (
	"context"
	"testing"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/transport/grpc/v1/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

// --- mock ---

type mockQueryService struct {
	GetAuditEventFunc   func(ctx context.Context, eventID uuid.UUID) (*svcmodel.GetAuditEventResult, error)
	ListAuditEventsFunc func(ctx context.Context, filter *svcmodel.AuditEventFilter, cursor *int64) (*svcmodel.ListAuditEventsResult, error)
}

func (m *mockQueryService) GetAuditEvent(ctx context.Context, eventID uuid.UUID) (*svcmodel.GetAuditEventResult, error) {
	if m.GetAuditEventFunc != nil {
		return m.GetAuditEventFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *mockQueryService) ListAuditEvents(ctx context.Context, filter *svcmodel.AuditEventFilter, cursor *int64) (*svcmodel.ListAuditEventsResult, error) {
	if m.ListAuditEventsFunc != nil {
		return m.ListAuditEventsFunc(ctx, filter, cursor)
	}
	return nil, nil
}

// --- GetAuditEvent ---

func TestGetAuditEvent_InvalidEventID_ReturnsInvalidArgument(t *testing.T) {
	s := NewQueryServiceServer(&mockQueryService{})

	_, err := s.GetAuditEvent(context.Background(), &auditv1.GetAuditEventRequest{
		EventId: "not-a-uuid",
	})

	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestGetAuditEvent_ServiceErrors(t *testing.T) {
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
			name:     "unexpected internal error",
			svcErr:   svcerrors.ErrEventDeserializationFailed,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockQueryService{
				GetAuditEventFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.GetAuditEventResult, error) {
					return nil, tc.svcErr
				},
			}
			s := NewQueryServiceServer(svc)

			_, err := s.GetAuditEvent(context.Background(), &auditv1.GetAuditEventRequest{
				EventId: validEventID,
			})
			assertGRPCCode(t, err, tc.wantCode)
		})
	}
}

func TestGetAuditEvent_Success_ResponseFieldsPopulated(t *testing.T) {
	result := &svcmodel.GetAuditEventResult{
		Payload:          validPayload,
		RevealedMetadata: map[string]any{"name": "Alice"},
		LeafIndex:        7,
		SignatureToken:   "sig-token",
	}

	svc := &mockQueryService{
		GetAuditEventFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.GetAuditEventResult, error) {
			return result, nil
		},
	}
	s := NewQueryServiceServer(svc)

	resp, err := s.GetAuditEvent(context.Background(), &auditv1.GetAuditEventRequest{
		EventId: validEventID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Event == nil {
		t.Error("expected Event in response")
	}
	if resp.RevealedMetadata == nil {
		t.Error("expected RevealedMetadata in response")
	}
	if resp.LeafIndex != result.LeafIndex {
		t.Errorf("LeafIndex: got %d, want %d", resp.LeafIndex, result.LeafIndex)
	}
	if resp.ProducerSignToken != result.SignatureToken {
		t.Errorf("ProducerSignToken: got %q, want %q", resp.ProducerSignToken, result.SignatureToken)
	}
}

func TestGetAuditEvent_Success_NilRevealedMetadata(t *testing.T) {
	result := &svcmodel.GetAuditEventResult{
		Payload:          validPayload,
		RevealedMetadata: nil,
	}

	svc := &mockQueryService{
		GetAuditEventFunc: func(_ context.Context, _ uuid.UUID) (*svcmodel.GetAuditEventResult, error) {
			return result, nil
		},
	}
	s := NewQueryServiceServer(svc)

	resp, err := s.GetAuditEvent(context.Background(), &auditv1.GetAuditEventRequest{
		EventId: validEventID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RevealedMetadata != nil {
		t.Errorf("expected nil RevealedMetadata, got %v", resp.RevealedMetadata)
	}
}

func TestGetAuditEvent_EventIDForwardedToService(t *testing.T) {
	expectedID := uuid.MustParse(validEventID)
	var capturedID uuid.UUID

	svc := &mockQueryService{
		GetAuditEventFunc: func(_ context.Context, id uuid.UUID) (*svcmodel.GetAuditEventResult, error) {
			capturedID = id
			return &svcmodel.GetAuditEventResult{Payload: validPayload}, nil
		},
	}
	s := NewQueryServiceServer(svc)

	_, err := s.GetAuditEvent(context.Background(), &auditv1.GetAuditEventRequest{
		EventId: validEventID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != expectedID {
		t.Errorf("eventID forwarded: got %v, want %v", capturedID, expectedID)
	}
}

// --- ListAuditEvents ---

func TestListAuditEvents_InvalidCursor_ReturnsInvalidArgument(t *testing.T) {
	s := NewQueryServiceServer(&mockQueryService{})

	cursor := "not-valid-cursor"
	_, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{
		Cursor: &cursor,
	})

	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestListAuditEvents_ServiceError_ReturnsInternal(t *testing.T) {
	svc := &mockQueryService{
		ListAuditEventsFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64) (*svcmodel.ListAuditEventsResult, error) {
			return nil, svcerrors.ErrAuditLogEntryNotFound
		},
	}
	s := NewQueryServiceServer(svc)

	_, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{})

	assertGRPCCode(t, err, codes.Internal)
}

func TestListAuditEvents_Success_ResponseFieldsPopulated(t *testing.T) {
	nextCursorVal := int64(99)
	result := &svcmodel.ListAuditEventsResult{
		Items: []*svcmodel.AuditEventListItem{
			{Payload: validPayload, SignatureToken: "tok-1", LeafIndex: 1},
			{Payload: validPayload, SignatureToken: "tok-2", LeafIndex: 2},
		},
		LedgerSize: 42,
		NextCursor: &nextCursorVal,
	}

	svc := &mockQueryService{
		ListAuditEventsFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64) (*svcmodel.ListAuditEventsResult, error) {
			return result, nil
		},
	}
	s := NewQueryServiceServer(svc)

	resp, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Fatalf("Items length: got %d, want 2", len(resp.Items))
	}
	if resp.LedgerSize != result.LedgerSize {
		t.Errorf("LedgerSize: got %d, want %d", resp.LedgerSize, result.LedgerSize)
	}
	if resp.NextCursor == nil {
		t.Fatal("expected NextCursor to be set")
	}
}

func TestListAuditEvents_Success_NoNextCursorWhenNil(t *testing.T) {
	result := &svcmodel.ListAuditEventsResult{
		Items:      []*svcmodel.AuditEventListItem{},
		LedgerSize: 5,
		NextCursor: nil,
	}

	svc := &mockQueryService{
		ListAuditEventsFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64) (*svcmodel.ListAuditEventsResult, error) {
			return result, nil
		},
	}
	s := NewQueryServiceServer(svc)

	resp, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NextCursor != nil {
		t.Errorf("expected nil NextCursor, got %v", *resp.NextCursor)
	}
}

func TestListAuditEvents_Success_ItemFieldsPopulated(t *testing.T) {
	result := &svcmodel.ListAuditEventsResult{
		Items: []*svcmodel.AuditEventListItem{
			{
				Payload:          validPayload,
				RevealedMetadata: map[string]any{"key": "val"},
				SignatureToken:   "sig-abc",
				LeafIndex:        5,
			},
		},
		LedgerSize: 10,
	}

	svc := &mockQueryService{
		ListAuditEventsFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, _ *int64) (*svcmodel.ListAuditEventsResult, error) {
			return result, nil
		},
	}
	s := NewQueryServiceServer(svc)

	resp, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	item := resp.Items[0]
	if item.Event == nil {
		t.Error("expected Event in item")
	}
	if item.RevealedMetadata == nil {
		t.Error("expected RevealedMetadata in item")
	}
	if item.LeafIndex != 5 {
		t.Errorf("LeafIndex: got %d, want 5", item.LeafIndex)
	}
	if item.ProducerSignToken != "sig-abc" {
		t.Errorf("ProducerSignToken: got %q, want %q", item.ProducerSignToken, "sig-abc")
	}
}

func TestListAuditEvents_CursorDecodedAndForwardedToService(t *testing.T) {
	var capturedCursor *int64

	svc := &mockQueryService{
		ListAuditEventsFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, cursor *int64) (*svcmodel.ListAuditEventsResult, error) {
			capturedCursor = cursor
			return &svcmodel.ListAuditEventsResult{}, nil
		},
	}
	s := NewQueryServiceServer(svc)

	// Encode cursor value 42 the same way the handler encodes it.
	encoded := model.EncodeCursor(42)
	_, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{
		Cursor: &encoded,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedCursor == nil || *capturedCursor != 42 {
		t.Errorf("cursor forwarded: got %v, want 42", capturedCursor)
	}
}

func TestListAuditEvents_NilCursorForwardedWhenAbsent(t *testing.T) {
	var capturedCursor *int64

	svc := &mockQueryService{
		ListAuditEventsFunc: func(_ context.Context, _ *svcmodel.AuditEventFilter, cursor *int64) (*svcmodel.ListAuditEventsResult, error) {
			capturedCursor = cursor
			return &svcmodel.ListAuditEventsResult{}, nil
		},
	}
	s := NewQueryServiceServer(svc)

	_, err := s.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedCursor != nil {
		t.Errorf("expected nil cursor forwarded, got %d", *capturedCursor)
	}
}
