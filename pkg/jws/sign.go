package jws

import (
	"crypto/ed25519"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
)

// Signer is an interface for signing payloads and producing JWS tokens.
type Signer interface {
	Sign(payload []byte) (string, error)
}

// Ed25519Signer implements the Signer interface using Ed25519 keys.
type Ed25519Signer struct {
	key ed25519.PrivateKey
	kid string
}

// NewEd25519Signer creates a new Ed25519Signer with the given private key.
func NewEd25519Signer(key ed25519.PrivateKey, kid string) (*Ed25519Signer, error) {
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("jws: invalid Ed25519 private key size")
	}
	return &Ed25519Signer{key: key, kid: kid}, nil
}

// Sign creates a JWS token by signing the given payload with the Ed25519 private key. It returns the compact serialization of the JWS token.
func (s *Ed25519Signer) Sign(payload []byte) (string, error) {
	headers := jws.NewHeaders()
	err := headers.Set(jws.KeyIDKey, s.kid)
	if err != nil {
		return "", fmt.Errorf("jws: set header error: %w", err)
	}

	token, err := jws.Sign(
		nil,
		jws.WithKey(jwa.EdDSA, s.key, jws.WithProtectedHeaders(headers)),
		jws.WithCompact(),
		jws.WithDetachedPayload(payload))
	if err != nil {
		return "", fmt.Errorf("jws: sign: %w", err)
	}
	return string(token), nil
}
