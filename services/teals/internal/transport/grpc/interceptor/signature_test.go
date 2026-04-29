package interceptor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func newTestLogger() *logger.Logger {
	handler := slog.NewTextHandler(io.Discard, nil)
	return &logger.Logger{
		Logger: slog.New(handler),
	}
}

// neverCalledHandler fails the test if the downstream handler is invoked.
func neverCalledHandler(t *testing.T) grpc.UnaryHandler {
	t.Helper()
	return func(ctx context.Context, req any) (any, error) {
		t.Error("handler must not be called")
		return nil, nil
	}
}

// incomingCtx builds a context with gRPC incoming metadata from key/value pairs.
func incomingCtx(pairs ...string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(pairs...))
}

// --- UnaryInterceptor: methods not requiring a signature ---

func TestUnaryInterceptor_UnprotectedMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"query service", "/audit.v1.QueryService/GetAuditEvent"},
		{"list method", "/audit.v1.QueryService/ListAuditEvents"},
		{"unknown service", "/some.Service/Method"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := NewSignatureInterceptor(newTestLogger())
			info := &grpc.UnaryServerInfo{FullMethod: tc.method}

			called := false
			handler := func(ctx context.Context, req any) (any, error) {
				called = true
				return nil, nil
			}

			// No metadata in context — must not matter for unprotected methods.
			_, err := i.UnaryInterceptor(context.Background(), nil, info, handler)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !called {
				t.Error("expected handler to be called")
			}
		})
	}
}

// --- UnaryInterceptor: method requires a signature, error paths ---

func TestUnaryInterceptor_ProtectedMethod_Errors(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantCode codes.Code
	}{
		{
			name:     "no gRPC metadata in context",
			ctx:      context.Background(),
			wantCode: codes.Internal,
		},
		{
			name:     "metadata present but signature header absent",
			ctx:      incomingCtx("unrelated-header", "value"),
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := NewSignatureInterceptor(newTestLogger())
			info := &grpc.UnaryServerInfo{FullMethod: "/audit.v1.IngestionService/Append"}

			_, err := i.UnaryInterceptor(tc.ctx, nil, info, neverCalledHandler(t))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got %T: %v", err, err)
			}
			if st.Code() != tc.wantCode {
				t.Errorf("code: got %v, want %v", st.Code(), tc.wantCode)
			}
		})
	}
}

// --- UnaryInterceptor: method requires a signature, happy path ---

func TestUnaryInterceptor_ProtectedMethod_SignatureInjectedIntoContext(t *testing.T) {
	expectedToken := "my-token"
	i := NewSignatureInterceptor(newTestLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/audit.v1.IngestionService/Append"}

	ctx := incomingCtx(SignatureHeaderKey, expectedToken)

	var capturedCtx context.Context
	handler := func(ctx context.Context, req any) (any, error) {
		capturedCtx = ctx
		return nil, nil
	}

	_, err := i.UnaryInterceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedCtx == nil {
		t.Fatal("handler was not called")
	}

	sig, ok := SignatureFromContext(capturedCtx)
	if !ok {
		t.Fatal("expected signature in handler context")
	}
	if sig.Token != expectedToken {
		t.Errorf("token: got %q, want %q", sig.Token, expectedToken)
	}
}

func TestUnaryInterceptor_ProtectedMethod_HandlerReturnValuePropagated(t *testing.T) {
	expectedResponse := "response-payload"
	i := NewSignatureInterceptor(newTestLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/audit.v1.IngestionService/Append"}

	ctx := incomingCtx(SignatureHeaderKey, "tok")
	handler := func(ctx context.Context, req any) (any, error) {
		return expectedResponse, nil
	}

	resp, err := i.UnaryInterceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != expectedResponse {
		t.Errorf("response: got %v, want %q", resp, expectedResponse)
	}
}

func TestUnaryInterceptor_ProtectedMethod_HandlerErrorPropagated(t *testing.T) {
	expectedErr := status.Error(codes.FailedPrecondition, "handler failed")
	i := NewSignatureInterceptor(newTestLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/audit.v1.IngestionService/Append"}

	ctx := incomingCtx(SignatureHeaderKey, "tok")
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, expectedErr
	}

	_, err := i.UnaryInterceptor(ctx, nil, info, handler)
	if !errors.Is(err, expectedErr) {
		t.Errorf("error: got %v, want %v", err, expectedErr)
	}
}
