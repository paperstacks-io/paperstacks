package paper

import (
	"errors"
	"slices"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/domain"
)

var (
	ErrPaperNotFound      = errors.New("not found")
	ErrPaperAlreadyExists = errors.New("paper already exists")
)

type MemoryRepo struct {
	data []domain.Paper
	// mu protects concurrent access to data. RWMutex has a usable zero value.
	mu sync.RWMutex
}

func NewMemoryRepo() Repository {
	data := []domain.Paper{
		{
			DOI:   "10.1109/isese.2005.1541817",
			Title: "Exploratory testing: a multiple case study",
			Authors: []domain.Author{
				{
					NameFirst: "J.",
					NameLast:  "Itkonen",
				},
				{
					NameFirst: "K.",
					NameLast:  "Rautiainen",
				},
			},
			Type: "proceedings-article",
			Metadata: domain.Metadata{
				Publisher:   "IEEE",
				PublishedIn: "2005 International Symposium on Empirical Software Engineering, 2005.",
				Pages:       "82-91",
			},
		},
		{
			DOI:   "10.1016/j.infsof.2023.107299",
			Title: "Code review guidelines for GUI-based testing artifacts",
			Authors: []domain.Author{
				{
					NameFirst:   "Andreas",
					NameLast:    "Bauer",
					Affiliation: "Blekinge Institute of Technology",
				},
				{
					NameFirst:   "Riccardo",
					NameLast:    "Coppola",
					Affiliation: "Politecnico di Torino",
				},
				{
					NameFirst:   "Emil",
					NameLast:    "Alégroth",
					Affiliation: "Blekinge Institute of Technology",
				},
				{
					NameFirst:   "Tony",
					NameLast:    "Gorschek",
					Affiliation: "Blekinge Institute of Technology",
				},
			},
			Type: "journal-article",
			Metadata: domain.Metadata{
				Publisher:   "Elsevier BV",
				PublishedIn: "Information and Software Technology",
				Pages:       "107299",
			},
		},
		{
			DOI:   "10.1109/icstw58534.2023.00015",
			Title: "We Tried and Failed: An Experience Report on a Collaborative Workflow for GUI-based Testing",
			Authors: []domain.Author{
				{
					NameFirst:   "Andreas",
					NameLast:    "Bauer",
					Affiliation: "Blekinge Institute of Technology",
				},
				{
					NameFirst:   "Emil",
					NameLast:    "Alégroth",
					Affiliation: "Blekinge Institute of Technology",
				},
			},
			Type: "proceedings-article",
			Metadata: domain.Metadata{
				Publisher:   "IEEE",
				PublishedIn: "2023 IEEE International Conference on Software Testing, Verification and Validation Workshops (ICSTW)",
				Pages:       "1-9",
			},
		},
		{
			DOI:   "10.1007/s10664-024-10522-z",
			Title: "Augmented testing to support manual GUI-based regression testing: An empirical study",
			Authors: []domain.Author{
				{
					NameFirst:   "Andreas",
					NameLast:    "Bauer",
					Affiliation: "Blekinge Institute of Technology",
				},
				{
					NameFirst:   "Julian",
					NameLast:    "Frattini",
					Affiliation: "Blekinge Institute of Technology",
				},
				{
					NameFirst:   "Emil",
					NameLast:    "Alégroth",
					Affiliation: "Blekinge Institute of Technology",
				},
			},
			Type: "journal-article",
		},
	}

	return &MemoryRepo{
		data: data,
	}
}

func (r *MemoryRepo) Create(paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if slices.ContainsFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == paper.DOI
	}) {
		return ErrPaperAlreadyExists
	}

	r.data = append(r.data, paper)
	return nil
}

func (r *MemoryRepo) ReadAll() ([]domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.data) == 0 {
		return nil, ErrPaperNotFound
	}

	return r.data, nil
}

func (r *MemoryRepo) Read(id string) (domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idx := slices.IndexFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == id
	})

	if idx == -1 {
		return domain.Paper{}, ErrPaperNotFound
	}

	return r.data[idx], nil
}

func (r *MemoryRepo) Update(id string, paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := slices.IndexFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == id
	})

	if idx == -1 {
		return ErrPaperNotFound
	}

	paper.DOI = id
	r.data[idx] = paper

	return nil
}

func (r *MemoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := slices.IndexFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == id
	})

	if idx == -1 {
		return ErrPaperNotFound
	}

	r.data = append(r.data[:idx], r.data[idx+1:]...)

	return nil
}
