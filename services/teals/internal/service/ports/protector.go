package ports

import "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"

type MetadataProtector interface {
	// Protect encrypts PII metadata and produces a protected metadata + salt.
	Protect(secret []byte, metadata map[string]any) (*model.ProtectedMetadata, []byte, error)
	// Reveal decrypts protected metadata back to plaintext PII metadata object.
	Reveal(subjectSecret []byte, pm *model.ProtectedMetadata) (map[string]any, error)
}
