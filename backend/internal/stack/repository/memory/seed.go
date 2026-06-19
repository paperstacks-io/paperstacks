package memory

import (
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const demoUserExternalID = "dbe3febc-ab91-486c-b51f-38ab0f59a4d9"

func seedData() []domain.Stack {
	return []domain.Stack{
		{
			UUID:     "9e1a819a-24ab-47b6-be29-92b49325e4c2",
			Name:     "Code Review",
			IsPublic: true,
			Owner: userDomain.User{
				ExternalID: demoUserExternalID,
			},
		},
		{
			UUID:     "a8de9118-e7d9-4a0b-9e77-7872f08d8efa",
			Name:     "Testing",
			IsPublic: true,
			Owner: userDomain.User{
				ExternalID: demoUserExternalID,
			},
		},
		{
			UUID:     "c6ff032d-104f-4f5f-a9d7-87f874c75c0a",
			Name:     "Secondary Studies",
			IsPublic: true,
			Owner: userDomain.User{
				ExternalID: demoUserExternalID,
			},
		},
		{
			UUID:     "d67f909d-84a3-4c4e-823c-0c9a20e89790",
			Name:     "Bayesian",
			IsPublic: true,
			Owner: userDomain.User{
				ExternalID: demoUserExternalID,
			},
		},
		{
			UUID:     "f4152eb2-b303-461c-a683-5bfe80258f8e",
			Name:     "Research Methodology",
			IsPublic: true,
			Owner: userDomain.User{
				ExternalID: demoUserExternalID,
			},
		},
	}
}
