package protector

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha3"
	"encoding/json"
	"fmt"
	"hash"
	"io"

	pkgcannon "github.com/andrlikjirka/dp-teals/pkg/canonical"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"golang.org/x/crypto/hkdf"
)

const (
	dekLength = 32 // 256 bits
	nonceSize = 12 // 96 bits
	saltSize  = 32 // 256 bits
	kekInfo   = "teals-subject-kek"
)

// AesGcmProtector implements the MetadataProtector interface using AES-256-GCM for encryption and HKDF for key derivation. It uses a master KEK to derive subject-specific KEKs for encrypting DEKs, which in turn are used to encrypt metadata. The protector ensures that the same metadata encrypted with the same secret will produce different ciphertexts due to the use of random salts and DEKs, while still allowing for deterministic commitment generation.
type AesGcmProtector struct {
	masterKEK []byte
}

// NewAesGcmProtector creates a new AesGcmProtector with the provided master KEK. The master KEK must be 32 bytes long (for AES-256). An error is returned if the master KEK does not meet the required length.
func NewAesGcmProtector(masterKEK []byte) (*AesGcmProtector, error) {
	if len(masterKEK) != dekLength {
		return nil, fmt.Errorf("master KEK must be 32 bytes, got %d", len(masterKEK))
	}
	return &AesGcmProtector{masterKEK: masterKEK}, nil
}

// Protect encrypts the given metadata using AES-256-GCM. It derives a subject-specific KEK using HKDF with the master KEK, the provided secret, and a fixed info string. A random DEK is generated to encrypt the metadata, and then the DEK is wrapped with the subject KEK. The function returns the protected metadata containing the ciphertext, wrapped DEK, and commitment, along with the salt used for key derivation. An error is returned if any step of the protection process fails.
func (p *AesGcmProtector) Protect(secret []byte, metadata map[string]any) (*svcmodel.ProtectedMetadata, []byte, error) {
	canonicalMeta, err := pkgcannon.CanonicalizeMetadata(metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("canonicalize metadata: %w", err)
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, fmt.Errorf("generate salt: %w", err)
	}
	h := sha3.New256()
	if _, err = h.Write(canonicalMeta); err != nil {
		return nil, nil, err
	}
	if _, err = h.Write(salt); err != nil {
		return nil, nil, err
	}
	commitment := h.Sum(nil)

	subjectKEK, err := deriveKey(p.masterKEK, secret, []byte(kekInfo))
	if err != nil {
		return nil, nil, fmt.Errorf("derive subject kek: %w", err)
	}

	dek := make([]byte, dekLength)
	if _, err := rand.Read(dek); err != nil {
		return nil, nil, fmt.Errorf("generate DEK: %w", err)
	}

	ciphertext, err := aesGcmSeal(dek, canonicalMeta)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt metadata: %w", err)
	}

	wrappedDEK, err := aesGcmSeal(subjectKEK, dek)
	if err != nil {
		return nil, nil, fmt.Errorf("wrap dek: %w", err)
	}

	return &svcmodel.ProtectedMetadata{
		Ciphertext: ciphertext,
		WrappedDEK: wrappedDEK,
		Commitment: commitment,
	}, salt, nil
}

// Reveal decrypts the protected metadata using the provided secret. It derives the subject KEK using HKDF with the master KEK, the secret, and a fixed info string. The DEK is unwrapped using the subject KEK, and then the metadata is decrypted using the DEK. The function returns the decrypted metadata as a map, or an error if any step of the reveal process fails, such as key derivation, DEK unwrapping, metadata decryption, or JSON unmarshalling.
func (p *AesGcmProtector) Reveal(secret []byte, metadata *svcmodel.ProtectedMetadata) (map[string]any, error) {
	subjectKEK, err := deriveKey(p.masterKEK, secret, []byte(kekInfo))
	if err != nil {
		return nil, fmt.Errorf("derive subject kek: %w", err)
	}

	dek, err := aesGcmOpen(subjectKEK, metadata.WrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("unwrap dek: %w", err)
	}

	decryptedMeta, err := aesGcmOpen(dek, metadata.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt metadata: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(decryptedMeta, &m); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return m, nil
}

// deriveKey derives a DEK using HKDF with the given input key material, salt (key isolation), and info. It returns the derived key or an error if key derivation fails.
func deriveKey(ikm, salt, info []byte) ([]byte, error) {
	r := hkdf.New(func() hash.Hash { return sha3.New256() }, ikm, salt, info)
	key := make([]byte, dekLength)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, err
	}
	return key, nil
}

// aesGcmSeal encrypts the plaintext using AES-256-GCM with the provided key. It generates a random nonce, performs encryption, and returns the combined nonce and ciphertext. An error is returned if any step of the encryption process fails.
func aesGcmSeal(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return append(nonce, gcm.Seal(nil, nonce, plaintext, nil)...), nil
}

// aesGcmOpen decrypts the sealed data using AES-256-GCM with the provided key. It extracts the nonce from the beginning of the sealed data, performs decryption, and returns the plaintext. An error is returned if the sealed data is too short, if any step of the decryption process fails, or if authentication fails.
func aesGcmOpen(key, sealed []byte) ([]byte, error) {
	if len(sealed) < nonceSize {
		return nil, fmt.Errorf("sealed data too short: %d bytes", len(sealed))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
}
