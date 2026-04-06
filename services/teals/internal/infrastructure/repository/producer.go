package repository

import (
	"context"

	"github.com/google/uuid"
)

type ProducerRepository interface {
	GetProducerByID(ctx context.Context, id uuid.UUID) (any, error)
}

// TODO: change the interface to struct implementation
