package paper

import (
	"context"
	"errors"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

var (
	ErrPaperNotFound      = errors.New("not found")
	ErrPaperAlreadyExists = errors.New("paper already exists")
)

type Repository interface {
	Create(ctx context.Context, paper domain.Paper) error
	ReadAll(ctx context.Context) ([]domain.Paper, error)
	Read(ctx context.Context, id string) (domain.Paper, error)
	Update(ctx context.Context, id string, paper domain.Paper) error
	Delete(ctx context.Context, id string) error
}
