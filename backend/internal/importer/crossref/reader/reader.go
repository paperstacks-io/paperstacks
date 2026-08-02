// Package reader streams Crossref work records out of the annual public
// data export: a directory of gzip-compressed JSONL files, one Crossref
// "work" per line. It owns file-system traversal, gzip decompression and
// line scanning; turning a line into the domain model is delegated to
// the parser package.
package reader

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/domain"
	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/parser"
)

// Record wraps either a successfully parsed Paper or an error for one
// line of a JSONL file. LineNumber helps with locating malformed input.
type Record struct {
	Paper      *domain.Paper
	LineNumber int
	Err        error
}

// ReadFile opens a single .jsonl or .jsonl.gz file and sends each parsed
// record to the returned channel. The channel is closed once every line
// has been processed or a fatal I/O error occurs.
//
// Errors for individual malformed lines are reported as Record.Err
// entries and do not abort the file; a fatal I/O error (e.g. a truncated
// gzip stream) is sent as a final Record with a nil Paper, after which
// the channel is closed.
func ReadFile(path string) (<-chan Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	var r io.Reader = f
	if isGzip(path) {
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("open gzip stream %s: %w", path, err)
		}
		r = gz
	}

	out := make(chan Record, 256) // buffered to decouple I/O from downstream processing

	go func() {
		defer f.Close()
		defer close(out)

		scanner := bufio.NewScanner(r)
		// Crossref records — especially ones with long reference lists —
		// can exceed the scanner's 64 KB default token size.
		const maxTokenSize = 16 * 1024 * 1024
		scanner.Buffer(make([]byte, 64*1024), maxTokenSize)

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			paper, err := parser.Parse(line)
			if err != nil {
				out <- Record{LineNumber: lineNum, Err: fmt.Errorf("line %d: %w", lineNum, err)}
				continue
			}
			out <- Record{Paper: paper, LineNumber: lineNum}
		}

		if err := scanner.Err(); err != nil {
			out <- Record{Err: fmt.Errorf("scan %s: %w", path, err)}
		}
	}()

	return out, nil
}

// WalkDumpDir walks dir for Crossref dump files (named e.g. 0.jsonl.gz,
// 1.jsonl.gz, …) and calls fn once per file, in lexicographic file-name
// order. Only *.jsonl and *.jsonl.gz files are visited; fn is not called
// for anything else.
func WalkDumpDir(dir string, fn func(path string) error) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && isDumpFile(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if err := fn(filepath.Join(dir, name)); err != nil {
			return fmt.Errorf("processing %s: %w", name, err)
		}
	}
	return nil
}

func isGzip(path string) bool {
	return strings.HasSuffix(path, ".gz")
}

func isDumpFile(name string) bool {
	return strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".jsonl.gz")
}
