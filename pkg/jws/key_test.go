package jws

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestThumbprint_ValidKey(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	thumb, err := Thumbprint(pub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if thumb == "" {
		t.Fatal("expected non-empty thumbprint")
	}
}

func TestThumbprint_Deterministic(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t1, err := Thumbprint(pub)
	if err != nil {
		t.Fatal(err)
	}
	t2, err := Thumbprint(pub)
	if err != nil {
		t.Fatal(err)
	}
	if t1 != t2 {
		t.Fatalf("thumbprint is not deterministic: %q != %q", t1, t2)
	}
}

func TestThumbprint_DifferentKeys_DifferentThumprints(t *testing.T) {
	pub1, _, _ := ed25519.GenerateKey(rand.Reader)
	pub2, _, _ := ed25519.GenerateKey(rand.Reader)
	t1, err := Thumbprint(pub1)
	if err != nil {
		t.Fatal(err)
	}
	t2, err := Thumbprint(pub2)
	if err != nil {
		t.Fatal(err)
	}
	if t1 == t2 {
		t.Fatal("different keys produced the same thumbprint")
	}
}
