package interceptor

import (
	"context"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const SignatureHeaderKey = "x-jws-event-signature"

var requiresSignature = map[string]bool{
	"/audit.v1.IngestionService/Append": true,
}

// SignatureInterceptor is a gRPC interceptor that validates JWS signatures on incoming requests for audit events.
type SignatureInterceptor struct {
	logger *logger.Logger
}

// NewSignatureInterceptor creates a new SignatureInterceptor with the provided JWS verifier.
func NewSignatureInterceptor(log *logger.Logger) *SignatureInterceptor {
	return &SignatureInterceptor{logger: log}
}

// UnaryInterceptor is a gRPC unary interceptor that checks for the presence of a JWS signature in the incoming request metadata for specific methods. If the method requires a signature, it extracts the token and adds it to the context for downstream handlers to use. If the signature is missing or if there is an error extracting it, the interceptor returns an appropriate gRPC error status.
func (i *SignatureInterceptor) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if !requiresSignature[info.FullMethod] {
		return handler(ctx, req)
	}

	// 1. Extract token from incoming metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		i.logger.Error("failed to extract gRPC metadata from context", "method", info.FullMethod)
		return nil, status.Error(codes.Internal, "missing metadata in context")
	}
	values := md.Get(SignatureHeaderKey)
	if len(values) == 0 {
		i.logger.Warn("request rejected: missing JWS signature header", "method", info.FullMethod)
		return nil, status.Error(codes.Unauthenticated, "missing x-jws-signature header")
	}
	token := values[0]

	// 4. Delegate to the next handler in the chain.
	ctx = ContextWithSignature(ctx, token)
	return handler(ctx, req)
}
