package http

import (
	"time"

	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type UserResponse struct {
	ExternalID string    `json:"external_id"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
	OwnerID   string    `json:"owner_id"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
