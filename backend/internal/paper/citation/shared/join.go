package shared

import (
	"strings"
)

func JoinNonEmpty(separator string, parts ...string) string {
	var nonEmptyParts []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}

	return strings.Join(nonEmptyParts, separator)
}
