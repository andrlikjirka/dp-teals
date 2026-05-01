package ports

import "context"

// SubjectSecretStore defines the interface for creating or retrieving data subject secrets associated with a subject ID. It abstracts the underlying storage mechanism, allowing for different implementations to be used without affecting the service logic.
type SubjectSecretStore interface {
	// GetOrCreateSecret takes a context and a subject ID as input and returns the corresponding secret as a byte slice. If a secret does not already exist for the given subject ID, it creates a new one and returns it. An error is returned if the operation fails.
	GetOrCreateSecret(ctx context.Context, subjectID string) ([]byte, error)
	// GetSecretBySubjectId takes a context and a subject ID as input and returns the corresponding secret as a byte slice or an error if the retrieval fails.
	GetSecretBySubjectId(ctx context.Context, subjectID string) ([]byte, error)
	// DeleteSecretBySubjectId takes a context and a subject ID as input and deletes the corresponding secret. It returns an error if the deletion fails. This method is essential for implementing the "right to be forgotten" functionality, allowing for the removal of secrets associated with a subject when they request to be forgotten.
	DeleteSecretBySubjectId(ctx context.Context, subjectID string) error // new
}
