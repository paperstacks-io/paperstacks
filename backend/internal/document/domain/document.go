package domain

import (
	"errors"
)

type Document struct {
	Key         string
	PaperUUID   string
	UserID      string
	FileName    string
	ContentType string
	Size        int64
}

func (d Document) Validate() error {
	if d.PaperUUID == "" {
		return errors.New("paper UUID is required")
	}

	if d.FileName == "" {
		return errors.New("file name is required")
	}

	return nil
}
