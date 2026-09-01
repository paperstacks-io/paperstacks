package application

import (
	"context"

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
}

// NewUserProvisioner creates a user onboarding service.
func NewUserProvisioner(
	userCreator UserCreator,
	defaultStackEnsurer DefaultStackEnsurer,
) *UserProvisioner {
	return &UserProvisioner{
		userCreator:         userCreator,
		defaultStackEnsurer: defaultStackEnsurer,
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
