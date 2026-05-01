package v1

import (
	"context"
	"errors"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DataSubjectServiceServer implements the gRPC server for the DataSubjectService defined in the protobuf. It contains a reference to the SubjectForgetter service, which is used to perform the actual subject forgetting logic when handling gRPC requests.
type DataSubjectServiceServer struct {
	auditv1.UnimplementedDataSubjectServiceServer
	service service.SubjectForgetter
}

// NewDataSubjectServiceServer creates a new instance of DataSubjectServiceServer with the provided SubjectForgetter service. This allows the gRPC server to delegate the actual subject forgetting logic to the service layer, keeping the transport layer focused on handling gRPC requests and responses.
func NewDataSubjectServiceServer(s service.SubjectForgetter) *DataSubjectServiceServer {
	return &DataSubjectServiceServer{
		service: s,
	}
}

// ForgetSubject handles incoming ForgetSubjectRequest messages, calls the service layer to perform the subject forgetting logic, and returns an appropriate gRPC response. It checks for specific errors from the service layer to return meaningful gRPC status codes and messages.
func (s *DataSubjectServiceServer) ForgetSubject(ctx context.Context, req *auditv1.ForgetSubjectRequest) (*auditv1.ForgetSubjectResponse, error) {
	resp, err := s.service.ForgetSubject(ctx, req.GetSubjectId())
	if err != nil {
		if errors.Is(err, svcerrors.ErrMissingSubjectID) {
			return nil, status.Error(codes.InvalidArgument, "subject_id is required")
		}
		if errors.Is(err, svcerrors.ErrSubjectSecretNotFound) {
			return nil, status.Errorf(codes.NotFound, "no secret found for subject %s", req.GetSubjectId())
		}
		return nil, status.Error(codes.Internal, "failed to forget subject")
	}
	return &auditv1.ForgetSubjectResponse{
		SubjectId:   resp.SubjectID,
		ForgottenAt: timestamppb.New(resp.ForgottenAt),
	}, nil
}
