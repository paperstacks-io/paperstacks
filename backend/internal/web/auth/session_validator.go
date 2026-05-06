package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SessionValidator defines the interface for session validation
type SessionValidator interface {
	ValidateSession(token string) (bool, error)
}

// HankoSessionValidator implements SessionValidator
type HankoSessionValidator struct {
	apiURL string
}

// ValidationResponse represents the Hanko API response
type ValidationResponse struct {
	IsValid bool `json:"is_valid"`
}

func NewHankoSessionValidator(apiURL string) *HankoSessionValidator {
	return &HankoSessionValidator{apiURL: apiURL}
}

func (v *HankoSessionValidator) ValidateSession(token string) (bool, error) {
	payload := strings.NewReader(fmt.Sprintf(`{"session_token":"%s"}`, token))

	req, err := http.NewRequest(http.MethodPost, v.apiURL+"/sessions/validate", payload)
	if err != nil {
		return false, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("Failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("Failed to read response: %w", err)
	}

	var validationRes ValidationResponse
	if err := json.Unmarshal(body, &validationRes); err != nil {
		return false, fmt.Errorf("Failed to parse response: %w", err)
	}

	return validationRes.IsValid, nil
}
