package domain

import (
	"context"
	"errors"
	"io"
)

var (
	ErrFileSizeExceeded = errors.New("file size exceeds maximum limit")
	ErrInvalidFileType  = errors.New("invalid file type: only valid PDFs allowed")
)

type Storage interface {
	Put(ctx context.Context, key string, r io.Reader) (string, error)
}
