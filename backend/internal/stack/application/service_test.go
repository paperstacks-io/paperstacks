package application

import (
	"testing"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
)

var TestUser = userDomain.User{
	Email: "test@example.com",
}

// Create tests
func TestServiceCreateCreatesValidStack(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceCreateTrimsName(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceCreateRejectsEmptyName(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceCreateRejectsWhitespaceName(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceCreateRejectsDuplicateNameForSameUser(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceCreateSetsCreatedAt(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceCreateSetsUpdatedAt(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}

// Update tests
func TestServiceUpdateReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceUpdateChangesUpdatedAt(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
func TestServiceUpdateUpdatesStackName(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}

// Delete tests
func TestServiceDeleteReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}

// List tests
func TestServiceListReturnsStacks(t *testing.T) {
	t.Parallel()
	t.Skip("Not implemented")
}
