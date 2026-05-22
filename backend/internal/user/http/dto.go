package http

import (
	"time"

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
