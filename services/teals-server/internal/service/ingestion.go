package service

import (
	"context"
	"fmt"
)

type Service struct {
}

func NewIngestionService() *Service {
	return &Service{}
}

func (*Service) AppendEvent(ctx context.Context, in AppendEventInput) (*AppendEventOutput, error) {
	fmt.Printf("Appending event with ID: %s\n", in.EventID)

	output := &AppendEventOutput{Success: true}
	return output, nil
}
