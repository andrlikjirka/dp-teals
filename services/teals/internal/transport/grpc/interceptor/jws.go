package interceptor

import (
	"context"
	"fmt"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	pkgjws "github.com/andrlikjirka/dp-teals/pkg/jws"
	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const SignatureHeaderKey = "x-jws-event-signature"

var requiresSignature = map[string]bool{
	"/audit.v1.IngestionService/Append": true,
}

// JwsAuditInterceptor is a gRPC interceptor that validates JWS signatures on incoming requests for audit events.
type JwsAuditInterceptor struct {
	verifier pkgjws.Verifier
	logger   *logger.Logger
}

// NewJwsInterceptor creates a new JwsAuditInterceptor with the provided JWS verifier.
func NewJwsInterceptor(verifier pkgjws.Verifier, log *logger.Logger) *JwsAuditInterceptor {
	return &JwsAuditInterceptor{verifier: verifier, logger: log}
}

// UnaryInterceptor returns a new unary server interceptor for JWS validation.
func (i *JwsAuditInterceptor) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
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

	// 2. Re-marshal the proto message deterministically.
	protoReq, ok := req.(*auditv1.AppendRequest)
	if !ok {
		i.logger.Error("unexpected request type during JWS verification",
			"method", info.FullMethod,
			"type", fmt.Sprintf("%T", req),
		)
		return nil, status.Errorf(codes.Internal, "unexpected request type: %T", req)
	}

	opts := proto.MarshalOptions{Deterministic: true}
	payload, err := opts.Marshal(protoReq.GetEvent())
	if err != nil {
		i.logger.Error("failed to marshal audit event for signature verification",
			"method", info.FullMethod,
			"event_id", protoReq.GetEvent().GetId(),
			"error", err,
		)
		return nil, status.Error(codes.Internal, "failed to marshal event")
	}

	// 3. Verify the JWS signature using the verifier.
	kid, err := i.verifier.Verify(ctx, token, payload)
	if err != nil {
		i.logger.Warn("request rejected: invalid JWS signature",
			"method", info.FullMethod,
			"event_id", protoReq.GetEvent().GetId(),
			"error", err,
		)
		return nil, status.Errorf(codes.Unauthenticated, "invalid audit event signature: %v", err)
	}

	i.logger.Info("JWS signature verified",
		"method", info.FullMethod,
		"event_id", protoReq.GetEvent().GetId(),
	)

	// 4. Delegate to the next handler in the chain.
	ctx = ContextWithSignature(ctx, token, kid)
	return handler(ctx, req)
}
