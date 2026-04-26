package service

import (
	"context"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// auditEventPageSize defines the maximum number of audit events to return in a single page when listing audit events. If more than this number of events are available, a next cursor will be provided for pagination.
const auditEventPageSize = 100

// QueryService provides methods to query audit events and their inclusion proofs.
type QueryService struct {
	tx         ports.TransactionProvider
	serializer ports.Serializer
	logger     *logger.Logger
}

// NewQueryService creates a new instance of QueryService with the provided transaction provider, serializer, and logger.
func NewQueryService(tx ports.TransactionProvider, s ports.Serializer, l *logger.Logger) *QueryService {
	return &QueryService{
		tx:         tx,
		serializer: s,
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
			Event:          event,
			LeafIndex:      entry.LeafIndex,
			SignatureToken: entry.SignatureToken,
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
		size, err := r.LedgerProver.Size(ctx)
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
				Event:          event,
				SignatureToken: entry.SignatureToken,
				LeafIndex:      entry.LeafIndex,
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
