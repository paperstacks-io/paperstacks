package http

import (
	"time"

	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type UserResponse struct {
	ExternalID string    `json:"externalId"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func NewUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ExternalID: user.ExternalID,
		Email:      user.Email,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

type StackResponse struct {
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"ownerId"`
	IsPublic  bool      `json:"isPublic"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewStackResponse(stack stackDomain.Stack) StackResponse {
	return StackResponse{
		UUID:      stack.UUID,
		Name:      stack.Name,
		OwnerID:   stack.Owner.ExternalID,
		IsPublic:  stack.IsPublic,
		CreatedAt: stack.CreatedAt,
		UpdatedAt: stack.UpdatedAt,
	}
}

func NewStackResponses(stacks []stackDomain.Stack) []StackResponse {
	out := make([]StackResponse, 0, len(stacks))
	for _, stack := range stacks {
		out = append(out, NewStackResponse(stack))
	}

	return out
}
