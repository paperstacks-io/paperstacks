package reader

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
)

// CountRecords scans every dump file in dir and returns the total number
// of non-empty JSONL lines across all of them, without parsing any of
// them. It's a cheap way to learn the size of an import — e.g. for a
// progress bar — before running the real, parsing pass.
//
// It checks ctx between files, so a cancelled context stops the scan
// promptly even on a large directory.
func CountRecords(ctx context.Context, dir string) (int, error) {
	total := 0
	err := WalkDumpDir(dir, func(path string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		n, err := countFileLines(path)
		if err != nil {
			return err
		}
		total += n
		return nil
	})
	return total, err
}

func countFileLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var r io.Reader = f
	if isGzip(path) {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return 0, fmt.Errorf("open gzip stream %s: %w", path, err)
		}
		defer gz.Close()
		r = gz
	}

	scanner := bufio.NewScanner(r)
	const maxTokenSize = 16 * 1024 * 1024
	scanner.Buffer(make([]byte, 64*1024), maxTokenSize)

	n := 0
	for scanner.Scan() {
		if len(scanner.Bytes()) > 0 {
			n++
		}
	}
	return n, scanner.Err()
}
