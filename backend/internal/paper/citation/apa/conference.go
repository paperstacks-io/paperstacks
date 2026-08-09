package apa

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
)

func formatConferenceSource(source shared.Source) string {
	proceedings := strings.TrimSpace(source.PublishedIn)
	publisher := strings.TrimSpace(source.Publisher)
	pages := shared.NormalizePages(source.Pages)

	if proceedings != "" {
		proceedings = "In " + proceedings

		if pages != "" {
			proceedings += " (pp. " + pages + ")"
		}
	}

	formatted := shared.JoinNonEmpty(
		". ",
		proceedings,
		publisher,
	)

	if formatted == "" {
		return ""
	}

	return formatted + "."
}
