package application

import (
	"context"
	"net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type UserCreator interface {
	CreateIfNotExist(ctx context.Context, externalID, email string) (domain.User, error)
}

type DefaultStackEnsurer interface {
	EnsureDefault(ctx context.Context, user domain.User) error
}

// UserProvisioner creates users and ensures their initial stack setup.
type UserProvisioner struct {
	userCreator         UserCreator
	defaultStackEnsurer DefaultStackEnsurer
	httpClient          http.Client
	authAPIURL          string
}

// NewUserProvisioner creates a user onboarding service.
func NewUserProvisioner(
	userCreator UserCreator,
	defaultStackEnsurer DefaultStackEnsurer,
	authAPIURL string,
	httpClient *http.Client,
) *UserProvisioner {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &UserProvisioner{
		userCreator:         userCreator,
		defaultStackEnsurer: defaultStackEnsurer,
		authAPIURL:          strings.TrimRight(strings.TrimSpace(authAPIURL), "/"),
		httpClient:          *httpClient,
	}
}

// Provision creates a user when needed and ensures its default stack exists.
func (p *UserProvisioner) Provision(ctx context.Context, externalID, email string) (domain.User, error) {
	user, err := p.userCreator.CreateIfNotExist(ctx, externalID, email)
	if err != nil {
		return domain.User{}, err
	}

	if err := p.defaultStackEnsurer.EnsureDefault(ctx, user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// ResolveByAuthToken validates an auth token and provisions its local user.
func (p *UserProvisioner) ResolveByAuthToken(ctx context.Context, token string) (domain.User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.User{}, domain.ErrInvalidAuthToken
	}

	user, err := p.fetchUserFromAuthProvider(ctx, token)
	if err != nil {
		return domain.User{}, err
	}

	return p.Provision(ctx, user.ExternalID, user.Email)
}
