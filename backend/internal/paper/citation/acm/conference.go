package acm

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
)

func formatConferenceSource(source shared.Source) string {
	publishedIn := strings.TrimSpace(source.PublishedIn)
	publisher := strings.TrimSpace(source.Publisher)
	pages := shared.NormalizePages(source.Pages)

	if publishedIn == "" {
		return ""
	}

	proceedings := "In " + publishedIn + "."

	publication := shared.JoinNonEmpty(
		", ",
		publisher,
		pages,
	)

	if publication == "" {
		return proceedings
	}

	return proceedings + " " + publication + "."
}
