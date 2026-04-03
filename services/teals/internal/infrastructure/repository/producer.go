package repository

import (
	"context"

	"github.com/google/uuid"
)

type ProducerRepository interface {
	RegisterProducer(ctx context.Context, producerID string) error
	GetProducerByID(ctx context.Context, id uuid.UUID) (any, error)
	GetProducerActiveKey(ctx context.Context, id uuid.UUID) (any, error)
}

// TODO: change the interface to struct implementation
