package v1

import (
	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/application/ingestion"
)

func toAppendEventInput(req *ingestionv1.AppendRequest) ingestion.AppendEventInput {
	ev := req.GetEvent()

	return ingestion.AppendEventInput{
		EventID: ev.GetId(),
	}
}

func toAppendResponse(out *ingestion.AppendEventOutput) *ingestionv1.AppendResponse {
	return &ingestionv1.AppendResponse{
		Success: out.Success,
	}
}
