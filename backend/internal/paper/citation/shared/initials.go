package shared

import (
	"strings"
	"unicode"
)

func FormatInitials(names ...string) string {
	var initials []string

	for _, name := range names {
		for part := range strings.FieldsSeq(name) {
			if initial := formatNameInitial(part); initial != "" {
				initials = append(initials, initial)
			}
		}
	}

	return strings.Join(initials, " ")
}

func formatNameInitial(name string) string {
	parts := strings.Split(name, "-")
	initials := make([]string, 0, len(parts))

	for _, part := range parts {
		if initial := firstLetterInitial(part); initial != "" {
			initials = append(initials, initial)
		}
	}

	return strings.Join(initials, "-")
}

func firstLetterInitial(name string) string {
	for _, letter := range name {
		if unicode.IsLetter(letter) {
			return strings.ToUpper(string(letter)) + "."
		}
	}

	return ""
}
