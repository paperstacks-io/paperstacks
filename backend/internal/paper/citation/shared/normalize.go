package shared

import "strings"

func NormalizePages(pages string) string {
	pages = strings.TrimSpace(pages)
	pages = strings.ReplaceAll(pages, "--", "–")

	return strings.ReplaceAll(pages, "-", "–")
}
