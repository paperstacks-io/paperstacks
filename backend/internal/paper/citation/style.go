package citation

type CitationStyle struct {
	Name    string
	Style   Style
	CSLPath string
}

const (
	StyleAPA  Style = "apa"
	StyleIEEE Style = "ieee"
	StyleACM  Style = "acm"
)

type Style string

func CitationStyles(styles ...CitationStyle) []CitationStyle {
	return styles
}
