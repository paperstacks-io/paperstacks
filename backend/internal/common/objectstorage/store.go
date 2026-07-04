// Package objectstorage provides low-level object storage primitives backed by
// S3-compatible services such as RustFS.
package objectstorage

import (
	"context"
	"io"
)

type Store interface {
	Put(ctx context.Context, input PutObjectInput) (ObjectInfo, error)
	Get(ctx context.Context, key string) (*Object, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type PutObjectInput struct {
	Key         string
	Body        io.Reader
	Size        int64
	ContentType string
}

type Object struct {
	Body io.ReadCloser
	Info ObjectInfo
}

type ObjectInfo struct {
	Key         string
	ETag        string
	Size        int64
	ContentType string
}
