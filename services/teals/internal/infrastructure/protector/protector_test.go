package protector

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func randomKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

// --- NewAesGcmProtector ---

func TestNewAesGcmProtector_ValidKey(t *testing.T) {
	_, err := NewAesGcmProtector(randomKey(t))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewAesGcmProtector_ShortKey(t *testing.T) {
	_, err := NewAesGcmProtector(make([]byte, 16))
	if err == nil {
		t.Fatal("expected error for 16-byte key, got nil")
	}
}

func TestNewAesGcmProtector_LongKey(t *testing.T) {
	_, err := NewAesGcmProtector(make([]byte, 64))
	if err == nil {
		t.Fatal("expected error for 64-byte key, got nil")
	}
}

func TestNewAesGcmProtector_EmptyKey(t *testing.T) {
	_, err := NewAesGcmProtector([]byte{})
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

// --- Protect ---

func TestProtect_ReturnsNonNilResult(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"name": "Alice", "age": 30}

	protected, salt, err := p.Protect(secret, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if protected == nil {
		t.Fatal("expected non-nil ProtectedMetadata")
	}
	if len(salt) != saltSize {
		t.Errorf("expected salt length %d, got %d", saltSize, len(salt))
	}
}

func TestProtect_FieldsPopulated(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	protected, salt, err := p.Protect(secret, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(protected.Ciphertext) == 0 {
		t.Error("ciphertext is empty")
	}
	if len(protected.WrappedDEK) == 0 {
		t.Error("wrapped DEK is empty")
	}
	if len(protected.Commitment) == 0 {
		t.Error("commitment is empty")
	}
	if len(salt) == 0 {
		t.Error("salt is empty")
	}
}

func TestProtect_DifferentSaltsOnEachCall(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	_, salt1, _ := p.Protect(secret, meta)
	_, salt2, _ := p.Protect(secret, meta)

	if bytes.Equal(salt1, salt2) {
		t.Error("expected different salts on each Protect call")
	}
}

func TestProtect_DifferentCiphertextsOnEachCall(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	pm1, _, _ := p.Protect(secret, meta)
	pm2, _, _ := p.Protect(secret, meta)

	if bytes.Equal(pm1.Ciphertext, pm2.Ciphertext) {
		t.Error("expected different ciphertexts on each Protect call (randomized DEK/nonce)")
	}
}

func TestProtect_DifferentCommitmentsOnEachCall(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	pm1, _, _ := p.Protect(secret, meta)
	pm2, _, _ := p.Protect(secret, meta)

	// Commitments bind to the random salt, so they must differ.
	if bytes.Equal(pm1.Commitment, pm2.Commitment) {
		t.Error("expected different commitments on each Protect call due to random salt")
	}
}

func TestProtect_EmptyMetadata(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{}

	protected, salt, err := p.Protect(secret, meta)
	if err != nil {
		t.Fatalf("unexpected error with empty metadata: %v", err)
	}
	if protected == nil || len(salt) == 0 {
		t.Error("expected non-nil result for empty metadata")
	}
}

// --- Reveal ---

func TestReveal_RoundTrip(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"name": "Bob", "role": "admin"}

	protected, _, err := p.Protect(secret, meta)
	if err != nil {
		t.Fatalf("Protect error: %v", err)
	}

	revealed, err := p.Reveal(secret, protected)
	if err != nil {
		t.Fatalf("Reveal error: %v", err)
	}

	if revealed["name"] != meta["name"] {
		t.Errorf("name mismatch: got %v, want %v", revealed["name"], meta["name"])
	}
	if revealed["role"] != meta["role"] {
		t.Errorf("role mismatch: got %v, want %v", revealed["role"], meta["role"])
	}
}

func TestReveal_WrongSecret(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	meta := map[string]any{"key": "value"}

	protected, _, err := p.Protect([]byte("correct-secret"), meta)
	if err != nil {
		t.Fatalf("Protect error: %v", err)
	}

	_, err = p.Reveal([]byte("wrong-secret"), protected)
	if err == nil {
		t.Fatal("expected error when revealing with wrong secret, got nil")
	}
}

func TestReveal_WrongMasterKEK(t *testing.T) {
	p1, _ := NewAesGcmProtector(randomKey(t))
	p2, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	protected, _, err := p1.Protect(secret, meta)
	if err != nil {
		t.Fatalf("Protect error: %v", err)
	}

	_, err = p2.Reveal(secret, protected)
	if err == nil {
		t.Fatal("expected error when revealing with different master KEK, got nil")
	}
}

func TestReveal_TamperedCiphertext(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	protected, _, err := p.Protect(secret, meta)
	if err != nil {
		t.Fatalf("Protect error: %v", err)
	}

	protected.Ciphertext[len(protected.Ciphertext)-1] ^= 0xFF // flip the last byte of the ciphertext to simulate tampering

	_, err = p.Reveal(secret, protected)
	if err == nil {
		t.Fatal("expected authentication error for tampered ciphertext, got nil")
	}
}

func TestReveal_TamperedWrappedDEK(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("subject-secret")
	meta := map[string]any{"key": "value"}

	protected, _, err := p.Protect(secret, meta)
	if err != nil {
		t.Fatalf("Protect error: %v", err)
	}

	protected.WrappedDEK[len(protected.WrappedDEK)-1] ^= 0xFF

	_, err = p.Reveal(secret, protected)
	if err == nil {
		t.Fatal("expected authentication error for tampered wrapped DEK, got nil")
	}
}

func TestReveal_NumericValuesPreservedAsFloat64(t *testing.T) {
	p, _ := NewAesGcmProtector(randomKey(t))
	secret := []byte("s")
	meta := map[string]any{"count": float64(42)}

	protected, _, _ := p.Protect(secret, meta)
	revealed, err := p.Reveal(secret, protected)
	if err != nil {
		t.Fatalf("Reveal error: %v", err)
	}
	if revealed["count"] != float64(42) {
		t.Errorf("count mismatch: got %v (%T)", revealed["count"], revealed["count"])
	}
}

// --- aesGcmSeal / aesGcmOpen ---

func TestAesGcmSealOpen_RoundTrip(t *testing.T) {
	key := randomKey(t)
	plaintext := []byte("hello, world")

	sealed, err := aesGcmSeal(key, plaintext)
	if err != nil {
		t.Fatalf("seal error: %v", err)
	}

	opened, err := aesGcmOpen(key, sealed)
	if err != nil {
		t.Fatalf("open error: %v", err)
	}

	if !bytes.Equal(opened, plaintext) {
		t.Errorf("plaintext mismatch: got %q, want %q", opened, plaintext)
	}
}

func TestAesGcmOpen_TooShort(t *testing.T) {
	key := randomKey(t)
	_, err := aesGcmOpen(key, make([]byte, nonceSize-1))
	if err == nil {
		t.Fatal("expected error for sealed data shorter than nonce, got nil")
	}
}

func TestAesGcmOpen_WrongKey(t *testing.T) {
	key1 := randomKey(t)
	key2 := randomKey(t)
	plaintext := []byte("secret data")

	sealed, err := aesGcmSeal(key1, plaintext)
	if err != nil {
		t.Fatalf("seal error: %v", err)
	}

	_, err = aesGcmOpen(key2, sealed)
	if err == nil {
		t.Fatal("expected authentication error with wrong key, got nil")
	}
}

func TestAesGcmSeal_NonDeterministic(t *testing.T) {
	key := randomKey(t)
	plaintext := []byte("same input")

	s1, _ := aesGcmSeal(key, plaintext)
	s2, _ := aesGcmSeal(key, plaintext)

	if bytes.Equal(s1, s2) {
		t.Error("expected different sealed outputs due to random nonce")
	}
}

// --- deriveKey ---

func TestDeriveKey_Deterministic(t *testing.T) {
	ikm := randomKey(t)
	salt := []byte("fixed-salt")
	info := []byte("test-info")

	k1, err := deriveKey(ikm, salt, info)
	if err != nil {
		t.Fatalf("deriveKey error: %v", err)
	}
	k2, err := deriveKey(ikm, salt, info)
	if err != nil {
		t.Fatalf("deriveKey error: %v", err)
	}

	if !bytes.Equal(k1, k2) {
		t.Error("expected deterministic key derivation with same inputs")
	}
}

func TestDeriveKey_DifferentSaltsProduceDifferentKeys(t *testing.T) {
	ikm := randomKey(t)
	info := []byte("test-info")

	k1, _ := deriveKey(ikm, []byte("salt-a"), info)
	k2, _ := deriveKey(ikm, []byte("salt-b"), info)

	if bytes.Equal(k1, k2) {
		t.Error("expected different keys for different salts")
	}
}

func TestDeriveKey_DifferentIKMsProduceDifferentKeys(t *testing.T) {
	salt := []byte("fixed-salt")
	info := []byte("test-info")

	k1, _ := deriveKey(randomKey(t), salt, info)
	k2, _ := deriveKey(randomKey(t), salt, info)

	if bytes.Equal(k1, k2) {
		t.Error("expected different keys for different IKMs")
	}
}

func TestDeriveKey_OutputLength(t *testing.T) {
	k, err := deriveKey(randomKey(t), []byte("salt"), []byte("info"))
	if err != nil {
		t.Fatalf("deriveKey error: %v", err)
	}
	if len(k) != dekLength {
		t.Errorf("expected key length %d, got %d", dekLength, len(k))
	}
}
