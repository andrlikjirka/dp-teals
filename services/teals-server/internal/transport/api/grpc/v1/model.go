package v1

import (
	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service"
)

func toAppendEventInput(req *ingestionv1.AppendRequest) service.AppendEventInput {
	ev := req.GetEvent()

	return service.AppendEventInput{
		EventID: ev.GetId(),
	}
}

func toAppendResponse(out *service.AppendEventOutput) *ingestionv1.AppendResponse {
	return &ingestionv1.AppendResponse{
		Success: out.Success,
	}
}
