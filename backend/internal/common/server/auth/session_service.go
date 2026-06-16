package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
)

type HankoSessionService struct {
	apiURL     string
	httpClient *http.Client

	mu    sync.RWMutex
	cache map[string]*Session

	userService userApp.UserService
}

func NewHankoSessionService(apiURL string, userService userApp.UserService, httpClient *http.Client) *HankoSessionService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &HankoSessionService{
		apiURL:      apiURL,
		httpClient:  httpClient,
		cache:       make(map[string]*Session),
		userService: userService,
	}
}

func (s *HankoSessionService) ResolveSession(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}

	s.mu.RLock()
	cachedSession, ok := s.cache[token]
	s.mu.RUnlock()

	now := time.Now()
	if ok {
		if cachedSession.ExpirationTime.After(now) {
			return cachedSession, nil
		}

		s.mu.Lock()
		delete(s.cache, token)
		s.mu.Unlock()
	}

	session, err := s.fetchSession(ctx, token)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache[token] = session
	s.mu.Unlock()
	_, err = s.userService.CreateIfNotExist(ctx, session.UserID, session.Email)
	if err != nil {
		return session, err
	}

	return session, nil
}

func (s *HankoSessionService) fetchSession(ctx context.Context, token string) (*Session, error) {
	payload, err := json.Marshal(struct {
		SessionToken string `json:"session_token"`
	}{
		SessionToken: token,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal session request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+"/sessions/validate", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create session request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send session request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("validate session: unexpected status %d", res.StatusCode)
	}

	var response sessionResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode session response: %w", err)
	}

	session := response.Session(token)

	return session, nil
}

func (s *HankoSessionService) LogoutSession(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+"/logout", nil)
	if err != nil {
		return fmt.Errorf("create logout request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send logout request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout session: unexpected status %d", res.StatusCode)
	}

	s.mu.Lock()
	delete(s.cache, token)
	s.mu.Unlock()

	return nil
}
