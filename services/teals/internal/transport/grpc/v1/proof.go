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

	proof, err := s.service.GetInclusionProof(ctx, id, req.GetLedgerSize())
	if err != nil {
		if errors.Is(err, svcerrors.ErrAuditLogEntryNotFound) {
			return nil, status.Errorf(codes.NotFound, "audit event %s not found", req.GetEventId())
		}
		if errors.Is(err, svcerrors.ErrInvalidInclusionProofLedgerSize) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid inclusion proof ledger size: size must be gte leaf position and lte current ledger size")
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

// GetConsistencyProof handles incoming GetConsistencyProofRequest messages, extracts the from and to sizes, and calls the service layer to generate a consistency proof between the specified ledger sizes. It returns a GetConsistencyProofResponse with the proof if successful, or an appropriate gRPC error status if there was an error during proof generation.
func (s *ProofServiceServer) GetConsistencyProof(ctx context.Context, req *auditv1.GetConsistencyProofRequest) (*auditv1.GetConsistencyProofResponse, error) {
	result, err := s.service.GetConsistencyProof(ctx, req.GetFromSize(), req.GetToSize())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate consistency proof: %v", err)
	}
	protoPaths := make([]*auditv1.ConsistencyPath, len(result.Proof.ConsistencyPaths))
	for i, p := range result.Proof.ConsistencyPaths {
		protoPaths[i] = &auditv1.ConsistencyPath{Siblings: p.Siblings, Left: p.Left}
	}

	return &auditv1.GetConsistencyProofResponse{
		Proof: &auditv1.ConsistencyProof{
			OldSize:          int64(result.Proof.OldSize),
			NewSize:          int64(result.Proof.NewSize),
			OldPeaksHashes:   result.Proof.OldPeaksHashes,
			ConsistencyPaths: protoPaths,
			RightPeaks:       result.Proof.RightPeaks,
		},
	}, nil
}

// GetRootHash handles incoming GetRootHashRequest messages and calls the service layer to retrieve the current root hash of the MMR ledger. It returns a GetRootHashResponse with the root hash if successful, or an appropriate gRPC error status if there was an error retrieving the root hash.
func (s *ProofServiceServer) GetRootHash(ctx context.Context, req *auditv1.GetRootHashRequest) (*auditv1.GetRootHashResponse, error) {
	root, err := s.service.GetRootHash(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get root hash: %v", err)
	}

	return &auditv1.GetRootHashResponse{RootHash: root}, nil
}
