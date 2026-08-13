package domain

import "testing"

func TestDateIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		date Date
		want bool
	}{
		{name: "unknown", date: Date{}, want: true},
		{name: "year", date: Date{Year: 2026}, want: true},
		{name: "year and month", date: Date{Year: 2026, Month: 8}, want: true},
		{name: "leap day", date: Date{Year: 2024, Month: 2, Day: 29}, want: true},
		{name: "month without year", date: Date{Month: 8}, want: false},
		{name: "day without month", date: Date{Year: 2026, Day: 1}, want: false},
		{name: "invalid month", date: Date{Year: 2026, Month: 13}, want: false},
		{name: "invalid day", date: Date{Year: 2026, Month: 2, Day: 29}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.date.IsValid(); got != tt.want {
				t.Fatalf("Date.IsValid() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestDateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		date Date
		want string
	}{
		{date: Date{}, want: ""},
		{date: Date{Year: 2026}, want: "2026"},
		{date: Date{Year: 2026, Month: 8}, want: "2026-08"},
		{date: Date{Year: 2026, Month: 8, Day: 13}, want: "2026-08-13"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := tt.date.String(); got != tt.want {
				t.Fatalf("Date.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDateCompare(t *testing.T) {
	t.Parallel()

	if got := (Date{Year: 2026, Month: 8}).Compare(Date{Year: 2026, Month: 8, Day: 13}); got >= 0 {
		t.Fatalf("earlier date comparison = %d, want negative", got)
	}
	if got := (Date{Year: 2026, Month: 8, Day: 13}).Compare(Date{Year: 2026, Month: 8, Day: 13}); got != 0 {
		t.Fatalf("equal date comparison = %d, want 0", got)
	}
	if got := (Date{Year: 2026, Month: 9}).Compare(Date{Year: 2026, Month: 8, Day: 13}); got <= 0 {
		t.Fatalf("later date comparison = %d, want positive", got)
	}
}
