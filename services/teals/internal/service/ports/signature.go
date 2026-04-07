package ports

import "context"

// SignatureVerifier defines the interface for verifying JWS signatures of incoming audit events. It abstracts the logic for validating the authenticity and integrity of the event data using the provided JWS token and payload. The Verify method takes a context, the JWS token, and the event payload as input, and returns the key ID (kid) used for signing if the verification is successful, or an error if the verification fails.
type SignatureVerifier interface {
	// Verify validates the JWS token against the provided payload and returns the key ID (kid) if the signature is valid. It returns an error if the signature is invalid or if any issues occur during the verification process.
	Verify(ctx context.Context, token string, payload []byte) (string, error)
}
