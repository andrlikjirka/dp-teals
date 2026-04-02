package service

import (
	"context"
	"errors"

	"github.com/andrlijirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlijirka/dp-teals/services/teals-server/internal/service/errors"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/model"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/ports"
	"github.com/google/uuid"
)

type AuditService struct {
	tx         ports.TransactionProvider
	serializer ports.Serializer
	logger     *logger.Logger
}

func NewAuditService(tx ports.TransactionProvider, s ports.Serializer, l *logger.Logger) *AuditService {
	return &AuditService{
		tx:         tx,
		serializer: s,
		logger:     l,
	}
}

func (s *AuditService) IngestAuditEvent(ctx context.Context, event *model.AuditEvent) (uuid.UUID, error) {
	payloadBytes, err := s.serializer.SerializeCanonicalAuditEvent(event)
	if err != nil {
		s.logger.Error("failed to serialize audit event payload", "error", err)
		return uuid.Nil, svcerrors.ErrEventSerializationFailed
	}

	err = s.tx.Transact(ctx, func(r ports.Repositories) error {
		return r.AuditLog.StoreAuditLogEntry(ctx, event.ID, payloadBytes)
	})

	if err != nil {
		if errors.Is(err, svcerrors.ErrDuplicateEventID) {
			s.logger.Warn("duplicate audit event rejected", "event_id", event.ID)
			return uuid.Nil, svcerrors.ErrDuplicateEventID
		}
		s.logger.Error("failed to append audit event", "error", err)
		return uuid.Nil, svcerrors.ErrEventAppendFailed
	}
	s.logger.Info("successfully appended audit event", "event_id", event.ID)

	return event.ID, nil
}
