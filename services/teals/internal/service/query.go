package service

import (
	"context"
	"errors"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// auditEventPageSize defines the maximum number of audit events to return in a single page when listing audit events. If more than this number of events are available, a next cursor will be provided for pagination.
const auditEventPageSize = 10

// AuditQuerier defines the interface for querying audit events.
type AuditQuerier interface {
	GetAuditEvent(ctx context.Context, eventID uuid.UUID) (*model.GetAuditEventResult, error)
	ListAuditEvents(ctx context.Context, filter *model.AuditEventFilter, cursor *int64) (*model.ListAuditEventsResult, error)
}

// QueryService provides methods to query audit events and their inclusion proofs.
type QueryService struct {
	tx         ports.TransactionProvider
	serializer ports.Serializer
	protector  ports.MetadataProtector
	logger     *logger.Logger
}

// NewQueryService creates a new instance of QueryService with the provided transaction provider, serializer, metadata protector and logger.
func NewQueryService(tx ports.TransactionProvider, s ports.Serializer, p ports.MetadataProtector, l *logger.Logger) *QueryService {
	return &QueryService{
		tx:         tx,
		serializer: s,
		protector:  p,
		logger:     l,
	}
}

// GetAuditEvent retrieves an audit event by its event ID, including the event data, leaf index, and signature token. It returns an error if the event is not found or if there was an issue during retrieval or deserialization.
func (s *QueryService) GetAuditEvent(ctx context.Context, eventID uuid.UUID) (*model.GetAuditEventResult, error) {
	var result *model.GetAuditEventResult

	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		entry, err := r.AuditLog.GetAuditLogEntryByEventID(ctx, eventID)
		if err != nil {
			s.logger.Error("failed to retrieve audit log entry by event ID", "event_id", eventID, "error", err)
			return svcerrors.ErrAuditLogEntryNotFound
		}

		event, err := s.serializer.DeserializeCanonicalProtectedAuditEvent(entry.Payload)
		if err != nil {
			s.logger.Error("failed to deserialize audit event payload", "event_id", eventID, "error", err)
			return svcerrors.ErrEventDeserializationFailed
		}

		result = &model.GetAuditEventResult{
			Event:            event,
			Payload:          entry.Payload,
			LeafIndex:        entry.LeafIndex,
			SignatureToken:   entry.SignatureToken,
			RevealedMetadata: s.tryRevealMetadata(ctx, r, event),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("successfully retrieved audit event", "event_id", eventID)
	return result, nil
}

// ListAuditEvents retrieves a list of audit events based on the provided filter, along with the current ledger size. It returns an error if there was an issue during retrieval or deserialization of any of the events.
func (s *QueryService) ListAuditEvents(ctx context.Context, filter *model.AuditEventFilter, cursor *int64) (*model.ListAuditEventsResult, error) {
	var result *model.ListAuditEventsResult

	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		size, err := r.Ledger.Size(ctx)
		if err != nil {
			s.logger.Error("failed to get ledger size", "error", err)
			return svcerrors.ErrLedgerSizeFailed
		}

		entries, err := r.AuditLog.ListAuditLogEntries(ctx, filter, cursor, auditEventPageSize+1)
		if err != nil {
			s.logger.Error("failed to list audit log entries", "error", err)
			return svcerrors.ErrAuditLogEntryNotFound
		}

		var nextCursor *int64
		if len(entries) == auditEventPageSize+1 {
			entries = entries[:auditEventPageSize]
			nextCursor = entries[auditEventPageSize-1].ID
		}

		items := make([]*model.AuditEventListItem, len(entries))
		for i, entry := range entries {
			event, err := s.serializer.DeserializeCanonicalProtectedAuditEvent(entry.Payload)
			if err != nil {
				s.logger.Error("failed to deserialize audit event payload", "event_id", entry.EventID, "error", err)
				return svcerrors.ErrEventDeserializationFailed
			}
			items[i] = &model.AuditEventListItem{
				Event:            event,
				Payload:          entry.Payload,
				SignatureToken:   entry.SignatureToken,
				LeafIndex:        entry.LeafIndex,
				RevealedMetadata: s.tryRevealMetadata(ctx, r, event),
			}
		}

		result = &model.ListAuditEventsResult{Items: items, LedgerSize: size, NextCursor: nextCursor}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("audit events listed", "count", len(result.Items))
	return result, nil
}

// tryRevealMetadata attempts to reveal the protected metadata of an audit event using the subject's secret. If the event does not have protected metadata or if there was an issue during retrieval or revelation, it returns nil and logs the error.
func (s *QueryService) tryRevealMetadata(ctx context.Context, r ports.Repositories, event *model.ProtectedAuditEvent) map[string]any {
	if event.ProtectedMetadata == nil {
		return nil
	}

	secret, err := r.SubjectSecretStore.GetSecretBySubjectId(ctx, event.Subject.ID)
	if err != nil {
		if !errors.Is(err, svcerrors.ErrSubjectSecretNotFound) {
			s.logger.Error("failed to retrieve subject secret for metadata reveal", "subject_id", event.Subject.ID, "error", err)
		}
		return nil
	}

	revealed, err := s.protector.Reveal(secret, event.ProtectedMetadata)
	if err != nil {
		s.logger.Error("failed to reveal protected metadata", "subject_id", event.Subject.ID, "error", err)
		return nil
	}

	return revealed
}
