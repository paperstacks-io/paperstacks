package citation

type CitationStyle struct {
	Name  string
	Style Style
	Path  string
}

const (
	APA  Style = "apa"
	IEEE Style = "ieee"
	ACM  Style = "acm"
)

type Style string

func CitationStyles(styles ...CitationStyle) []CitationStyle {
	return styles
}
