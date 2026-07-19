package domain

import "context"

type Repository interface {
	Save(ctx context.Context, doc Document) (Document, error)
}
