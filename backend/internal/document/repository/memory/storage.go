package memory

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/document/domain"
)

var _ domain.Storage = (*Storage)(nil)

type Storage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string][]byte),
	}
}

func (s *Storage) Put(ctx context.Context, key string, r io.Reader) error {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read file bytes: %w", err)
	}

	s.mu.Lock()
	s.data[key] = fileBytes
	s.mu.Unlock()

	return nil
}
