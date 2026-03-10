package paper

import "github.com/paperstacks.io/paperstacks/internal/domain"

type Repository interface {
	Create(paper domain.Paper) error
	ReadAll() ([]domain.Paper, error)
	Read(id string) (domain.Paper, error)
	Update(id string, paper domain.Paper) error
	Delete(id string) error
}
