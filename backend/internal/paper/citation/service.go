package citation

import (
	"github.com/paperstacks.io/paperstacks/internal/paper/citation/acm"
	"github.com/paperstacks.io/paperstacks/internal/paper/citation/apa"
	"github.com/paperstacks.io/paperstacks/internal/paper/citation/ieee"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type CitationService struct {
	formatters map[CitationStyle]Formatter
}

func NewCitationService() CitationService {
	return CitationService{
		formatters: map[CitationStyle]Formatter{
			CitationStyleAPA:  apa.NewFormatter(),
			CitationStyleIEEE: ieee.NewFormatter(),
			CitationStyleACM:  acm.NewFormatter(),
		},
	}
}

func (s CitationService) Format(paper domain.Paper, style CitationStyle) (string, error) {
	formatter, ok := s.formatters[style]
	if !ok {
		return "", ErrUnsupportedCitationStyle
	}

	return formatter.Format(paper.Normalize()), nil
}
