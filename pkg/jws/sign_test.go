package jws

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestEd25519Signer_Sign(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	kid := "test-key-v1"

	signer, err := NewEd25519Signer(priv, kid)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("test payload")
	token, err := signer.Sign(payload)
	if err != nil {
		t.Fatalf("sign returned unexpected error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3-part compact JWS, got %d parts", len(parts))
	}

	if parts[1] != "" {
		t.Errorf("expected detached payload (empty middle part), got %q", parts[1])
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("failed to base64 decode header: %v", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		t.Fatalf("failed to unmarshal header JSON: %v", err)
	}

	if header["kid"] != kid {
		t.Errorf("expected kid %q in header, got %v", kid, header["kid"])
	}

	expectedAlg := "EdDSA"
	if header["alg"] != expectedAlg {
		t.Errorf("expected alg '%v' in header, got %v", expectedAlg, header["alg"])
	}
}

func TestEd25519Signer_Sign_InvalidKey(t *testing.T) {
	kid := "test-key-v1"
	_, err := NewEd25519Signer([]byte("invalid key"), kid)
	if err == nil {
		t.Fatal("expected error for invalid key size")
	}
}
