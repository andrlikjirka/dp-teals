package jws

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// Thumbprint computes the JWK thumbprint for the given Ed25519 public key (RFC 7638). It returns the thumbprint as a base64url-encoded string. The thumbprint is computed using the SHA-256 hash of the JWK representation of the public key.
func Thumbprint(pub ed25519.PublicKey) (string, error) {
	jwkKey, err := jwk.FromRaw(pub)
	if err != nil {
		return "", fmt.Errorf("jws: create JWK: %w", err)
	}
	raw, err := jwkKey.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", fmt.Errorf("jws: compute thumbprint: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
