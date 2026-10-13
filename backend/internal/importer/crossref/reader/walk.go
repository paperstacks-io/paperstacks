package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

func isDumpFile(name string) bool {
	return strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".jsonl.gz")
}
