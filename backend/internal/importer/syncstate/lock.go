package syncstate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// ErrLocked means another run of the same filter holds the lock.
var ErrLocked = errors.New("sync already running")

// Lock creates <dir>/<filter>.lock exclusively so that two runs of the
// same filter (e.g. overlapping cron jobs) cannot crawl concurrently.
// The returned func removes the lock. A lock left behind by a killed
// process must be deleted by hand.
func Lock(dir, filter string) (unlock func() error, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create state directory: %w", err)
	}
	path := filepath.Join(dir, filter+".lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("%w: lock file %s exists (delete it if no sync is running)", ErrLocked, path)
	}
	if err != nil {
		return nil, fmt.Errorf("create lock file: %w", err)
	}
	_, werr := f.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	if err := errors.Join(werr, f.Close()); err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("write lock file: %w", err)
	}
	return func() error { return os.Remove(path) }, nil
}
