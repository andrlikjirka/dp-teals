package interceptor

import (
	"context"
	"testing"
)

func TestSignatureFromContext_RoundTrip(t *testing.T) {
	expectedToken := "test-token"

	ctx := context.Background()
	ctx = ContextWithSignature(ctx, expectedToken)

	sig, ok := SignatureFromContext(ctx)
	if !ok {
		t.Fatal("Expected signature to be present in context")
	}
	if sig.Token != expectedToken {
		t.Fatalf("Expected token to be 'test-token', got '%s'", sig.Token)
	}
}

func TestSignatureFromContext_MissingKey(t *testing.T) {
	ctx := context.Background()
	_, ok := SignatureFromContext(ctx)
	if ok {
		t.Fatal("Expected signature to be absent in context")
	}
}

func TestContextWithSignature_TokenEmpty(t *testing.T) {
	ctx := ContextWithSignature(context.Background(), "")
	sig, ok := SignatureFromContext(ctx)
	if !ok {
		t.Fatal("Expected signature to be present in context")
	}
	if sig.Token != "" {
		t.Fatalf("Expected token to be empty, got '%s'", sig.Token)
	}
}

func TestSignatureFromContext_DerivedContext(t *testing.T) {
	expectedToken := "test-token"
	ctx := ContextWithSignature(context.Background(), expectedToken)
	derived, cancel := context.WithCancel(ctx)
	defer cancel()

	sig, ok := SignatureFromContext(derived)
	if !ok {
		t.Fatal("expected signature in derived context")
	}
	if sig.Token != expectedToken {
		t.Errorf("token: got %q, want %q", sig.Token, "token")
	}
}
