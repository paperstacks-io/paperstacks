package ieee

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
)

func formatConferenceSource(
	source shared.Source,
) string {
	proceedings := strings.TrimSpace(source.PublishedIn)
	pages := formatPages(source.Pages)

	if proceedings == "" {
		return ""
	}

	return shared.JoinNonEmpty(
		", ",
		"in "+proceedings,
		pages,
	)
}
