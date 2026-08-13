package domain

import (
	"cmp"
	"fmt"
	"time"
)

// Date represents a calendar date with optional month and day precision.
// A zero Date means that the publication date is unknown.
type Date struct {
	Year  int
	Month int
	Day   int
}

func (d Date) IsZero() bool {
	return d.Year == 0 && d.Month == 0 && d.Day == 0
}

func (d Date) IsValid() bool {
	if d.IsZero() {
		return true
	}

	if d.Year < 1 || d.Month < 0 || d.Month > 12 || d.Day < 0 {
		return false
	}

	if d.Month == 0 {
		return d.Day == 0
	}

	if d.Day == 0 {
		return true
	}

	return d.Day <= time.Date(d.Year, time.Month(d.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Compare returns -1, 0, or 1 according to chronological order.
func (d Date) Compare(other Date) int {
	if d.Year != other.Year {
		return cmp.Compare(d.Year, other.Year)
	}
	if d.Month != other.Month {
		return cmp.Compare(d.Month, other.Month)
	}

	return cmp.Compare(d.Day, other.Day)
}

func (d Date) String() string {
	switch {
	case d.IsZero():
		return ""
	case d.Month == 0:
		return fmt.Sprintf("%04d", d.Year)
	case d.Day == 0:
		return fmt.Sprintf("%04d-%02d", d.Year, d.Month)
	default:
		return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
	}
}
