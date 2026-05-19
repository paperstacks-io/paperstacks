package http

import "github.com/paperstacks.io/paperstacks/internal/stack/domain"

type StackResponse struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

func NewStackResponse(s domain.Stack) StackResponse {
	return StackResponse{
		UUID:     s.UUID,
		Name:     s.Name,
		IsPublic: s.IsPublic,
	}
}

func NewStackResponses(stacks []domain.Stack) []StackResponse {
	resp := make([]StackResponse, len(stacks))

	for i, stack := range stacks {
		resp[i] = NewStackResponse(stack)
	}

	return resp
}
