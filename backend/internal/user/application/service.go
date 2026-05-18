package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type UserService struct {
	repo       domain.Repository
	authAPIURL string
	httpClient *http.Client
}

func NewUserService(repo domain.Repository, authAPIURL string, httpClient *http.Client) *UserService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &UserService{
		repo:       repo,
		authAPIURL: authAPIURL,
		httpClient: httpClient,
	}
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetByExternalID(ctx context.Context, externalID string) (domain.User, error) {
	return s.repo.GetByExternalID(ctx, strings.TrimSpace(externalID))
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

func (s *UserService) CreateFromToken(ctx context.Context, token string) (domain.User, error) {
	user, err := s.fetchUser(ctx, token)
	if err != nil {
		return domain.User{}, err
	}

	user = user.Normalize()
	if err := user.Validate(); err != nil {
		return domain.User{}, err
	}

	return s.repo.Save(ctx, user)
}

func (s *UserService) fetchUser(ctx context.Context, token string) (domain.User, error) {
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
		return domain.User{}, fmt.Errorf("validate session: unexpected status %d", res.StatusCode)
	}

	var response meResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return domain.User{}, fmt.Errorf("decode session response: %w", err)
	}

	user := response.ToUser()
	return user, nil
}

func (s *UserService) Update(ctx context.Context, externalID string, user domain.User) error {
	externalID = strings.TrimSpace(externalID)
	user = user.Normalize()

	if user.ExternalID != externalID {
		return domain.ErrExternalIDMismatch
	}

	if err := user.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, externalID, user)
}

func (s *UserService) Delete(ctx context.Context, externalID string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(externalID))
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
