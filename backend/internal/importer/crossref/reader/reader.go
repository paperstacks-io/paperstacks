// Package reader streams Crossref work records out of the annual public
// data export: a directory of gzip-compressed JSONL files, one Crossref
// "work" per line. It owns gzip decompression and line scanning (see
// walk.go for directory traversal); turning a line into the domain
// model is delegated to the parser package.
package reader

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
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

// WalkFile opens a single .jsonl or .jsonl.gz file and calls visit for each
// parsed record.
//
// Errors for individual malformed lines are reported as Record.Err entries
// and do not abort the file; a fatal I/O error (e.g. a truncated gzip stream)
// is passed to visit as a final Record with a nil Paper.
func WalkFile(path string, visit func(Record) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var r io.Reader = f
	if isGzip(path) {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("open gzip stream %s: %w", path, err)
		}
		r = gz
	}

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
		record := Record{Paper: paper, LineNumber: lineNum}
		if err != nil {
			record.Err = fmt.Errorf("line %d: %w", lineNum, err)
		}
		if err := visit(record); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return visit(Record{Err: fmt.Errorf("scan %s: %w", path, err)})
	}
	return nil
}

func isGzip(path string) bool {
	return strings.HasSuffix(path, ".gz")
}
