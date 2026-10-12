package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/common/config"
)

func TestRunRejectsInvalidDatabaseURL(t *testing.T) {
	err := run(context.Background(), config.Config{DatabaseURL: "postgres://%invalid"})
	if err == nil || !strings.Contains(err.Error(), "database") || !strings.Contains(err.Error(), "cannot parse") {
		t.Fatalf("expected database configuration error, got %v", err)
	}
}

func TestRunPropagatesDatabaseConnectionCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := run(ctx, config.Config{DatabaseURL: "postgres://postgres@127.0.0.1:1/paperstacks?sslmode=disable"})
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "ping database") {
		t.Fatalf("expected cancelled connection check, got %v", err)
	}
}
