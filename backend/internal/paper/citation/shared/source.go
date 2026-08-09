package shared

import "github.com/paperstacks.io/paperstacks/internal/paper/domain"

type Source struct {
	Type        string
	PublishedIn string
	Publisher   string
	Volume      string
	Issue       string
	Pages       string
}

func NewSource(paper domain.Paper) Source {
	return Source{
		Type:        paper.Type,
		PublishedIn: paper.Metadata.PublishedIn,
		Publisher:   paper.Metadata.Publisher,
		Volume:      paper.Metadata.Volume,
		Issue:       paper.Metadata.Issue,
		Pages:       paper.Metadata.Pages,
	}
}

type SourceFormatter func(Source) string

type SourceRegistry struct {
	formatters map[string]SourceFormatter
}

func NewSourceRegistry(
	formatters map[string]SourceFormatter,
) SourceRegistry {
	return SourceRegistry{
		formatters: formatters,
	}
}

func (r SourceRegistry) Format(source Source) string {
	formatter, ok := r.formatters[source.Type]
	if !ok {
		return ""
	}

	return formatter(source)
}
