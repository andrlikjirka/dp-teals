package repository

import (
	"context"

	"github.com/google/uuid"
)

type SubjectKeyRepository interface {
	SaveSubjectKey(ctx context.Context, subjectID uuid.UUID, encryptedDEK any) error
	GetKeyBySubjectId(ctx context.Context, subjectID uuid.UUID) ([]byte, error)
	DeleteKeyBySubjectId(ctx context.Context, subjectID uuid.UUID) error
}

// TODO: change the interface to struct implementation
