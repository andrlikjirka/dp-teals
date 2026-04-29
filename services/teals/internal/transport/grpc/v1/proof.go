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
	ledgerService     service.LedgerProver
	checkpointService service.CheckpointProvider
}

// NewProofServiceServer creates a new instance of ProofServiceServer with the provided LedgerService. This allows the gRPC server to delegate the actual proof generation logic to the ledgerService layer, keeping the transport layer focused on handling gRPC requests and responses.
func NewProofServiceServer(s service.LedgerProver, c service.CheckpointProvider) *ProofServiceServer {
	return &ProofServiceServer{
		ledgerService:     s,
		checkpointService: c,
	}
}

// GetInclusionProof handles incoming GetInclusionProofRequest messages, parses the event ID, and calls the ledgerService layer to generate an inclusion proof for the specified audit event. It returns a GetInclusionProofResponse with the proof if successful, or an appropriate gRPC error status if the request is invalid or if there was an error during proof generation.
func (s *ProofServiceServer) GetInclusionProof(ctx context.Context, req *auditv1.GetInclusionProofRequest) (*auditv1.GetInclusionProofResponse, error) {
	id, err := uuid.Parse(req.GetEventId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	proof, err := s.ledgerService.GetInclusionProof(ctx, id, req.GetLedgerSize())
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

// GetConsistencyProof handles incoming GetConsistencyProofRequest messages, extracts the from and to sizes, and calls the ledgerService layer to generate a consistency proof between the specified ledger sizes. It returns a GetConsistencyProofResponse with the proof if successful, or an appropriate gRPC error status if there was an error during proof generation.
func (s *ProofServiceServer) GetConsistencyProof(ctx context.Context, req *auditv1.GetConsistencyProofRequest) (*auditv1.GetConsistencyProofResponse, error) {
	result, err := s.ledgerService.GetConsistencyProof(ctx, req.GetFromSize(), req.GetToSize())
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

// GetLatestSignedCheckpoint handles incoming GetLatestSignedCheckpointRequest messages and calls the checkpoint ledgerService to retrieve the most recently anchored checkpoint. It returns a GetLatestSignedCheckpointResponse with the checkpoint details if successful, or an appropriate gRPC error status if there was an error during retrieval.
func (s *ProofServiceServer) GetLatestSignedCheckpoint(ctx context.Context, req *auditv1.GetLatestSignedCheckpointRequest) (*auditv1.GetLatestSignedCheckpointResponse, error) {
	ch, err := s.checkpointService.GetLatestCheckpoint(ctx)
	if err != nil {
		if errors.Is(err, svcerrors.ErrCheckpointNotFound) {
			return nil, status.Error(codes.NotFound, "no checkpoint exists yet")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve latest signed checkpoint: %v", err)
	}

	return &auditv1.GetLatestSignedCheckpointResponse{
		Checkpoint: &auditv1.Checkpoint{
			Id:             ch.ID.String(),
			Size:           ch.Checkpoint.Size,
			RootHash:       ch.Checkpoint.RootHash,
			AnchoredAt:     ch.Checkpoint.AnchoredAt.String(),
			Kid:            ch.Kid,
			SignatureToken: ch.SignatureToken,
		},
	}, nil
}

// GetServerPublicKey handles incoming GetServerPublicKeyRequest messages and calls the checkpoint ledgerService to retrieve the server's public key and key ID. It returns a GetServerPublicKeyResponse with the public key details if successful, or an appropriate gRPC error status if there was an error during retrieval.
func (s *ProofServiceServer) GetServerPublicKey(ctx context.Context, req *auditv1.GetServerPublicKeyRequest) (*auditv1.GetServerPublicKeyResponse, error) {
	return &auditv1.GetServerPublicKeyResponse{
		PublicKey: s.checkpointService.ServerPublicKey(),
		Kid:       s.checkpointService.ServerKid(),
	}, nil
}
