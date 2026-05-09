package application

import (
	"testing"
)

func TestServiceCreateNormalizesAndValidatesStack(t *testing.T) {
	t.Parallel()
	t.Skip("service not fully implemented")
}

func TestServiceDeleteReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()
	t.Skip("service not implemented yet")
}

func TestServiceCreateRejectsDuplicateStackNameForSameUser(t *testing.T) {
	t.Parallel()
	t.Skip("service not implemented yet")
}
