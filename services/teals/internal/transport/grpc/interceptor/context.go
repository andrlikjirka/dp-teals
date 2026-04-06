package interceptor

import "context"

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// signatureContextKey is the key used to store signature information in the context.
const signatureContextKey contextKey = "x-audit-event-signature"

// SignatureContext holds the JWS signature information extracted from the gRPC metadata.
type SignatureContext struct {
	Token string
	KeyID string
}

// ContextWithSignature creates a new context with the provided JWS token and key ID for signature verification.
func ContextWithSignature(ctx context.Context, token string, kid string) context.Context {
	ctx = context.WithValue(ctx, signatureContextKey, &SignatureContext{
		Token: token,
		KeyID: kid,
	})
	return ctx
}

// SignatureFromContext retrieves the JWS signature information from the context. It returns the SignatureContext and a boolean indicating whether the signature information was present in the context.
func SignatureFromContext(ctx context.Context) (*SignatureContext, bool) {
	value := ctx.Value(signatureContextKey)
	if value == nil {
		return nil, false
	}

	signatureContext, ok := value.(*SignatureContext)
	return signatureContext, ok
}
