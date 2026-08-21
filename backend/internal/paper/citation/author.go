package citation

import "strings"

type Author struct {
	Given  string `json:"given,omitempty"`
	Family string `json:"family,omitempty"`
}

func given(first, middle string) string {
	return strings.TrimSpace(first + " " + middle)
}
