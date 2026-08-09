package citation

import "github.com/paperstacks.io/paperstacks/internal/paper/domain"

type Formatter interface {
	Format(domain.Paper) string
}
