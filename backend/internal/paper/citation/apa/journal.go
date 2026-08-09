package apa

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
)

func formatJournalSource(source shared.Source) string {
	publishedIn := strings.TrimSpace(source.PublishedIn)
	volume := strings.TrimSpace(source.Volume)
	issue := strings.TrimSpace(source.Issue)
	pages := shared.NormalizePages(source.Pages)

	var publicationInfo string

	switch {
	case volume != "" && issue != "":
		publicationInfo = volume + "(" + issue + ")"

	case volume != "":
		publicationInfo = volume

	case issue != "":
		publicationInfo = "(" + issue + ")"
	}

	formatted := shared.JoinNonEmpty(
		", ",
		publishedIn,
		publicationInfo,
		pages,
	)

	if formatted == "" {
		return ""
	}

	return formatted + "."
}
