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

// AuditService provides methods to ingest audit events, handling the necessary serialization, database transactions, and error management. It interacts with the TransactionProvider to manage database operations and the Serializer to convert audit events into a storable format.
type AuditService struct {
	tx         ports.TransactionProvider
	serializer ports.Serializer
	verifier   ports.SignatureVerifier
	logger     *logger.Logger
}

// NewAuditService creates a new instance of AuditService with the provided TransactionProvider, Serializer, and Logger. This allows the service to manage database transactions, serialize audit events, and log important information and errors during the ingestion process.
func NewAuditService(tx ports.TransactionProvider, s ports.Serializer, v ports.SignatureVerifier, l *logger.Logger) *AuditService {
	return &AuditService{
		tx:         tx,
		serializer: s,
		verifier:   v,
		logger:     l,
	}
}

// IngestAuditEvent handles the ingestion of an audit event by serializing the event, retrieving the associated producer key, and storing the event in the database within a transaction. It returns the event ID if successful, or an appropriate error if any step of the process fails, including serialization errors, producer key retrieval failures, or database insertion issues such as duplicate event IDs.
func (s *AuditService) IngestAuditEvent(ctx context.Context, event *model.AuditEvent, sigToken string) (uuid.UUID, error) {
	// 1. Serialize to canonical form — this is the exact bytes that were signed and the exact bytes that will be stored.
	payloadBytes, err := s.serializer.SerializeCanonicalAuditEvent(event)
	if err != nil {
		s.logger.Error("failed to serialize audit event payload", "error", err)
		return uuid.Nil, svcerrors.ErrEventSerializationFailed
	}

	// 2. Verify the JWS signature against the canonical payload. KID is extracted from the token's protected header by the verifier.
	kid, err := s.verifier.Verify(ctx, sigToken, payloadBytes)
	if err != nil {
		s.logger.Warn("audit event rejected: invalid signature", "event_id", event.ID, "error", err)
		return uuid.Nil, svcerrors.ErrInvalidSignature
	}
	s.logger.Info("signature verified successfully", "event_id", event.ID, "kid", kid)

	err = s.tx.Transact(ctx, func(r ports.Repositories) error {
		producerKey, err := r.ProducerKeys.GetProducerKeyByKid(ctx, kid)
		if err != nil {
			s.logger.Error("failed to retrieve producer key by kid", "kid", kid, "error", err)
			return svcerrors.ErrProducerKeyRetrievalFailed
		}

		return r.AuditLog.StoreAuditLogEntry(ctx, event.ID, payloadBytes, sigToken, producerKey.ID)
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
