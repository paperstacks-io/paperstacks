package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func (s *UserService) fetchUserFromAuthProvider(ctx context.Context, token string) (domain.User, error) {
	if s.authAPIURL == "" {
		return domain.User{}, fmt.Errorf("validate session: %w", domain.ErrInvalidAuthToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.authAPIURL+"/me", nil)
	if err != nil {
		return domain.User{}, fmt.Errorf("create session request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return domain.User{}, fmt.Errorf("send session request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return domain.User{}, fmt.Errorf("validate session: unexpected status %d: %w", res.StatusCode, domain.ErrInvalidAuthToken)
	}

	var response meResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return domain.User{}, fmt.Errorf("decode session response: %w", err)
	}

	user := response.ToUser()
	return user, nil
}

type meResponse struct {
	UserID string `json:"user_id"`

	Emails []struct {
		ID         string `json:"id"`
		Address    string `json:"address"`
		IsVerified bool   `json:"is_verified"`
		IsPrimary  bool   `json:"is_primary"`
	} `json:"emails"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Metadata struct {
		PublicMetadata map[string]interface{} `json:"public_metadata"`
		UnsafeMetadata map[string]interface{} `json:"unsafe_metadata"`
	} `json:"metadata"`

	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`

	Username struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Username  string    `json:"username"`
	} `json:"username"`
}

func (res *meResponse) ToUser() domain.User {
	primaryEmail := ""

	for _, email := range res.Emails {
		primaryEmail = email.Address
		if email.IsPrimary {
			break
		}
	}
	return domain.User{
		ExternalID: res.UserID,
		Email:      primaryEmail,
	}
}
