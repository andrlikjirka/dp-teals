package service

import (
	"context"
	"errors"
	"time"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
)

// SubjectForgetter defines the interface for forgetting a data subject by deleting its associated cryptographic secret from the database. It provides a method ForgetSubject that takes a context and a subject ID, and returns an error if the operation fails. This interface allows for abstraction of the subject forgetting logic, enabling different implementations if needed, while ensuring that the service layer can interact with it in a consistent manner.
type SubjectForgetter interface {
	ForgetSubject(ctx context.Context, subjectID string) (model.ForgetSubjectResult, error)
}

// SubjectService provides methods to manage data subjects, including the ability to "forget" a subject by deleting its associated cryptographic secret from the database. It interacts with the TransactionProvider to perform database operations within transactions and uses the Logger to log important information and errors during the execution of service methods.
type SubjectService struct {
	tx     ports.TransactionProvider
	logger *logger.Logger
}

// NewSubjectService creates a new instance of SubjectService with the provided TransactionProvider and Logger. This service is responsible for managing operations related to data subjects, such as forgetting a subject by deleting its associated cryptographic secret from the database. The TransactionProvider allows the service to perform database operations within transactions, while the Logger is used to log important information and errors during the execution of service methods.
func NewSubjectService(tx ports.TransactionProvider, logger *logger.Logger) *SubjectService {
	return &SubjectService{
		tx:     tx,
		logger: logger,
	}
}

// ForgetSubject deletes the cryptographic secret associated with the given subject ID from the database, effectively "forgetting" the subject. It performs the deletion within a transaction and logs relevant information and errors during the process. If the subject ID is missing or if no secret is found for the specified subject ID, it returns appropriate errors. On successful deletion, it returns a ForgetSubjectResult containing the subject ID and the timestamp of when the subject was forgotten.
func (s *SubjectService) ForgetSubject(ctx context.Context, subjectID string) (model.ForgetSubjectResult, error) {
	if subjectID == "" {
		return model.ForgetSubjectResult{}, svcerrors.ErrMissingSubjectID
	}
	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		return r.SubjectSecretStore.DeleteSecretBySubjectId(ctx, subjectID)
	})
	if err != nil {
		if errors.Is(err, svcerrors.ErrSubjectSecretNotFound) {
			s.logger.Warn("forget subject requested but no secret found", "subject_id", subjectID)
			return model.ForgetSubjectResult{}, svcerrors.ErrSubjectSecretNotFound
		}
		s.logger.Error("failed to delete subject secret", "subject_id", subjectID, "error", err)
		return model.ForgetSubjectResult{}, svcerrors.ErrSubjectSecretDeletionFailed
	}
	s.logger.Info("subject secret deleted", "subject_id", subjectID)
	return model.ForgetSubjectResult{
		SubjectID:   subjectID,
		ForgottenAt: time.Now().UTC(),
	}, nil
}
