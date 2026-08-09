package ieee

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
)

func formatJournalSource(
	source shared.Source,
) string {
	publishedIn := strings.TrimSpace(source.PublishedIn)
	volume := formatVolume(source.Volume)
	issue := formatIssue(source.Issue)
	pages := formatPages(source.Pages)

	return shared.JoinNonEmpty(
		", ",
		publishedIn,
		volume,
		issue,
		pages,
	)
}

func formatVolume(volume string) string {
	volume = strings.TrimSpace(volume)

	if volume == "" {
		return ""
	}

	return "vol. " + volume
}

func formatIssue(issue string) string {
	issue = strings.TrimSpace(issue)

	if issue == "" {
		return ""
	}

	return "no. " + issue
}

func formatPages(pages string) string {
	pages = shared.NormalizePages(pages)

	if pages == "" {
		return ""
	}

	if strings.Contains(pages, "–") {
		return "pp. " + pages
	}

	return "p. " + pages
}
