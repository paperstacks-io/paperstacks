package citation

import "github.com/paperstacks.io/paperstacks/internal/paper/domain"

type Date struct {
	Parts [][]int `json:"date-parts"`
}

func date(date domain.Date) Date {
	switch {
	case date.Day != 0:
		return Date{Parts: [][]int{{date.Year, date.Month, date.Day}}}
	case date.Month != 0:
		return Date{Parts: [][]int{{date.Year, date.Month}}}
	case date.Year != 0:
		return Date{Parts: [][]int{{date.Year}}}
	default:
		return Date{}
	}
}
