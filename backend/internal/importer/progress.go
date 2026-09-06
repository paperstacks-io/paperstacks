package importer

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// progressRedrawInterval throttles how often the bar repaints so it
// doesn't dominate I/O during a fast import.
const progressRedrawInterval = 150 * time.Millisecond

// progressBar renders how many of total items have been processed so
// far. On a terminal it redraws a single line in place; when out isn't
// a terminal (e.g. redirected to a file or pipe) it prints plain
// periodic lines instead, so it never leaks raw escape codes into logs.
type progressBar struct {
	out        io.Writer
	total      int
	width      int
	tty        bool
	lastRender time.Time
}

func newProgressBar(out io.Writer, total int) *progressBar {
	b := &progressBar{out: out, total: total, width: 30}
	if f, ok := out.(*os.File); ok {
		b.tty = isTerminal(f)
	}
	return b
}

// Set redraws the bar for the given current count. Redraws are
// throttled; the final call (current >= total) always draws.
func (b *progressBar) Set(current int) {
	done := b.total > 0 && current >= b.total
	if !done && time.Since(b.lastRender) < progressRedrawInterval {
		return
	}
	b.lastRender = time.Now()

	line := b.render(current)
	if b.tty {
		fmt.Fprint(b.out, "\x1b[2K\r"+line)
		return
	}
	fmt.Fprintln(b.out, line)
}

func (b *progressBar) render(current int) string {
	if b.total <= 0 {
		return fmt.Sprintf("%d entries parsed", current)
	}

	frac := float64(current) / float64(b.total)
	if frac > 1 {
		frac = 1
	}
	filled := int(frac * float64(b.width))

	bar := strings.Repeat("#", filled) + strings.Repeat("-", b.width-filled)
	return fmt.Sprintf("[%s] %d/%d (%3.0f%%)", bar, current, b.total, frac*100)
}

// Done finalizes the bar, moving the cursor past it onto a new line.
func (b *progressBar) Done() {
	if b.tty {
		fmt.Fprintln(b.out)
	}
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
