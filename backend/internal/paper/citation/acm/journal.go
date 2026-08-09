package acm

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
)

func formatJournalSource(source shared.Source) string {
	publishedIn := strings.TrimSpace(source.PublishedIn)
	volume := strings.TrimSpace(source.Volume)
	issue := strings.TrimSpace(source.Issue)
	pages := shared.NormalizePages(source.Pages)

	if publishedIn == "" {
		return ""
	}

	publication := shared.JoinNonEmpty(
		" ",
		publishedIn,
		volume,
	)

	return shared.JoinNonEmpty(
		", ",
		publication,
		issue,
		pages,
	) + "."
}
