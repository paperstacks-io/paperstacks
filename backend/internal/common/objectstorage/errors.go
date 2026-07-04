package objectstorage

import "errors"

var (
	ErrInvalidConfig  = errors.New("invalid object storage config")
	ErrObjectNotFound = errors.New("object not found")
)
