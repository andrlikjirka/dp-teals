package jws

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"testing"
)

type staticKeyProvider struct {
	keys map[string]ed25519.PublicKey
}

func (p *staticKeyProvider) PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	pub, ok := p.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key not found for kid: %s", kid)
	}
	return pub, nil
}

func TestEd25519Verifier_Verify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	kid := "test-key-v1"

	signer, err := NewEd25519Signer(priv, kid)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewEd25519Verifier(&staticKeyProvider{keys: map[string]ed25519.PublicKey{kid: pub}})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("test payload")
	token, _ := signer.Sign(payload)

	err = verifier.Verify(context.Background(), token, payload)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestEd25519Verifier_Verify_PayloadMismatch(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	kid := "test-key-v1"

	signer, err := NewEd25519Signer(priv, kid)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewEd25519Verifier(&staticKeyProvider{keys: map[string]ed25519.PublicKey{kid: pub}})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("test payload")
	token, _ := signer.Sign(payload)

	wrongPayload := []byte("wrong payload")
	err = verifier.Verify(context.Background(), token, wrongPayload)
	if err == nil {
		t.Fatal("expected payload mismatch error")
	}
}

func TestEd25519Verifier_Verify_WrongKey(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	kid := "test-key-v1"

	signer, err := NewEd25519Signer(priv, kid)
	if err != nil {
		t.Fatal(err)
	}
	wrongPubKey, _, _ := ed25519.GenerateKey(rand.Reader)
	verifier, err := NewEd25519Verifier(&staticKeyProvider{keys: map[string]ed25519.PublicKey{kid: wrongPubKey}})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("test payload")
	token, _ := signer.Sign(payload)

	err = verifier.Verify(context.Background(), token, payload)
	if err == nil {
		t.Fatal("expected signature verification error with wrong key")
	}
}

func TestEd25519Verifier_Verify_MissingKid(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	kid := "test-key-v1"

	signer, err := NewEd25519Signer(priv, kid)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewEd25519Verifier(&staticKeyProvider{keys: map[string]ed25519.PublicKey{}})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("test payload")
	token, _ := signer.Sign(payload)

	err = verifier.Verify(context.Background(), token, payload)
	if err == nil {
		t.Fatal("expected key-not-found error")
	}
}
