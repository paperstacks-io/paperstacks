package memory

import (
	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func seedData() []domain.Stack {
	stacks := make([]domain.Stack, 0)
	stacks = append(stacks, *domain.NewStack("Example Stack", userDomain.User{ExternalID: "0"}))
	return stacks
}
