package jws

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jws"
)

type KeyProvider interface {
	PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error)
}

// Verifier is an interface for verifying JWS tokens and their associated payloads.
type Verifier interface {
	Verify(ctx context.Context, token string, payload []byte) error
}

// Ed25519Verifier implements the Verifier interface using Ed25519 public keys.
type Ed25519Verifier struct {
	provider KeyProvider
}

// NewEd25519Verifier creates a new Ed25519Verifier with the given public key.
func NewEd25519Verifier(p KeyProvider) (*Ed25519Verifier, error) {
	return &Ed25519Verifier{provider: p}, nil
}

// Verify checks the JWS token's signature against the provided payload using the Ed25519 public key. It returns an error if the signature is invalid or if the payload does not match the one in the token.
func (v *Ed25519Verifier) Verify(ctx context.Context, token string, payload []byte) error {
	msg, err := jws.Parse([]byte(token))
	if err != nil {
		return fmt.Errorf("jws: parse token error: %w", err)
	}

	kid, ok := msg.Signatures()[0].ProtectedHeaders().KeyID()
	if !ok {
		return fmt.Errorf("jws: missing kid header")
	}

	pub, err := v.provider.PublicKey(ctx, kid)
	if err != nil {
		return fmt.Errorf("jws: get public key error: %w", err)
	}

	_, err = jws.Verify([]byte(token), jws.WithKey(jwa.EdDSA(), pub), jws.WithDetachedPayload(payload))
	if err != nil {
		return fmt.Errorf("jws: verify signature error: %w", err)
	}

	return nil
}
