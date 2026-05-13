package memory

import (
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func seedData() []domain.Stack {
	return []domain.Stack{
		{
			UUID: "00000000-0000-0000-0000-000000000001",
			Name: "Example Stack",
			Owner: userDomain.User{
				ExternalID: "0",
			},
		},
	}
}
