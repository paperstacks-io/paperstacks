package paper_

import (
	"errors"

	"github.com/paperstacks.io/paperstacks/internal/domain_"
)

var (
	ErrPaperNotFound      = errors.New("not found")
	ErrPaperAlreadyExists = errors.New("paper_ already exists")
)

type Repository interface {
	Create(paper domain_.Paper) error
	ReadAll() ([]domain_.Paper, error)
	Read(id string) (domain_.Paper, error)
	Update(id string, paper domain_.Paper) error
	Delete(id string) error
}
