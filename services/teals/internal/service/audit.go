package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
)

// AuditIngestor defines the interface for ingesting audit events.
type AuditIngestor interface {
	IngestAuditEvent(ctx context.Context, event *model.AuditEvent, signature string) (*model.IngestAuditEventResult, error)
}

// AuditService provides methods to ingest audit events, handling the necessary serialization, database transactions, and error management. It interacts with the TransactionProvider to manage database operations and the Serializer to convert audit events into a storable format.
type AuditService struct {
	tx         ports.TransactionProvider
	serializer ports.Serializer
	verifier   ports.SignatureVerifier
	protector  ports.MetadataProtector
	logger     *logger.Logger
}

// NewAuditService creates a new instance of AuditService with the provided TransactionProvider, Serializer, and Logger. This allows the service to manage database transactions, serialize audit events, and log important information and errors during the ingestion process.
func NewAuditService(tx ports.TransactionProvider, s ports.Serializer, v ports.SignatureVerifier, p ports.MetadataProtector, l *logger.Logger) *AuditService {
	return &AuditService{
		tx:         tx,
		serializer: s,
		verifier:   v,
		protector:  p,
		logger:     l,
	}
}

// IngestAuditEvent handles the ingestion of an audit event by verifying its signature, protecting its metadata, and storing it in the ledger and audit log. It manages the entire process within a database transaction to ensure atomicity, and returns the result of the ingestion or any errors that occur during the process, such as invalid signatures, serialization failures, or database errors.
func (s *AuditService) IngestAuditEvent(ctx context.Context, event *model.AuditEvent, sigToken string) (*model.IngestAuditEventResult, error) {
	kid, err := s.verifyEventSignature(ctx, event, sigToken)
	if err != nil {
		return nil, err
	}

	var nodeID int64
	var size int64
	err = s.tx.Transact(ctx, func(r ports.Repositories) error {
		producerKey, err := r.ProducerKeys.GetProducerKeyByKid(ctx, kid)
		if err != nil {
			s.logger.Error("failed to retrieve producer key by kid", "kid", kid, "error", err)
			return svcerrors.ErrProducerKeyRetrievalFailed
		}

		protectedEvent, salt, err := s.protectAuditEvent(ctx, r, event)
		if err != nil {
			return err
		}
		protectedPayloadBytes, err := s.serializer.SerializeCanonicalProtectedAuditEvent(protectedEvent)
		if err != nil {
			s.logger.Error("failed to serialize protected audit event", "event_id", event.ID, "error", err)
			return svcerrors.ErrEventSerializationFailed
		}

		nodeID, size, err = r.Ledger.AppendLeaf(ctx, protectedPayloadBytes)
		if err != nil {
			s.logger.Error("failed to append audit event to ledger", "error", err)
			return svcerrors.ErrLedgerAppendFailed
		}

		return r.AuditLog.StoreAuditLogEntry(ctx, event.ID, protectedPayloadBytes, sigToken, producerKey.ID, nodeID, salt)
	})

	if err != nil {
		if errors.Is(err, svcerrors.ErrDuplicateEventID) {
			s.logger.Warn("duplicate audit event rejected", "event_id", event.ID)
			return nil, svcerrors.ErrDuplicateEventID
		}
		return nil, err
	}
	s.logger.Info("successfully appended audit event", "event_id", event.ID)

	return &model.IngestAuditEventResult{
		EventID:    event.ID,
		LedgerSize: size,
		IngestedAt: event.Timestamp,
	}, nil
}

// verifyEventSignature handles the verification of an audit event's signature by serializing the event to its canonical form and using the SignatureVerifier to validate the JWS signature against the serialized payload. It returns the key ID (KID) extracted from the signature if verification is successful, or an appropriate error if serialization fails or if the signature is invalid.
func (s *AuditService) verifyEventSignature(ctx context.Context, event *model.AuditEvent, sigToken string) (string, error) {
	// 1. Serialize to canonical form — this is the exact bytes that were signed and the exact bytes that will be stored.
	payloadBytes, err := s.serializer.SerializeCanonicalAuditEvent(event)
	if err != nil {
		s.logger.Error("failed to serialize audit event payload", "error", err)
		return "", svcerrors.ErrEventSerializationFailed
	}

	// 2. Verify the JWS signature against the canonical payload. KID is extracted from the token's protected header by the verifier.
	kid, err := s.verifier.Verify(ctx, sigToken, payloadBytes)
	if err != nil {
		s.logger.Warn("audit event rejected: invalid signature", "event_id", event.ID, "error", err)
		return "", svcerrors.ErrInvalidSignature
	}
	s.logger.Info("signature verified successfully", "event_id", event.ID, "kid", kid)
	return kid, nil
}

// protectAuditEvent handles the protection of an audit event's metadata by retrieving or creating a subject secret, using the MetadataProtector to encrypt the metadata, and constructing a ProtectedAuditEvent with the protected metadata. It returns the ProtectedAuditEvent, the salt used for protection, and any error that occurs during the process, such as failures in retrieving the subject secret or protecting the metadata.
func (s *AuditService) protectAuditEvent(ctx context.Context, r ports.Repositories, event *model.AuditEvent) (*model.ProtectedAuditEvent, []byte, error) {
	var protectedMeta *model.ProtectedMetadata
	var salt []byte

	if event.Metadata != nil {
		secret, err := r.SubjectSecretStore.GetOrCreateSecret(ctx, event.Subject.ID)
		if err != nil {
			s.logger.Error("failed to get or create subject secret", "subject_id", event.Subject.ID, "error", err)
			return nil, nil, err
		}
		protectedMeta, salt, err = s.protector.Protect(secret, event.Metadata)
		if err != nil {
			s.logger.Error("failed to protect audit event metadata", "event_id", event.ID, "error", err)
			return nil, nil, svcerrors.ErrProtectionFailed
		}
	}

	protectedEvent, err := model.NewProtectedAuditEvent(model.CreateProtectedAuditEventParams{
		BaseEventParams: model.BaseEventParams{
			ID:          event.ID,
			Timestamp:   event.Timestamp,
			Environment: event.Environment,
			Actor:       event.Actor,
			Subject:     event.Subject,
			Action:      event.Action,
			Resource:    event.Resource,
			Result:      event.Result,
		},
		ProtectedMetadata: protectedMeta,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build protected audit event: %w", err)
	}

	return protectedEvent, salt, nil
}
