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

// ProofServiceServer implements the gRPC server for the ProofService defined in the protobuf.
type ProofServiceServer struct {
	auditv1.UnimplementedProofServiceServer
	service *service.LedgerService
}

// NewProofServiceServer creates a new instance of ProofServiceServer with the provided LedgerService. This allows the gRPC server to delegate the actual proof generation logic to the service layer, keeping the transport layer focused on handling gRPC requests and responses.
func NewProofServiceServer(s *service.LedgerService) *ProofServiceServer {
	return &ProofServiceServer{
		service: s,
	}
}

// GetInclusionProof handles incoming GetInclusionProofRequest messages, parses the event ID, and calls the service layer to generate an inclusion proof for the specified audit event. It returns a GetInclusionProofResponse with the proof if successful, or an appropriate gRPC error status if the request is invalid or if there was an error during proof generation.
func (s *ProofServiceServer) GetInclusionProof(ctx context.Context, req *auditv1.GetInclusionProofRequest) (*auditv1.GetInclusionProofResponse, error) {
	id, err := uuid.Parse(req.GetEventId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	proof, err := s.service.GetInclusionProof(ctx, id)
	if err != nil {
		if errors.Is(err, svcerrors.ErrAuditLogEntryNotFound) {
			return nil, status.Errorf(codes.NotFound, "audit event %s not found", req.GetEventId())
		}
		return nil, status.Errorf(codes.Internal, "failed to generate inclusion proof for event %s", req.GetEventId())
	}

	return &auditv1.GetInclusionProofResponse{
		EventId:    proof.EventID.String(),
		LeafIndex:  proof.LeafIndex,
		LedgerSize: proof.LedgerSize,
		LeafHash:   proof.LeafEventHash,
		RootHash:   proof.RootHash,
		Proof: &auditv1.InclusionProof{
			Siblings: proof.Proof.Siblings,
			Left:     proof.Proof.Left,
		},
	}, nil
}
