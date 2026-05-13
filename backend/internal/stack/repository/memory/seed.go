package memory

import (
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func seedData() []domain.Stack {
	return []domain.Stack{
		{
			UUID: "Test-UUID-00",
			Name: "Example Stack",
			Owner: userDomain.User{
				ExternalID: "0",
			},
		},
	}
}
