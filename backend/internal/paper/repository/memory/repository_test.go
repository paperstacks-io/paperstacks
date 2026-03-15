package memory

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestRepositorySaveReturnsAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.Save(context.Background(), domain.Paper{
		DOI:   "10.1109/isese.2005.1541817",
		Title: "duplicate",
	})
	if err != domain.ErrPaperAlreadyExists {
		t.Fatalf("Save() error = %v, want %v", err, domain.ErrPaperAlreadyExists)
	}
}

func TestRepositoryUpdateIsAtomicUnderConcurrency(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()
	doi := "10.1109/isese.2005.1541817"

	const workers = 20

	var wg sync.WaitGroup
	start := make(chan struct{})
	successes := make(chan int, workers)

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		worker := i

		go func() {
			defer wg.Done()

			<-start
			log.Printf("Worker %d", worker)

			paper, err := repo.GetByDOI(ctx, doi)
			if err != nil {
				return
			}

			if paper.Title == "winner" {
				return
			}

			paper.Title = "winner"

			err = repo.Update(ctx, doi, paper)
			if err == nil {
				successes <- worker
			}
		}()
	}

	close(start)

	wg.Wait()
	close(successes)

	count := 0
	for range successes {
		count++
	}

	if count != 1 {
		t.Fatalf("expected exactly 1 successful update, got %d", count)
	}

	persisted, err := repo.GetByDOI(ctx, doi)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	if persisted.Title != "winner" {
		t.Fatalf("expected title %q, got %q", "winner", persisted.Title)
	}
}
