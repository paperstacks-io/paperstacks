package memory

import (
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func seedData() []domain.Stack {
	return []domain.Stack{
		{
			UUID: "9e1a819a-24ab-47b6-be29-92b49325e4c2",
			Name: "Example Stack",
			Owner: userDomain.User{
				ExternalID: "dbe3febc-ab91-486c-b51f-38ab0f59a4d9",
			},
		},
	}
}
