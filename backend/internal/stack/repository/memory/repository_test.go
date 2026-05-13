package memory

import (
	"context"
	"testing"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func TestRepositoryCreateReturnsAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.Create(context.Background(), domain.Stack{
		Name: "Example Stack",
		Owner: userDomain.User{
			ExternalID: "0",
		},
	})

	if err == domain.ErrStackAlreadyExists {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrStackAlreadyExists)
	}
}
