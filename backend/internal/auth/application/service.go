package application

import "github.com/paperstacks.io/paperstacks/internal/common/config"

type AuthService struct {
	HankoAPIURL string
}

func NewAuthService(cfg config.Config) *AuthService {
	return &AuthService{
		HankoAPIURL: cfg.HankoAPIURL,
	}
}

func (s *AuthService) GetUser(token string) {
}

func (s *AuthService) IsValid() {
}
