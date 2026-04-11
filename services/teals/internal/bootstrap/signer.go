package bootstrap

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	pkgjws "github.com/andrlikjirka/dp-teals/pkg/jws"
)

// NewServerSigner decodes the Ed25519 private key from config, derives the KID, and returns a signer ready for checkpoint signing.
func NewServerSigner(cfg Config) (*pkgjws.Ed25519Signer, error) {
	privBytes, err := base64.StdEncoding.DecodeString(cfg.ServerPrivateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode server private key: %w", err)
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid server private key: got %d bytes, want %d", len(privBytes), ed25519.PrivateKeySize)
	}

	priv := ed25519.PrivateKey(privBytes)
	kid, err := pkgjws.Thumbprint(priv.Public().(ed25519.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("compute server key thumbprint: %w", err)
	}

	return pkgjws.NewEd25519Signer(priv, kid)
}
