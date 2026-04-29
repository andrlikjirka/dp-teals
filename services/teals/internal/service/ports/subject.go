package ports

import "context"

// SubjectSecretStore defines the interface for creating or retrieving data subject secrets associated with a subject ID. It abstracts the underlying storage mechanism, allowing for different implementations to be used without affecting the service logic.
type SubjectSecretStore interface {
	// GetOrCreateSecret takes a context and a subject ID as input and returns the corresponding secret as a byte slice. If a secret does not already exist for the given subject ID, it creates a new one and returns it. An error is returned if the operation fails.
	GetOrCreateSecret(ctx context.Context, subjectID string) ([]byte, error)
	// GetSecretBySubjectId takes a context and a subject ID as input and returns the corresponding secret as a byte slice or an error if the retrieval fails.
	GetSecretBySubjectId(ctx context.Context, subjectID string) ([]byte, error)
}
