package interceptor

import "context"

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// signatureContextKey is the key used to store signature information in the context.
const signatureContextKey contextKey = "x-audit-event-signature"

// SignatureContext holds the JWS signature information extracted from the gRPC metadata.
type SignatureContext struct {
	Token string
}

// ContextWithSignature adds the JWS signature information to the context. It takes a context and a JWS token as input and returns a new context with the signature information stored under the signatureContextKey.
func ContextWithSignature(ctx context.Context, token string) context.Context {
	ctx = context.WithValue(ctx, signatureContextKey, &SignatureContext{
		Token: token,
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
