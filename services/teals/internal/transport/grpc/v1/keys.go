package v1

import (
	"context"
	"errors"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// KeyRegistrationServiceServer implements the gRPC server for the KeyRegistrationService defined in the protobuf. It provides an endpoint for producers to register their public keys, which are necessary for signing audit events.
type KeyRegistrationServiceServer struct {
	auditv1.UnimplementedKeyRegistrationServiceServer
	service *service.KeyService
}

// NewKeyRegistrationServiceServer creates a new instance of KeyRegistrationServiceServer with the provided KeyService. This allows the gRPC server to delegate the actual key registration logic to the service layer, keeping the transport layer focused on handling gRPC requests and responses.
func NewKeyRegistrationServiceServer(s *service.KeyService) *KeyRegistrationServiceServer {
	return &KeyRegistrationServiceServer{
		service: s,
	}
}

// RegisterKey handles incoming RegisterKeyRequest messages, validates the producer ID and public key, and calls the service layer to register the key. It returns a RegisterKeyResponse with the key ID (kid) if successful, or an appropriate gRPC error status if the request is invalid or if there was an error during registration.
func (s *KeyRegistrationServiceServer) RegisterKey(ctx context.Context, req *auditv1.RegisterKeyRequest) (*auditv1.RegisterKeyResponse, error) {
	producerId, err := uuid.Parse(req.GetProducerId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid producer_id: %v", err)
	}

	kid, err := s.service.RegisterProducerKey(ctx, producerId, req.GetPublicKey())
	if err != nil {
		switch {
		case errors.Is(err, svcerrors.ErrInvalidPublicKey):
			return nil, status.Error(codes.InvalidArgument, "invalid producer public key")
		case errors.Is(err, svcerrors.ErrDuplicateProducerKey):
			return nil, status.Error(codes.AlreadyExists, "producer public key already registered")
		default:
			return nil, status.Error(codes.Internal, "failed to register producer public key")
		}
	}

	return &auditv1.RegisterKeyResponse{KeyId: kid}, nil
}
